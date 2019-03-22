package zbus

import (
	"fmt"
	"reflect"
	"strings"
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

// Return results from a call
type Return []interface{}

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
func (s *Surrogate) Call(name string, args ...interface{}) (Return, error) {
	method, err := s.getMethod(name)
	if err != nil {
		return nil, err
	}

	methodType := method.Type()
	if err := s.isValid(methodType, len(args)); err != nil {
		return nil, err
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
			return nil, fmt.Errorf("invalid argument type [%d] expecting %s got %s", i, expect, got)
		}

		values = append(values, reflect.ValueOf(arg))
	}

	if methodType.IsVariadic() {
		expect := methodType.In(methodType.NumIn() - 1).Elem()
		for i, arg := range args[expected:] {
			got := reflect.TypeOf(arg)

			if !got.AssignableTo(expect) {
				return nil, fmt.Errorf("invalid argument type [%d] expecting %s got %s", i+expected, expect, got)
			}

			values = append(values, reflect.ValueOf(arg))
		}
	}

	results := method.Call(values)
	ret := make(Return, 0, len(results))

	for _, res := range results {
		ret = append(ret, res.Interface())
	}

	return ret, nil
}

// CallRequest calls a method defined by request
func (s *Surrogate) CallRequest(request *Request) (Return, error) {
	method, err := s.getMethod(request.Method)
	if err != nil {
		return nil, err
	}

	methodType := method.Type()
	if err := s.isValid(methodType, request.NumArguments()); err != nil {
		return nil, err
	}

	expected := methodType.NumIn()

	if methodType.IsVariadic() {
		expected--
	}

	values := make([]reflect.Value, 0, len(request.Arguments))
	for i := 0; i < expected; i++ {
		expect := methodType.In(i)
		value, err := request.Argument(i, expect)
		if err != nil {
			return nil, fmt.Errorf("invalid argument type [%d] expecting %s got", i, expect)
		}

		values = append(values, value)
	}

	if methodType.IsVariadic() {
		expect := methodType.In(methodType.NumIn() - 1).Elem()
		for i := expected; i < request.NumArguments(); i++ {
			value, err := request.Argument(i, expect)
			if err != nil {
				return nil, fmt.Errorf("invalid argument type [%d] expecting %s", i+expected, expect)
			}

			values = append(values, value)
		}

	}

	results := method.Call(values)
	ret := make(Return, 0, len(results))

	for _, res := range results {
		ret = append(ret, res.Interface())
	}

	return ret, nil
}
