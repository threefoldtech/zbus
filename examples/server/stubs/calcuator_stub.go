package stubs

import zbus "github.com/threefoldtech/zbus"

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

func (s *CalculatorStub) Add(arg0 ...float64) (ret0 float64) {
	args := []interface{}{}
	for _, argv := range arg0 {
		args = append(args, argv)
	}
	result, err := s.client.Request(s.module, s.object, "Add", args...)
	if err != nil {
		panic(err)
	}
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	return
}

func (s *CalculatorStub) Avg(arg0 []float64) (ret0 float64) {
	args := []interface{}{arg0}
	result, err := s.client.Request(s.module, s.object, "Avg", args...)
	if err != nil {
		panic(err)
	}
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	return
}

func (s *CalculatorStub) Divide(arg0 float64, arg1 float64) (ret0 float64, ret1 error) {
	args := []interface{}{arg0, arg1}
	result, err := s.client.Request(s.module, s.object, "Divide", args...)
	if err != nil {
		panic(err)
	}
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	ret1 = new(zbus.RemoteError)
	if err := result.Unmarshal(1, &ret1); err != nil {
		panic(err)
	}
	return
}

func (s *CalculatorStub) Pow(arg0 float64, arg1 float64) (ret0 float64) {
	args := []interface{}{arg0, arg1}
	result, err := s.client.Request(s.module, s.object, "Pow", args...)
	if err != nil {
		panic(err)
	}
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	return
}
