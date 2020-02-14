package zbus

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	log "github.com/rs/zerolog/log"
)

var (
	// NoOP request will cause the worker to try polling again from the queue
	// without doing anything. The idea is that we can use this to check
	// if there are any free workers, by pusing this to the channel in a select
	// and see if any of the workers receives it.
	NoOP Request
)

// Callback defines a callback method signature for responses
type Callback func(request *Request, response *Response)

// EventCallback is calld by the base server once an event is available
type EventCallback func(key string, event interface{})

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

func (s *BaseServer) call(request *Request) (ret Return, err error) {
	s.m.RLock()

	surrogate, ok := s.objects[request.Object]
	s.m.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown object")
	}

	defer func() {
		if p := recover(); p != nil {
			stack := debug.Stack()
			log.Error().Msg(string(stack))
			err = fmt.Errorf("remote method call %s.%s() paniced: %s", request.Object, request.Method, p)
		}
	}()

	return surrogate.CallRequest(request)
}

func (s *BaseServer) process(request *Request) (*Response, error) {

	ret, err := s.call(request)
	var msg string
	if err != nil {
		msg = err.Error()
	}

	return NewResponse(request.ID, msg, ret...)
}

func (s *BaseServer) worker(ctx context.Context, wg *sync.WaitGroup, ch <-chan *Request, cb Callback) {
	defer wg.Done()
	for {
		select {
		case request := <-ch:
			if request == nil {
				//channel has been closed
				return
			} else if request == &NoOP {
				continue
			}

			response, err := s.process(request)
			if err != nil {
				log.Error().Err(err).Msg("failed to create response object")
				continue
			}

			cb(request, response)
		case <-ctx.Done():
			return
		}
	}
}

func (s *BaseServer) streamWorker(ctx context.Context, key ObjectID, stream Stream, cb EventCallback) {
	fqn := fmt.Sprintf("%s.%s", key, stream.Name())
	for event := range stream.Run(ctx) {
		cb(fqn, event)
	}
}

// StartStreams start the stream (events) workers in the background
// use the ctx to cancel the streams workers
func (s *BaseServer) StartStreams(ctx context.Context, cb EventCallback) {
	s.m.RLock()
	defer s.m.RUnlock()
	for key, obj := range s.objects {
		for _, stream := range obj.Streams() {
			go s.streamWorker(ctx, key, stream, cb)
		}
	}
}

// Start starts the workers. Workers will call cb with results of requests. the call will
// feed requests to workers by feeding requests to channel.
// panics if workers number is zero.
func (s *BaseServer) Start(ctx context.Context, wg *sync.WaitGroup, workers uint, cb Callback) chan<- *Request {
	if workers == 0 {
		panic("invalid number of workers")
	}
	ch := make(chan *Request)
	var i uint
	for ; i < workers; i++ {
		wg.Add(1)
		go s.worker(ctx, wg, ch, cb)
	}

	return ch
}
