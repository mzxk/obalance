package obalance

import (
	"net/rpc"

	"github.com/gomodule/redigo/redis"
	"github.com/mzxk/omongo"
	"github.com/mzxk/oredis"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Trans struct {
	ID       primitive.ObjectID `bson:"_id"` //交易id
	Name     string             //交易名称
	Currency string             //交易币种
	Amount   float64            //交易数量
	Fee      float64            //交易手续费比例
	From     string             //从哪里来
	FromType string             //如locked
	To       string             //到哪里去
	ToType   string             //avail
	Detail   string             //详情，如deposit的id

	FromDB      int    //跨钱包转账用，默认6
	ToDB        int    //跨钱包转账用，默认6
	FromBalance string //当前账户转出成功后的余额
	ToBalance   string //目标账户收款成功后的余额

	Times int //重试次数，这个只在remote模式有用
	rds   *oredis.Oredis
	clt   *rpc.Client
}

const (
	Avail   = "avail"
	Locked  = "locked"
	OnOrder = "onOrder"
)

//New 创建转账的类型，币种，详情，创建交易，将会自动生成一个唯一id，并默认使用db6
func (t *Balance) New(name, currency, detail string) *Trans {
	tt := Trans{
		ID:       omongo.ID(""),
		Currency: currency,
		Name:     name,
		Detail:   detail,
		FromDB:   6,
		ToDB:     6,
		rds:      t.Rds,
		clt:      t.Clt,
	}
	return &tt
}
func (t *Trans) SetRedis(rds *oredis.Oredis) {
	t.rds = rds
}

//DBMove 设置转移的db名称，这个不是必须的，默认为db6,这代表一个跨余额转账，所以只能转移avail
func (t *Trans) DBMove(from, to int, id string, amount float64) *Trans {
	t.FromDB = from
	t.ToDB = to
	t.From = id
	t.To = id
	t.Amount = amount
	t.FromType = Avail
	t.ToType = Avail
	return t
}

//SetFee 设置手续费，这将减少入账的金额，并入到对应的手续费账号里
func (t *Trans) SetFee(fee float64) *Trans {
	t.Fee = fee
	return t
}

//Lock 锁定某用户余额，从avail到locked
func (t *Trans) Lock(id string, amount float64) *Trans {
	return t.setSingle(id, Avail, Locked, amount)
}

//UnLock 锁定某用户余额，从locked到avail
func (t *Trans) UnLock(id string, amount float64) *Trans {
	return t.setSingle(id, Locked, Avail, amount)
}

//IncrAvail 直接增加余额, 这将转出者设置为system
func (t *Trans) IncrAvail(id string, amount float64) *Trans {
	t.FromType = Avail
	t.ToType = Avail
	t.From = "system"
	t.To = id
	t.Amount = amount
	return t
}

//DecrAvail 这将直接减少可用余额，将会把余额转移到name类型的账户里, 相当于用户消费
func (t *Trans) DecrAvail(id string, amount float64) *Trans {
	t.FromType = Avail
	t.ToType = Avail
	t.From = id
	t.To = t.Name
	t.Amount = amount
	return t
}

//DecrLocked 这将直接减少锁定余额，通常用于提现锁定的直接减少
func (t *Trans) DecrLocked(id string, amount float64) *Trans {
	t.FromType = Locked
	t.ToType = Locked
	t.From = id
	t.To = t.Name
	t.Amount = amount
	return t
}

//setSingle 设置单一用户转账，比如自己解锁锁定之类
func (t *Trans) setSingle(id, fromType, toType string, amount float64) *Trans {
	t.From = id
	t.To = id
	t.FromType = fromType
	t.ToType = toType
	t.Amount = amount
	return t
}

//Run 启动转账，将自动区分本地模式和远程模式
func (t *Trans) Run() (*Trans, error) {
	//如果rpc客户端不为空，那么调用客户端
	if t.clt != nil {
		t.Times++
		reply := Trans{}
		err := t.clt.Call("Balance.Do", *t, &reply)
		t.FromBalance = reply.FromBalance
		t.ToBalance = reply.ToBalance
		return t, err
	}
	//调用本地模式
	toAmount := t.Amount
	var feeAmount float64
	//判断入账金额是否要扣除手续费
	if t.Fee > 0 {
		feeAmount = t.Amount * t.Fee
		toAmount = t.Amount - feeAmount
	}
	//虽然之前初始化代码写了6，但是这里还是判断一下比较好
	if t.FromDB == 0 {
		t.FromDB = 6
	}
	if t.ToDB == 0 {
		t.ToDB = 6
	}
	//调用redis
	result, err := redis.Strings(t.rds.Eval(Move, []interface{}{
		t.ID.Hex(),
		t.Name,
		t.Currency,
		t.From,
		join(t.Currency, t.FromType),
		t.To,
		join(t.Currency, t.ToType),
		t.Detail,
	},
		t.Amount,
		toAmount,
		feeAmount,
		t.Times,
		t.FromDB,
		t.ToDB,
	))
	if err != nil {
		return nil, err
	}
	//写入转账之后的余额表
	t.FromBalance = result[10]
	t.ToBalance = result[11]
	return t, nil
}
