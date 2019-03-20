package zbus

import (
	"fmt"
	"reflect"

	"github.com/vmihailenco/msgpack"
)

// Message is base message object
type Message struct {
	ID        string
	Arguments [][]byte
}

// NewMessage creates a new message
func NewMessage(id string, args ...interface{}) (msg Message, err error) {
	// We encode arguments separately before we encode the full msg object
	// To make sure we can decode each argument to its correct type at the
	// receiver end.
	data := make([][]byte, 0, len(args))
	for _, arg := range args {
		if o, ok := arg.(error); ok {
			arg = RemoteError{o.Error()}
		}

		bytes, err := msgpack.Marshal(arg)
		if err != nil {
			return msg, err
		}
		data = append(data, bytes)
	}

	return Message{ID: id, Arguments: data}, nil
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
	if i >= len(m.Arguments) {
		return value, fmt.Errorf("index out of range")
	}

	value = reflect.New(t)
	element := value.Interface()
	if err := msgpack.Unmarshal(m.Arguments[i], element); err != nil {
		return value, err
	}

	return value.Elem(), nil
}

// NumArguments returns the length of the argument list
func (m *Message) NumArguments() int {
	return len(m.Arguments)
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

// Response object
type Response struct {
	Message
	// Error hear will carry any protocol error
	Error string
}

// NewResponse creates a response with id, and errMsg and return values
// note that errMsg is the protocol level errors (no such method, unknown object, etc...)
// errors returned by the service method itself should be encapsulated in the values
func NewResponse(id, errMsg string, values ...interface{}) (*Response, error) {
	msg, err := NewMessage(id, values...)
	if err != nil {
		return nil, err
	}

	return &Response{msg, errMsg}, nil
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
