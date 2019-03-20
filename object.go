package zbus

import (
	"fmt"
	"reflect"
)

//Version defines the object version
type Version string

//ObjectID defines an object id
type ObjectID struct {
	Name    string
	Version Version
}

func (o *ObjectID) String() string {
	if len(o.Version) == 0 {
		return o.Name
	}

	return fmt.Sprint(o.Name, "-", o.Version)
}

// Return results from a call
type Return []interface{}

// Surrogate a wrapper around an object to support dynamic method calls
type Surrogate struct {
	value reflect.Value
}

// NewSurrogate crates a new surrogate object
func NewSurrogate(object interface{}) Surrogate {
	return Surrogate{reflect.ValueOf(object)}
}

// Call dynamically call a method
func (s *Surrogate) Call(name string, args ...interface{}) (Return, error) {
	method := s.value.MethodByName(name)
	if method.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}
	//validating arguments
	methodType := method.Type()

	expected := methodType.NumIn()
	//	fmt.Println("expected", expected)
	if methodType.IsVariadic() {
		expected--
	}

	if len(args) < expected || !methodType.IsVariadic() && len(args) > expected {
		return nil, fmt.Errorf("invalid number of arguments expecting %d got %d", expected, len(args))
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
