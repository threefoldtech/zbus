package zbus

import "context"

//Server is server interface
type Server interface {
	Register(id ObjectID, object interface{}) error
	Run(ctx context.Context) error
}
