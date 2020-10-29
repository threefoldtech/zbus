package main

import (
	"fmt"
	"time"

	"github.com/threefoldtech/zbus"
	"github.com/threefoldtech/zbus/examples/server/stubs"
)

func main() {

	client, err := zbus.NewRedisClient("tcp://localhost:6379")
	if err != nil {
		panic(err)
	}

	//calculator := stubs.NewCalculatorStub(client)
	utils := stubs.NewUtilsStub(client)

	fmt.Println("calling sleep")
	utils.Sleep(10 * time.Second)
	fmt.Println("return from sleep")

	// fmt.Println(calculator.Add(1, 2, 3, 4))
	// fmt.Println(calculator.Divide(200, 0))
	// fmt.Println(utils.Capitalize("this is awesome"))
	// fmt.Println(utils.Panic())

	// fmt.Println("after the panic")
	// ctx := context.Background()
	// ch, err := utils.TikTok(ctx)
	// if err != nil {
	// 	panic(err)
	// }
	// for time := range ch {
	// 	fmt.Println(time)
	// }

}
