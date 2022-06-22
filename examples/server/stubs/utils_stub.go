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

func (s *UtilsStub) Capitalize(ctx context.Context, arg0 string) (ret0 string) {
	args := []interface{}{arg0}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "Capitalize", args...)
	if err != nil {
		panic(err)
	}
	result.PanicOnError()
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	return
}

func (s *UtilsStub) Panic(ctx context.Context) (ret0 int) {
	args := []interface{}{}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "Panic", args...)
	if err != nil {
		panic(err)
	}
	result.PanicOnError()
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	return
}

func (s *UtilsStub) Sleep(ctx context.Context, arg0 time.Duration) (ret0 error) {
	args := []interface{}{arg0}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "Sleep", args...)
	if err != nil {
		panic(err)
	}
	result.PanicOnError()
	ret0 = result.CallError()
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
			select {
			case <-ctx.Done():
				return
			case ch <- obj:
			default:
			}
		}
	}()
	return ch, nil
}

func (s *UtilsStub) Tuple(ctx context.Context) (ret0 int, ret1 string, ret2 float64, ret3 error) {
	args := []interface{}{}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "Tuple", args...)
	if err != nil {
		panic(err)
	}
	result.PanicOnError()
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	if err := result.Unmarshal(1, &ret1); err != nil {
		panic(err)
	}
	if err := result.Unmarshal(2, &ret2); err != nil {
		panic(err)
	}
	ret3 = result.CallError()
	return
}
