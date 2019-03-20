package zbus

import (
	"context"
	"fmt"
	"sync"
)

// Callback defines a callback method signature for responses
type Callback func(response *Response, err error)

// BaseServer implements the basic server functionality
// In case you are building your own zbus server
type BaseServer struct {
	objects map[ObjectID]*Surrogate
	m       sync.RWMutex
}

// Register registers an object on server
func (s *BaseServer) Register(id ObjectID, object interface{}) error {
	//validate objects methods goes here
	// wrap object in an abstract wrapper
	s.m.Lock()
	defer s.m.Unlock()

	if s.objects == nil {
		s.objects = make(map[ObjectID]*Surrogate)
	}

	if _, ok := s.objects[id]; ok {
		return fmt.Errorf("object already exists")
	}

	s.objects[id] = NewSurrogate(object)
	return nil
}

func (s *BaseServer) call(request *Request) (*Response, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	surrogate, ok := s.objects[request.Object]

	if !ok {
		return nil, fmt.Errorf("unknown object")
	}

	result, err := surrogate.CallRequest(request)
	if err != nil {
		return nil, err
	}

	return request.Response(result...)
}

func (s *BaseServer) worker(ctx context.Context, ch <-chan *Request, cb Callback) {
	for {
		select {
		case request := <-ch:
			if request == nil {
				//channel has been closed
				break
			}
			cb(s.call(request))
		case <-ctx.Done():
			break
		}
	}
}

// Start starts the workers. Workers will call cb with results of requests. the call will
// feed requests to workers by feeding requests to channel.
// panics if workers number is zero.
func (s *BaseServer) Start(ctx context.Context, workers uint, cb Callback) chan<- *Request {
	if workers == 0 {
		panic("invalid number of workers")
	}

	ch := make(chan *Request)
	var i uint
	for ; i < workers; i++ {
		go s.worker(ctx, ch, cb)
	}

	return ch
}
