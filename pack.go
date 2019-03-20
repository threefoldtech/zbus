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
		bytes, err := msgpack.Marshal(arg)
		if err != nil {
			return msg, err
		}
		data = append(data, bytes)
	}

	return Message{ID: id, Arguments: data}, nil
}

// Argument loads an argument into an object of the given type `t`
func (m *Message) Argument(i int, t reflect.Type) (interface{}, error) {
	if i >= len(m.Arguments) {
		return nil, fmt.Errorf("index out of range")
	}

	value := reflect.New(t)
	element := value.Interface()
	if err := msgpack.Unmarshal(m.Arguments[i], element); err != nil {
		return nil, err
	}
	return value.Elem().Interface(), nil
}

// Request is carrier of byte data. It does not assume any encoding types used for individual objects
type Request struct {
	Message
	Object string
	Method string
}

// NewRequest creates a message that carries the given values
func NewRequest(id, object, method string, args ...interface{}) (*Request, error) {
	base, err := NewMessage(id, args...)

	if err != nil {
		return nil, err
	}

	return &Request{
		Message: base,
		Object:  object,
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
