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

func (s *CalculatorStub) Add(ctx context.Context, arg0 float64, arg1 float64) (ret0 float64) {
	args := []interface{}{arg0, arg1}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "Add", args...)
	if err != nil {
		panic(err)
	}
	result.PanicOnError()
	loader := zbus.Loader{
		&ret0,
	}
	if err := result.Unmarshal(&loader); err != nil {
		panic(err)
	}
	return
}

func (s *CalculatorStub) AddSub(ctx context.Context, arg0 float64, arg1 float64) (ret0 float64, ret1 float64) {
	args := []interface{}{arg0, arg1}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "AddSub", args...)
	if err != nil {
		panic(err)
	}
	result.PanicOnError()
	loader := zbus.Loader{
		&ret0,
		&ret1,
	}
	if err := result.Unmarshal(&loader); err != nil {
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
	loader := zbus.Loader{
		&ret0,
	}
	if err := result.Unmarshal(&loader); err != nil {
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
	ret1 = result.CallError()
	loader := zbus.Loader{
		&ret0,
	}
	if err := result.Unmarshal(&loader); err != nil {
		panic(err)
	}
	return
}

func (s *CalculatorStub) Pow(ctx context.Context, arg0 float64, arg1 float64) (ret0 float64) {
	args := []interface{}{arg0, arg1}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "Pow", args...)
	if err != nil {
		panic(err)
	}
	result.PanicOnError()
	loader := zbus.Loader{
		&ret0,
	}
	if err := result.Unmarshal(&loader); err != nil {
		panic(err)
	}
	return
}
