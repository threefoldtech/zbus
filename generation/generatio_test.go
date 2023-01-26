package generation

func ExampleGenerate() {
	type Test interface {
		Hello(name string) string
		Add(a, b float64) string
		Divide(a, b float64) (float64, error)
	}

	var inf = (*Test)(nil)
	err := Generate(Options{Module: "example", Name: "test", Package: "stubs", Version: "1.0"}, inf)
	if err != nil {
		panic(err)
	}

	// Output: // GENERATED CODE
	// // --------------
	// // please do not edit manually instead use the "zbusc" to regenerate
	// package stubs

	// import (
	// 	"context"
	// 	zbus "github.com/threefoldtech/zbus"
	// )

	// type TestStub struct {
	// 	client zbus.Client
	// 	module string
	// 	object zbus.ObjectID
	// }

	// func NewTestStub(client zbus.Client) *TestStub {
	// 	return &TestStub{
	// 		client: client,
	// 		module: "example",
	// 		object: zbus.ObjectID{
	// 			Name:    "test",
	// 			Version: "1.0",
	// 		},
	// 	}
	// }

	// func (s *TestStub) Add(ctx context.Context, arg0 float64, arg1 float64) (ret0 string) {
	// 	args := []interface{}{arg0, arg1}
	// 	result, err := s.client.RequestContext(ctx, s.module, s.object, "Add", args...)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	result.PanicOnError()
	// 	loader := zbus.Loader{
	// 		&ret0,
	// 	}
	// 	if err := result.Unmarshal(&loader); err != nil {
	// 		panic(err)
	// 	}
	// 	return
	// }

	// func (s *TestStub) Divide(ctx context.Context, arg0 float64, arg1 float64) (ret0 float64, ret1 error) {
	// 	args := []interface{}{arg0, arg1}
	// 	result, err := s.client.RequestContext(ctx, s.module, s.object, "Divide", args...)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	result.PanicOnError()
	// 	ret1 = result.CallError()
	// 	loader := zbus.Loader{
	// 		&ret0,
	// 	}
	// 	if err := result.Unmarshal(&loader); err != nil {
	// 		panic(err)
	// 	}
	// 	return
	// }

	// func (s *TestStub) Hello(ctx context.Context, arg0 string) (ret0 string) {
	// 	args := []interface{}{arg0}
	// 	result, err := s.client.RequestContext(ctx, s.module, s.object, "Hello", args...)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	result.PanicOnError()
	// 	loader := zbus.Loader{
	// 		&ret0,
	// 	}
	// 	if err := result.Unmarshal(&loader); err != nil {
	// 		panic(err)
	// 	}
	// 	return
	// }
}
