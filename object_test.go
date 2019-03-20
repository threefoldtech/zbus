package zbus

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type T struct {
	Name string
}

func (t *T) GetName() string {
	return t.Name
}

func (t *T) Add(a, b int) int {
	return a + b
}

func (t *T) Concat(a ...string) string {
	return strings.Join(a, "")
}

func (t *T) Join(sep string, a ...string) string {
	return strings.Join(a, sep)
}

func (t *T) MakeError() (int, error) {
	return 0, fmt.Errorf("we made an error")
}

func TestSurrogate(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	_, err := s.Call("NotDefined")
	if ok := assert.EqualError(t, err, "not a function"); !ok {
		t.Error()
	}

	_, err = s.Call("Name")
	if ok := assert.EqualError(t, err, "not a function"); !ok {
		t.Error()
	}

	result, err := s.Call("GetName")
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, result, 1); !ok {
		t.Fatal()
	}

	if ok := assert.IsType(t, "", result[0]); !ok {
		t.Error()
	}
}

func TestSurrogateArgs(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	result, err := s.Call("Add", 10, 20)
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, result, 1); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, 30, result[0]); !ok {
		t.Fatal()
	}
}

func TestSurrogateVariadic(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	result, err := s.Call("Concat", "hello", "world")

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, result, 1); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "helloworld", result[0]); !ok {
		t.Error()
	}
}

func TestSurrogateVariadicWithLeadingArgs(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	result, err := s.Call("Join", "/", "hello", "world")

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, result, 1); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "hello/world", result[0]); !ok {
		t.Error()
	}
}

func TestSurrogateVariadicWithWrongTypes(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	_, err := s.Call("Join", "/", "hello", 10)

	if ok := assert.EqualError(t, err, "invalid argument type [2] expecting string got int"); !ok {
		t.Fatal()
	}

}
