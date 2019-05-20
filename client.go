package zbus

import "context"

// Client defines client interface
type Client interface {
	// Request makes a request and return the response data
	Request(module string, object ObjectID, method string, args ...interface{}) (*Response, error)

	// Stream listens to a stream of events from the server
	Stream(ctx context.Context, module string, object ObjectID, event string) (<-chan Event, error)
}
