package zbus

// RemoteError is a concrete type used to wrap all errors returned by services
// for example, if a method `f` returns `error` the return.Error() is stored in a RemoteError struct
type RemoteError struct {
	Message string
}

func (r *RemoteError) Error() string {
	return r.Message
}

type ProtocolError struct {
	Message string
}

func (r *ProtocolError) Error() string {
	return r.Message
}
