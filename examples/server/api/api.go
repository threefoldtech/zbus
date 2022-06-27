package api

import (
	"context"
	"time"
)

//go:generate rm -rf ../stubs
//go:generate mkdir ../stubs

//Calculator the calcuator interface
//go:generate zbusc -module server -version 1.0 -name calculator -package stubs github.com/threefoldtech/zbus/examples/server/api+Calculator ../stubs/calcuator_stub.go
type Calculator interface {
	Add(a, b float64) float64
	AddSub(a, b float64) (float64, float64)
	Divide(a, b float64) (float64, error)
	Pow(a, b float64) float64
	Avg(a []float64) float64
}

//go:generate zbusc -module server -version 1.0 -name utils -package stubs github.com/threefoldtech/zbus/examples/server/api+Utils ../stubs/utils_stub.go
type Utils interface {
	Capitalize(s string) string
	Tuple() (int, string, float64, error)
	TikTok(ctx context.Context) <-chan time.Time // event
	Panic() int
	Sleep(t time.Duration) error
}
