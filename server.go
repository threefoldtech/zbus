package zbus

import (
	"fmt"
	"sync"
)

type baseServer struct {
	objects map[ObjectID]interface{}
	m       sync.RWMutex
}

func (s *baseServer) Register(id ObjectID, object interface{}) error {
	//validate objects methods goes here
	// wrap object in an abstract wrapper

	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.objects[id]; ok {
		return fmt.Errorf("object already exists")
	}

	s.objects[id] = object
	return nil
}
