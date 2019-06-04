package stubs

import (
	"context"
	zbus "github.com/threefoldtech/zbus"
	"time"
)

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

func (s *UtilsStub) TikTok(ctx context.Context) (<-chan time.Time, error) {
	ch := make(chan time.Time)
	recv, err := s.client.Stream(ctx, s.module, s.object, "TikTok")
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(ch)
		for event := range recv {
			var obj time.Time
			if err := event.Unmarshal(&obj); err != nil {
				panic(err)
			}
			ch <- obj
		}
	}()
	return ch, nil
}
