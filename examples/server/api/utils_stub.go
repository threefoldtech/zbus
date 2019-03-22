package api

import zbus "github.com/threefoldtech/zbus"

type UtilsStub struct {
	client zbus.Client
	module string
	object zbus.ObjectID
}

func NewUtilsStub(client zbus.Client) *UtilsStub {
	return &UtilsStub{
		client: client,
		module: "server",
		object: zbus.ObjectID{
			Name:    "utils",
			Version: "1.0",
		},
	}
}

func (s *UtilsStub) Capitalize(arg0 string) (ret0 string) {
	args := []interface{}{arg0}
	result, err := s.client.Request(s.module, s.object, "Capitalize", args...)
	if err != nil {
		panic(err)
	}
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	return
}
