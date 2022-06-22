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

// Message is base message object
type Message struct {
	ID    string
	Tuple Tuple
}

// NewMessage creates a new message
func NewMessage(id string, args ...interface{}) (msg Message, err error) {
	// We encode arguments separately before we encode the full msg object
	// To make sure we can decode each argument to its correct type at the
	// receiver end.

	tuple, err := newTuple(args...)
	if err != nil {
		return Message{}, err
	}

	return Message{ID: id, Tuple: tuple}, nil
}

// Unmarshal argument at position i into value
func (m *Message) Unmarshal(i int, v interface{}) error {
	return m.Tuple.Unmarshal(i, v)
}

// Value gets the concrete value stored at argument index i
func (m *Message) Value(i int, t reflect.Type) (interface{}, error) {
	arg, err := m.Argument(i, t)
	if err != nil {
		return nil, err
	}

	return arg.Interface(), nil
}

// Argument loads an argument into a reflect.Value of type t
func (m *Message) Argument(i int, t reflect.Type) (value reflect.Value, err error) {
	if i < 0 || i >= len(m.Tuple) {
		return value, fmt.Errorf("index out of range")
	}

	value = reflect.New(t)
	element := value.Interface()
	if err := msgpack.Unmarshal(m.Tuple[i], element); err != nil {
		return value, err
	}

	return value.Elem(), nil
}

// NumArguments returns the length of the argument list
func (m *Message) NumArguments() int {
	return len(m.Tuple)
}

// Request is carrier of byte data. It does not assume any encoding types used for individual objects
type Request struct {
	Message
	Object  ObjectID
	ReplyTo string
	Method  string
}

// NewRequest creates a message that carries the given values
func NewRequest(id, replyTo string, object ObjectID, method string, args ...interface{}) (*Request, error) {
	base, err := NewMessage(id, args...)

	if err != nil {
		return nil, err
	}

	return &Request{
		Message: base,
		Object:  object,
		ReplyTo: replyTo,
		Method:  method,
	}, nil
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

// Return results from a call
type Return struct {
	Tuple Tuple
	Error RemoteError
}

func returnFromValues(values []reflect.Value) (Return, error) {
	var objs []interface{}
	for _, res := range values {
		obj := res.Interface()
		objs = append(objs, obj)
	}

	return returnFromObjects(objs...)
}

func returnFromObjects(objs ...interface{}) (Return, error) {
	var ret Return
	if len(objs) == 0 {
		return ret, nil
	}

	trim := len(objs)
	last := objs[len(objs)-1]
	if err, ok := last.(error); ok {
		ret.Error = RemoteError{err.Error()}
		trim = len(objs) - 1
	}

	tuple, err := newTuple(objs[:trim]...)
	if err != nil {
		return ret, err
	}

	ret.Tuple = tuple
	return ret, nil
}

// Unmarshal argument at position i into value
func (t *Return) Unmarshal(i int, v interface{}) error {
	return t.Tuple.Unmarshal(i, v)
}

// Response object
type Response struct {
	// ID of response
	ID string
	// Return is returned data by call
	Return Return
	// Error here is any protocol error that is
	// not related to error returned by the remote call
	Error string
}

// NewResponse creates a response with id, and errMsg and return values
// note that errMsg is the protocol level errors (no such method, unknown object, etc...)
// errors returned by the service method itself should be encapsulated in the values
func NewResponse(id string, ret Return, errMsg string) *Response {
	return &Response{ID: id, Return: ret, Error: errMsg}
}

// Panic causes this response to panic
// in case of a protocol error. It's an
// indication to a problem with code hence
// a panic is okay
func (m *Response) PanicOnError() {
	if len(m.Error) != 0 {
		panic(m.Error)
	}
}

// Unmarshal argument at position i into value
func (m *Response) Unmarshal(i int, v interface{}) error {
	return m.Return.Unmarshal(i, v)
}

func (m *Response) CallError() error {
	if len(m.Return.Error.Message) != 0 {
		return &RemoteError{m.Return.Error.Message}
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
