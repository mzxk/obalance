package obalance

import (
	"errors"
	"net/rpc"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/mzxk/oredis"
)

type Balance struct {
	Rds *oredis.Oredis
	Clt *rpc.Client
}
type UserBalance struct {
	B map[string]Amount
}
type Amount struct {
	Avail   string
	Locked  string
	OnOrder string
}

func (t *Balance) GetBalance(id string) (map[string]Amount, error) {
	result := map[string]Amount{}
	//如果是远程，就直接调用
	if t.Clt != nil {
		err := t.Clt.Call("Balance.Get", id, &result)
		return result, err
	}
	//打开默认的db，暂时不支持其他db的getbalance
	c, err := t.Rds.GetDB(6)
	if err != nil {
		return nil, err
	}
	rlt, err := redis.StringMap(c.Do("hgetall", id))
	if err != nil {
		return nil, err
	}
	//改成我们需要的格式
	for k := range rlt {
		ss := strings.Split(k, ".")
		if len(ss) != 2 {
			return nil, errors.New("wrongUserBalance" + id)
		}
		coin := ss[0]
		if _, ok := result[coin]; !ok {
			result[coin] = Amount{
				Avail:   s2s(rlt[join(coin, Avail)]),
				Locked:  s2s(rlt[join(coin, Locked)]),
				OnOrder: s2s(rlt[join(coin, OnOrder)]),
			}
		}
	}
	return result, nil
}

//NewLocal 使用这个初始化，将使用本地redis模式
func NewLocal(add, pwd string) *Balance {
	rds := oredis.New(add, pwd)
	InitScript(rds)
	return &Balance{
		Rds: rds,
	}
}

//NewRemote 使用这个初始化，将使用rpc远程模式
func NewRemote(url string) *Balance {
	t := &Balance{}
	client, err := rpc.Dial("tcp", url)
	if err != nil {
		panic(err)
	}
	t.Clt = client
	return t
}

var (
	ErrNotEnoughBalance = errors.New("notEnoughBalance")
	ErrIdExisted        = errors.New("idExisted")
)
