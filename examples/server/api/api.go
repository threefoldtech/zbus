package api

//Calculator the calcuator interface
//go:generate zbus -module server -version 1.0 -name calculator -package api github.com/threefoldtech/zbus/examples/server/api+Calculator calcuator_stub.go
type Calculator interface {
	Add(a ...float64) float64
	Pow(a, b float64) float64
	Divide(a, b float64) (float64, error)
	Avg(a []float64) float64
}

//go:generate zbus -module server -version 1.0 -name utils -package api github.com/threefoldtech/zbus/examples/server/api+Utils utils_stub.go
type Utils interface {
	Capitalize(s string) string
}
