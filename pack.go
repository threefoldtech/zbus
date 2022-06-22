package zbus

import (
	"fmt"
	"reflect"

	"github.com/vmihailenco/msgpack"
)

type Tuple [][]byte

func newTuple(args ...interface{}) (Tuple, error) {
	data := make([][]byte, 0, len(args))
	for _, arg := range args {

		bytes, err := msgpack.Marshal(arg)
		if err != nil {
			return nil, err
		}
		data = append(data, bytes)
	}

	return data, nil
}

// Unmarshal argument at position i into value
func (t Tuple) Unmarshal(i int, v interface{}) error {
	if i < 0 || i >= len(t) {
		return fmt.Errorf("index out of range")
	}

	return msgpack.Unmarshal(t[i], v)
}

// Request is carrier of byte data. It does not assume any encoding types used for individual objects
type Request struct {
	ID      string
	Inputs  Tuple
	Object  ObjectID
	ReplyTo string
	Method  string
}

// NewRequest creates a message that carries the given values
func NewRequest(id, replyTo string, object ObjectID, method string, args ...interface{}) (*Request, error) {
	inputs, err := newTuple(args...)
	if err != nil {
		return nil, err
	}

	return &Request{
		ID:      id,
		Inputs:  inputs,
		Object:  object,
		ReplyTo: replyTo,
		Method:  method,
	}, nil
}

// Unmarshal argument at position i into value
func (m *Request) Unmarshal(i int, v interface{}) error {
	return m.Inputs.Unmarshal(i, v)
}

// Value gets the concrete value stored at argument index i
func (m *Request) Value(i int, t reflect.Type) (interface{}, error) {
	arg, err := m.Argument(i, t)
	if err != nil {
		return nil, err
	}

	return arg.Interface(), nil
}

// Argument loads an argument into a reflect.Value of type t
func (m *Request) Argument(i int, t reflect.Type) (value reflect.Value, err error) {
	if i < 0 || i >= len(m.Inputs) {
		return value, fmt.Errorf("index out of range")
	}

	value = reflect.New(t)
	element := value.Interface()
	if err := msgpack.Unmarshal(m.Inputs[i], element); err != nil {
		return value, err
	}

	return value.Elem(), nil
}

// NumArguments returns the length of the argument list
func (m *Request) NumArguments() int {
	return len(m.Inputs)
}

// LoadRequest from bytes
func LoadRequest(data []byte) (*Request, error) {
	var request Request
	return &request, msgpack.Unmarshal(data, &request)
}

// Encode converts a message into byte data suitable to send over the wire
// Encode will always use msgpack.
func (m *Request) Encode() ([]byte, error) {
	return msgpack.Marshal(m)
}

// CallError is a concrete type used to wrap all errors returned by services
// for example, if a method `f` returns `error` the return.Error() is stored in a CallError struct
type CallError struct {
	Message string
}

func (r *CallError) Error() string {
	return r.Message
}

// Output results from a call
type Output struct {
	Data  []byte
	Error *CallError
}

func returnFromValues(values []reflect.Value) (Output, error) {
	var objs []interface{}
	var err error
	for _, res := range values {
		typ := res.Type()
		if typ.Kind() == reflect.Interface && typ.Name() == "error" {
			if v := res.Interface(); v != nil {
				err = v.(error)
			}
			continue
		}
		obj := res.Interface()
		objs = append(objs, obj)
	}

	return returnFromObjects(err, objs...)
}

func returnFromObjects(err error, objs ...interface{}) (Output, error) {
	var ret Output
	if err != nil {
		ret.Error = &CallError{err.Error()}
	}

	if len(objs) == 0 {
		return ret, nil
	}

	var data []byte
	var encErr error
	if len(objs) == 1 {
		data, encErr = msgpack.Marshal(objs[0])
	} else if len(objs) > 1 {
		data, encErr = msgpack.Marshal(objs)
	}

	if encErr != nil {
		return Output{}, err
	}

	ret.Data = data
	return ret, nil
}

// Unmarshal argument at position i into value
func (t *Output) Unmarshal(v *Loader) error {
	if len(*v) == 0 {
		return nil
	} else if len(*v) == 1 {
		return msgpack.Unmarshal(t.Data, (*v)[0])
	}
	// in case is more we assume we have a longer list of objects
	return msgpack.Unmarshal(t.Data, v)
}

// Response object
type Response struct {
	// ID of response
	ID string
	// Output is returned data by call
	Output Output
	// Error here is any protocol error that is
	// not related to error returned by the remote call
	Error *string
}

// NewResponse creates a response with id, and errMsg and return values
// note that errMsg is the protocol level errors (no such method, unknown object, etc...)
// errors returned by the service method itself should be encapsulated in the values
func NewResponse(id string, ret Output, errMsg string) *Response {
	var err *string
	if len(errMsg) != 0 {
		err = &errMsg
	}

	return &Response{ID: id, Output: ret, Error: err}
}

// Panic causes this response to panic
// in case of a protocol error. It's an
// indication to a problem with code hence
// a panic is okay
func (m *Response) PanicOnError() {
	if m.Error != nil {
		panic(*m.Error)
	}
}

type Loader []interface{}

// Unmarshal argument at position i into value
func (m *Response) Unmarshal(v *Loader) error {
	return m.Output.Unmarshal(v)
}

func (m *Response) CallError() error {
	if m.Output.Error == nil {
		return nil
	}

	if len(m.Output.Error.Message) != 0 {
		return &CallError{m.Output.Error.Message}
	}

	return nil
}

// Encode converts a response into byte data suitable to send over the wire
// Encode will always use msgpack.
func (m *Response) Encode() ([]byte, error) {
	return msgpack.Marshal(m)
}

// LoadResponse loads response from data
func LoadResponse(data []byte) (*Response, error) {
	var response Response
	return &response, msgpack.Unmarshal(data, &response)
}

type Event []byte

func (e Event) Unmarshal(o interface{}) error {
	return msgpack.Unmarshal(e, o)
}
