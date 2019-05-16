package zbus

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

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

func (t *T) TikTok(ctx context.Context) <-chan int {
	c := make(chan int)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer close(c)
		defer ticker.Stop()

		i := 0
		for range ticker.C {
			i++
			select {
			case <-ctx.Done():
				return
			case c <- i:
			}
		}
	}()

	return c
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

func TestSurrogateError(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	_, err := s.Call("MakeError", 10)
	if ok := assert.EqualError(t, err, "invalid number of arguments expecting 0 got 1"); !ok {
		t.Fatal()
	}

	result, err := s.Call("MakeError")
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, result, 2); !ok {
		t.Fatal()
	}

	if ok := assert.IsType(t, errors.New(""), result[1]); !ok {
		t.Error()
	}

}

func TestSurrogateRequest(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	request, err := NewRequest("id", "", ObjectID{}, "Join", "/", "hello", "world")
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	result, err := s.CallRequest(request)
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

func TestSurrogateRequestWithWrongTypes(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	request, err := NewRequest("id", "", ObjectID{}, "Join", "/", "hello", 10)
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	_, err = s.CallRequest(request)
	if ok := assert.EqualError(t, err, "invalid argument type [3] expecting string"); !ok {
		t.Fatal()
	}
}

func TestSurrogateRequestEncoded(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	request, err := NewRequest("id", "", ObjectID{}, "Join", "/", "hello", "world")
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	data, err := request.Encode()
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	request, err = LoadRequest(data)
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	result, err := s.CallRequest(request)
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

func TestSurrogateStreamsList(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})
	streams := s.Streams()

	if ok := assert.Len(t, streams, 1); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "TikTok", streams[0].Name()); !ok {
		t.Fatal()
	}
}

func TestStreamRun(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})
	streams := s.Streams()

	if ok := assert.Len(t, streams, 1); !ok {
		t.Fatal()
	}

	stream := streams[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	i := 0
	for range stream.Run(ctx) {
		i++
		if i == 3 {
			cancel()
		}
	}

	if ok := assert.True(t, i >= 3); !ok {
		t.Fatal()
	}
}
