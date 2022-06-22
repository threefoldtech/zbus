package zbus

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func (t *T) TupleError() (int, string, error) {
	return 10, "test", fmt.Errorf("some error")
}

func (t *T) Tuple() (int, string, string) {
	return 10, "hello", "world"
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

	if ok := assert.Len(t, result.Data, 1); !ok {
		t.Fatal()
	}

	require.Nil(t, result.Error)
	var v string
	err = result.Unmarshal(0, &v)
	require.NoError(t, err)
	require.Equal(t, "my-name", v)
}

func TestSurrogateArgs(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	result, err := s.Call("Add", 10, 20)
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, result.Data, 1); !ok {
		t.Fatal()
	}

	require.Nil(t, result.Error)
	var v int
	err = result.Unmarshal(0, &v)
	require.NoError(t, err)
	require.Equal(t, 30, v)
}

func TestSurrogateTupleReturn(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	result, err := s.Call("Tuple")

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, result.Data, 3); !ok {
		t.Fatal()
	}

	var v0 int
	err = result.Unmarshal(0, &v0)
	require.NoError(t, err)
	var v1 string
	err = result.Unmarshal(1, &v1)
	require.NoError(t, err)
	var v2 string
	err = result.Unmarshal(2, &v2)
	require.NoError(t, err)

	require.Equal(t, 10, v0)
	require.Equal(t, "hello", v1)
	require.Equal(t, "world", v2)

	require.Nil(t, result.Error)
}

func TestSurrogateTupleErrorReturn(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	result, err := s.Call("TupleError")

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, result.Data, 2); !ok {
		t.Fatal()
	}

	require.NotNil(t, result.Error)

	var v0 int
	err = result.Unmarshal(0, &v0)
	require.NoError(t, err)
	var v1 string
	err = result.Unmarshal(1, &v1)
	require.NoError(t, err)

	require.Equal(t, 10, v0)
	require.Equal(t, "test", v1)

	require.Equal(t, "some error", result.Error.Message)
}

func TestSurrogateVariadic(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	result, err := s.Call("Concat", "hello ", "world")

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, result.Data, 1); !ok {
		t.Fatal()
	}

	require.Nil(t, result.Error)

	var v string
	err = result.Unmarshal(0, &v)
	require.NoError(t, err)
	require.Equal(t, "hello world", v)
}

func TestSurrogateVariadicWithLeadingArgs(t *testing.T) {
	s := NewSurrogate(&T{"my-name"})

	result, err := s.Call("Join", "/", "hello", "world")

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, result.Data, 1); !ok {
		t.Fatal()
	}

	require.Nil(t, result.Error)

	var v string
	err = result.Unmarshal(0, &v)
	require.NoError(t, err)
	require.Equal(t, "hello/world", v)
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

	if ok := assert.Len(t, result.Data, 1); !ok {
		t.Fatal()
	}

	require.NotNil(t, result.Error)

	require.NotEmpty(t, result.Error.Message)
	require.Equal(t, "we made an error", result.Error.Message)

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

	if ok := assert.Len(t, result.Data, 1); !ok {
		t.Fatal()
	}

	require.Nil(t, result.Error)

	var v string
	err = result.Unmarshal(0, &v)
	require.NoError(t, err)
	require.Equal(t, "hello/world", v)

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

	if ok := assert.Len(t, result.Data, 1); !ok {
		t.Fatal()
	}

	require.Nil(t, result.Error)

	var v string
	err = result.Unmarshal(0, &v)
	require.NoError(t, err)
	require.Equal(t, "hello/world", v)
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
