package stubs

import (
	"context"
	zbus "github.com/threefoldtech/zbus"
)

type CalculatorStub struct {
	client zbus.Client
	module string
	object zbus.ObjectID
}

func NewCalculatorStub(client zbus.Client) *CalculatorStub {
	return &CalculatorStub{
		client: client,
		module: "server",
		object: zbus.ObjectID{
			Name:    "calculator",
			Version: "1.0",
		},
	}
}

func (s *CalculatorStub) Add(ctx context.Context, arg0 ...float64) (ret0 float64) {
	args := []interface{}{}
	for _, argv := range arg0 {
		args = append(args, argv)
	}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "Add", args...)
	if err != nil {
		panic(err)
	}
	result.PanicOnError()
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	return
}

func (s *CalculatorStub) Avg(ctx context.Context, arg0 []float64) (ret0 float64) {
	args := []interface{}{arg0}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "Avg", args...)
	if err != nil {
		panic(err)
	}
	result.PanicOnError()
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	return
}

func (s *CalculatorStub) Divide(ctx context.Context, arg0 float64, arg1 float64) (ret0 float64, ret1 error) {
	args := []interface{}{arg0, arg1}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "Divide", args...)
	if err != nil {
		panic(err)
	}
	result.PanicOnError()
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	ret1 = result.CallError()
	return
}

func (s *CalculatorStub) Pow(ctx context.Context, arg0 float64, arg1 float64) (ret0 float64) {
	args := []interface{}{arg0, arg1}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "Pow", args...)
	if err != nil {
		panic(err)
	}
	result.PanicOnError()
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	return
}
