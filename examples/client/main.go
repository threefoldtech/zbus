package main

import (
	"fmt"

	"github.com/threefoldtech/zbus"
	"github.com/threefoldtech/zbus/examples/server/stubs"
)

func main() {

	client, err := zbus.NewRedisClient("tcp://localhost:6379")
	if err != nil {
		panic(err)
	}

	calculator := stubs.NewCalculatorStub(client)
	utils := stubs.NewUtilsStub(client)

	fmt.Println(calculator.Add(1, 2, 3, 4))
	fmt.Println(calculator.Divide(200, 0))
	fmt.Println(utils.Capitalize("this is awesome"))
}
