package main

import (
	"fmt"

	"github.com/mzxk/obalance"
)

func main() {
	b := obalance.NewRemote("127.0.0.1:6666")
	result, err := b.GetBalance("system")
	fmt.Println(result, err)
}
