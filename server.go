package zbus

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	log "github.com/rs/zerolog/log"
)

var (
	// NoOP request will cause the worker to try polling again from the queue
	// without doing anything. The idea is that we can use this to check
	// if there are any free workers, by pusing this to the channel in a select
	// and see if any of the workers receives it.
	NoOP Request

	statusObjectID = ObjectID{Name: "zbus", Version: "1.0"}
)

const (
	// WorkerFree free state
	WorkerFree WorkerState = "free"
	// WorkerBusy busy state
	WorkerBusy WorkerState = "busy"
)

// Callback defines a callback method signature for responses
type Callback func(request *Request, response *Response)

// EventCallback is calld by the base server once an event is available
type EventCallback func(key string, event interface{})

// WorkerState represents curret worker state (free, or busy)
type WorkerState string

// WorkerStatus represents the full worker status including request time and method
// that it is working on.
type WorkerStatus struct {
	State     WorkerState `json:"state" yaml:"state"`
	StartTime time.Time   `json:"time,omitempty" yaml:"time,omitempty"`
	Action    string      `json:"action,omitempty" yaml:"action,omitempty"`
}

// Status is returned by the server Status method
type Status struct {
	Objects []ObjectID     `json:"objects" yaml:"objects"`
	Workers []WorkerStatus `json:"workers" yaml:"workers"`
}

// BaseServer implements the basic server functionality
// In case you are building your own zbus server
type BaseServer struct {
	objects map[ObjectID]*Surrogate
	m       sync.RWMutex

	status  []WorkerStatus
	statusM sync.RWMutex
}

// Register registers an object on server
func (s *BaseServer) Register(id ObjectID, object interface{}) error {
	//validate objects methods goes here
	// wrap object in an abstract wrapper
	s.m.Lock()
	defer s.m.Unlock()

	if id.String() == statusObjectID.String() {
		return fmt.Errorf("object id is reserved")
	}

	if s.objects == nil {
		s.objects = make(map[ObjectID]*Surrogate)
	}

	if _, ok := s.objects[id]; ok {
		return fmt.Errorf("object already exists")
	}

	s.objects[id] = NewSurrogate(object)
	return nil
}

func (s *BaseServer) call(request *Request) (ret Output, err error) {
	s.m.RLock()

	surrogate, ok := s.objects[request.Object]
	s.m.RUnlock()

	if !ok {
		return ret, fmt.Errorf("unknown object")
	}

	defer func() {
		if p := recover(); p != nil {
			stack := debug.Stack()
			fmt.Println(string(stack))
			log.Error().Msg(string(stack))
			err = fmt.Errorf("remote method call %s.%s() paniced: %s", request.Object, request.Method, p)
		}
	}()

	return surrogate.CallRequest(request)
}

func (s *BaseServer) process(request *Request) *Response {
	ret, err := s.call(request)
	var msg string
	if err != nil {
		msg = err.Error()
	}

	return NewResponse(request.ID, ret, msg)
}

func (s *BaseServer) statusIn(id uint, request *Request) {
	s.statusM.Lock()
	defer s.statusM.Unlock()

	s.status[id] = WorkerStatus{
		State:     WorkerBusy,
		StartTime: time.Now(),
		Action:    fmt.Sprintf("[%s].%s()", request.Object.String(), request.Method),
	}
}

func (s *BaseServer) statusOut(id uint) {
	s.statusM.Lock()
	defer s.statusM.Unlock()

	s.status[id] = WorkerStatus{
		State:     WorkerFree,
		StartTime: time.Now(),
	}
}

func (s *BaseServer) worker(ctx context.Context, id uint, wg *sync.WaitGroup, ch <-chan *Request, cb Callback) {
	defer wg.Done()
	s.statusOut(id)

	for {
		select {
		case request := <-ch:
			if request == nil {
				//channel has been closed
				return
			} else if request == &NoOP {
				continue
			}

			s.statusIn(id, request)
			response := s.process(request)
			s.statusOut(id)

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

// Status returns a copy of the internal worker status
func (s *BaseServer) Status() Status {
	s.statusM.RLock()
	s.m.RLock()

	defer s.m.RUnlock()
	defer s.statusM.RUnlock()

	ids := make([]ObjectID, 0, len(s.objects))
	for id := range s.objects {
		ids = append(ids, id)
	}

	results := make([]WorkerStatus, len(s.status))
	for i := 0; i < len(s.status); i++ {
		results[i] = s.status[i]
	}

	return Status{Objects: ids, Workers: results}
}

// Start starts the workers. Workers will call cb with results of requests. the call will
// feed requests to workers by feeding requests to channel.
// panics if workers number is zero.
func (s *BaseServer) Start(ctx context.Context, wg *sync.WaitGroup, workers uint, cb Callback) chan<- *Request {
	if workers == 0 {
		panic("invalid number of workers")
	}

	s.status = make([]WorkerStatus, workers)
	ch := make(chan *Request)
	var id uint
	for ; id < workers; id++ {
		wg.Add(1)
		go s.worker(ctx, id, wg, ch, cb)
	}

	return ch
}
