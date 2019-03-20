package zbus

// Client defines client interface
type Client interface {
	// Request makes a request and return the response data
	Request(module string, object ObjectID, method string, args ...interface{}) (*Response, error)
}
