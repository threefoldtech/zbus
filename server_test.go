package zbus

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	os.Exit(m.Run())
}

func TestBaseServer(t *testing.T) {
	s := BaseServer{}
	var o T

	id := ObjectID{Name: "calc"}
	s.Register(id, &o)

	ctx, shutdown := context.WithCancel(context.Background())
	defer shutdown()
	var result string
	loader := Loader{
		&result,
	}
	cb := func(request *Request, response *Response) {
		if err := response.Unmarshal(&loader); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	feed := s.Start(ctx, &wg, 1, cb)

	request, err := NewRequest("id", "reply-to", id, "Join", " ", "hello", "world")
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	select {
	case feed <- request:
	case <-time.After(1 * time.Second):
		t.Fatal("failed to schedule request")
	}
	shutdown()
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

	ctx, shutdown := context.WithCancel(context.Background())
	defer shutdown()
	var errorMsg *string
	cb := func(request *Request, response *Response) {
		errorMsg = response.Error
	}
	var wg sync.WaitGroup
	feed := s.Start(ctx, &wg, 1, cb)

	request, err := NewRequest("id", "reply-to", id, "DoesNotExist", " ", "hello", "world")
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	select {
	case feed <- request:
	case <-time.After(1 * time.Second):
		t.Fatal("failed to schedule request")
	}

	shutdown()
	wg.Wait()

	if ok := assert.Equal(t, "not a function", *errorMsg); !ok {
		t.Error()
	}
}

func TestBaseServerServiceError(t *testing.T) {
	s := BaseServer{}
	var o T

	id := ObjectID{Name: "calc"}
	s.Register(id, &o)

	ctx, shutdown := context.WithCancel(context.Background())
	defer shutdown()
	var result error
	cb := func(request *Request, response *Response) {

		if response.Error != nil {
			t.Fatal(*response.Error)
		}

		result = response.CallError()
	}
	var wg sync.WaitGroup
	feed := s.Start(ctx, &wg, 1, cb)

	request, err := NewRequest("id", "reply-to", id, "MakeError")
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	select {
	case feed <- request:
	case <-time.After(1 * time.Second):
		t.Fatal("failed to schedule request")
	}

	shutdown()
	wg.Wait()

	if ok := assert.Equal(t, &CallError{"we made an error"}, result); !ok {
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
