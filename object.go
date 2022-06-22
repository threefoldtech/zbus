package zbus

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
)

//Version defines the object version
type Version string

//ObjectID defines an object id
type ObjectID struct {
	Name    string
	Version Version
}

func (o ObjectID) String() string {
	if len(o.Version) == 0 {
		return o.Name
	}

	return fmt.Sprint(o.Name, "@", o.Version)
}

//ObjectIDFromString parses an object id from string
func ObjectIDFromString(id string) ObjectID {
	parts := strings.SplitN(id, "@", 2)
	if len(parts) == 1 {
		return ObjectID{Name: parts[0]}
	}

	return ObjectID{Name: parts[0], Version: Version(parts[1])}
}

// Surrogate a wrapper around an object to support dynamic method calls
type Surrogate struct {
	value reflect.Value
}

// NewSurrogate crates a new surrogate object
func NewSurrogate(object interface{}) *Surrogate {
	return &Surrogate{reflect.ValueOf(object)}
}

func (s *Surrogate) getMethod(name string) (method reflect.Value, err error) {
	method = s.value.MethodByName(name)
	if method.Kind() != reflect.Func {
		return method, fmt.Errorf("not a function")
	}

	return method, nil
}

// isValid validate the method number of arguments
func (s *Surrogate) isValid(method reflect.Type, args int) error {
	expected := method.NumIn()
	if method.IsVariadic() {
		expected--
	}

	if args < expected || !method.IsVariadic() && args > expected {
		return fmt.Errorf("invalid number of arguments expecting %d got %d", expected, args)
	}
	return nil
}

// Call dynamically call a method
func (s *Surrogate) Call(name string, args ...interface{}) (ret Return, err error) {
	method, err := s.getMethod(name)
	if err != nil {
		return ret, err
	}

	methodType := method.Type()
	if err := s.isValid(methodType, len(args)); err != nil {
		return ret, err
	}

	expected := methodType.NumIn()

	if methodType.IsVariadic() {
		expected--
	}

	values := make([]reflect.Value, 0, len(args))
	for i, arg := range args[:expected] {
		got := reflect.TypeOf(arg)
		expect := methodType.In(i)

		if !got.AssignableTo(expect) {
			return ret, fmt.Errorf("invalid argument type [%d] expecting %s got %s", i, expect, got)
		}

		values = append(values, reflect.ValueOf(arg))
	}

	if methodType.IsVariadic() {
		expect := methodType.In(methodType.NumIn() - 1).Elem()
		for i, arg := range args[expected:] {
			got := reflect.TypeOf(arg)

			if !got.AssignableTo(expect) {
				return ret, fmt.Errorf("invalid argument type [%d] expecting %s got %s", i+expected, expect, got)
			}

			values = append(values, reflect.ValueOf(arg))
		}
	}

	results := method.Call(values)

	return returnFromValues(results)

}

// CallRequest calls a method defined by request
func (s *Surrogate) CallRequest(request *Request) (ret Return, err error) {
	method, err := s.getMethod(request.Method)
	if err != nil {
		return ret, err
	}

	methodType := method.Type()
	if err := s.isValid(methodType, request.NumArguments()); err != nil {
		return ret, err
	}

	expected := methodType.NumIn()

	if methodType.IsVariadic() {
		expected--
	}

	values := make([]reflect.Value, 0, len(request.Tuple))
	for i := 0; i < expected; i++ {
		expect := methodType.In(i)
		value, err := request.Argument(i, expect)
		if err != nil {
			return ret, fmt.Errorf("invalid argument type [%d] expecting %s got", i, expect)
		}

		values = append(values, value)
	}

	if methodType.IsVariadic() {
		expect := methodType.In(methodType.NumIn() - 1).Elem()
		for i := expected; i < request.NumArguments(); i++ {
			value, err := request.Argument(i, expect)
			if err != nil {
				return ret, fmt.Errorf("invalid argument type [%d] expecting %s", i+expected, expect)
			}

			values = append(values, value)
		}

	}

	results := method.Call(values)

	return returnFromValues(results)

}

// Streams return all stream objects associated with this object
// stream methods only take one method (context) and must return
// a single value a chan of a static type (struct, or primitive)
func (s *Surrogate) Streams() []Stream {
	num := s.value.NumMethod()
	var streams []Stream
	for i := 0; i < num; i++ {
		method := s.value.Method(i)
		name := s.value.Type().Method(i).Name

		methodType := method.Type()

		// stream methods are always of type `fn(Context) -> chan T`
		if methodType.NumIn() != 1 || methodType.NumOut() != 1 {
			continue
		}

		// not a valid input type, expecting a context.Context
		if !methodType.In(0).Implements(contextType) {
			continue
		}

		if methodType.Out(0).Kind() != reflect.Chan {
			continue
		}

		//todo: validate the Elem of the chan is valid
		streams = append(streams, Stream{name: name, method: method})
	}

	return streams
}

// Stream represents a channel of events
type Stream struct {
	name   string
	method reflect.Value
}

func (s *Stream) Name() string {
	return s.name
}

// Run stream to completion
func (s *Stream) Run(ctx context.Context) <-chan interface{} {
	out := make(chan interface{})

	in := []reflect.Value{
		reflect.ValueOf(ctx),
	}

	//we already done validation of the number of inputs
	//and number of outputs in the "Streams" method, so
	//it's safe to access the values with index directly
	values := s.method.Call(in)
	ch := values[0]
	go func() {
		defer close(out)

		for {
			//if we used rect it will not be possible to select
			//on the context
			value, ok := ch.Recv()
			if !ok {
				return
			}

			select {
			case out <- value.Interface():
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}
