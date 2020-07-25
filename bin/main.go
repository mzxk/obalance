package main

import (
	"net"
	"net/rpc"
	"os"

	"github.com/mzxk/obalance"
)

var balance *obalance.Balance

func main() {
	err := rpc.Register(new(Balance))
	if err != nil {
		panic(err)
	}
	balance = obalance.NewLocal("127.0.0.1:6379", "")
	obalance.InitScript(balance.Rds)
	listen, err := net.Listen("tcp", "0.0.0.0"+os.Args[1])
	if err != nil {
		panic(err)
	}
	rpc.Accept(listen)
}
func (t *Balance) Do(t1 obalance.Trans, reply *obalance.Trans) error {
	t1.SetRedis(balance.Rds)
	_, err := t1.Run()
	if err != nil {
		return err
	}
	reply.FromBalance = t1.FromBalance
	reply.ToBalance = t1.ToBalance
	return nil
}
func (t *Balance) Get(id string, reply *map[string]obalance.Amount) error {
	result, err := balance.GetBalance(id)
	reply2 := *reply
	for k, v := range result {
		reply2[k] = v
	}
	return err
}

type Balance struct {
}
