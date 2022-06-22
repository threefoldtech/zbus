package main

import (
	"context"
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

	ctx := context.Background()

	fmt.Println(calculator.Add(ctx, 1, 2, 3, 4))
	fmt.Println(calculator.Divide(ctx, 200, 0))
	fmt.Println(utils.Capitalize(ctx, "this is awesome"))
	//fmt.Println(utils.Panic(ctx))
	fmt.Println(utils.Tuple(ctx))

	fmt.Println("after the panic")
	ch, err := utils.TikTok(ctx)
	if err != nil {
		panic(err)
	}
	for time := range ch {
		fmt.Println(time)
	}

}
