package zbus

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBaseServer(t *testing.T) {
	s := BaseServer{}
	var o T

	id := ObjectID{Name: "calc"}
	s.Register(id, &o)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	var result string
	cb := func(request *Request, response *Response) {
		defer wg.Done()
		if err := response.Unmarshal(0, &result); err != nil {
			t.Fatal(err)
		}
	}

	feed := s.Start(ctx, 1, cb)

	request, err := NewRequest("id", "reply-to", id, "Join", " ", "hello", "world")
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	wg.Add(1)

	select {
	case feed <- request:
	case <-time.After(1 * time.Second):
		t.Fatal("failed to schedule request")
	}

	wg.Wait()

	if ok := assert.Equal(t, "hello world", result); !ok {
		t.Error()
	}
}

func TestBaseServerProtocolError(t *testing.T) {
	s := BaseServer{}
	var o T

	id := ObjectID{Name: "calc"}
	s.Register(id, &o)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	var errorMsg string
	cb := func(request *Request, response *Response) {
		defer wg.Done()
		errorMsg = response.Error
	}

	feed := s.Start(ctx, 1, cb)

	request, err := NewRequest("id", "reply-to", id, "DoesNotExist", " ", "hello", "world")
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	wg.Add(1)

	select {
	case feed <- request:
	case <-time.After(1 * time.Second):
		t.Fatal("failed to schedule request")
	}

	wg.Wait()

	if ok := assert.Equal(t, "not a function", errorMsg); !ok {
		t.Error()
	}
}

func TestBaseServerServiceError(t *testing.T) {
	s := BaseServer{}
	var o T

	id := ObjectID{Name: "calc"}
	s.Register(id, &o)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	var result RemoteError
	cb := func(request *Request, response *Response) {
		defer wg.Done()
		if response.Error != "" {
			t.Fatal(response.Error)
		}

		if err := response.Unmarshal(1, &result); err != nil {
			t.Fatal(err)
		}
	}

	feed := s.Start(ctx, 1, cb)

	request, err := NewRequest("id", "reply-to", id, "MakeError")
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	wg.Add(1)

	select {
	case feed <- request:
	case <-time.After(1 * time.Second):
		t.Fatal("failed to schedule request")
	}

	wg.Wait()

	if ok := assert.Equal(t, RemoteError{"we made an error"}, result); !ok {
		t.Error()
	}
}

func TestBaseServerStream(t *testing.T) {
	s := BaseServer{}
	var o T

	id := ObjectID{Name: "calc"}
	s.Register(id, &o)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	var key string
	var obj interface{}
	cb := func(k string, o interface{}) {
		key = k
		obj = o

		cancel()
		wg.Done()
	}

	wg.Add(1)

	s.StartStreams(ctx, cb)
	wg.Wait()

	if ok := assert.Equal(t, "calc.TikTok", key); !ok {
		t.Error()
	}

	if ok := assert.IsType(t, 0, obj); !ok {
		t.Error()
	}
}
