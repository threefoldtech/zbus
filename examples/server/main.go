package main

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/threefoldtech/zbus"
	"github.com/threefoldtech/zbus/examples/server/api"
)

type Calculator struct{}
type Utils struct{}

var (
	_ api.Calculator = (*Calculator)(nil)
	_ api.Utils      = (*Utils)(nil)
)

func (c *Calculator) Add(a ...float64) float64 {
	var r float64
	for _, v := range a {
		r += v
	}

	return r
}

func (c *Calculator) Pow(a, b float64) float64 {
	return math.Pow(a, b)
}

func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}

	return a / b, nil
}

func (c *Calculator) Avg(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	var t float64
	for _, x := range v {
		t += x
	}
	return t / float64(len(v))
}

func (u *Utils) Capitalize(s string) string {
	return strings.ToUpper(s)
}

func main() {
	server, err := zbus.NewRedisServer("server", "tcp://localhost:6379", 1)

	if err != nil {
		panic(err)
	}
	var calc Calculator
	var utils Utils

	server.Register(zbus.ObjectID{Name: "calculator", Version: "1.0"}, &calc)
	server.Register(zbus.ObjectID{Name: "utils", Version: "1.0"}, &utils)

	if err := server.Run(context.Background()); err != nil {
		panic(err)
	}
}
