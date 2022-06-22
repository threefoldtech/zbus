package stubs

import (
	"context"
	"time"

	zbus "github.com/threefoldtech/zbus"
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
	loader := zbus.Loader{
		&ret0,
	}
	if err := result.Unmarshal(&loader); err != nil {
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
	loader := zbus.Loader{
		&ret0,
	}
	if err := result.Unmarshal(&loader); err != nil {
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
	loader := zbus.Loader{}
	if err := result.Unmarshal(&loader); err != nil {
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
	ret3 = result.CallError()
	loader := zbus.Loader{
		&ret0,
		&ret1,
		&ret2,
	}
	if err := result.Unmarshal(&loader); err != nil {
		panic(err)
	}
	return
}
