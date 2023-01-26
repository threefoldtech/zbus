![travis](https://travis-ci.com/threefoldtech/zbus.svg?branch=master) [![codecov](https://codecov.io/gh/threefoldtech/zbus/branch/master/graph/badge.svg)](https://codecov.io/gh/threefoldtech/zbus) [![GoDoc](https://godoc.org/github.com/threefoldtech/zbus?status.svg)](https://godoc.org/github.com/threefoldtech/zbus)

# Motivation
A light weight bus replacement for **local** inter-process communication. The main goal is to decouple separate
components from each other, by using a light-weight message bus (current implemented redis), to queue and
send message to the separate component that can serve it.

The keyword here is **local** zbus is not intended to be used over the network because it's intended ONLY for local inter process
communication. Allows local processes to talk to each other.

To keep it light, the ZBUS does not do any authentication or permissions

A public API then can expose a public API then internally make calls to other local components.


# Overview
- Each `module` has a name, a single `module` can host one or more `objects`
- While it's not required an `object` can implement one or more `interfaces`
- Each object must have a name and a version
- `interfaces` are mainly used to generate client `stubs`, but it's totally fine to not have one. In that case the client
  must know precisely the method `signature` (name and arguments number and types); same for the `return` value.
- A consumer who has connection to the message broker can call methods on the remote objects, knowing only the `module` name, `object` name, method `name`, and `argument` list. The current implementation of the client supports only synchronous calls. In that matter it's similar to RPC.
- A consumer of the component can use a `stub` to abstract the calls to the remote module
- Support for events where clients can listen to announcements from different components

`zbus` was built with `golang` only in mind so `zbusc` was only built to generate client stubs for `golang` from the service interface witch is `golang` interface! But the underlying protocol itself is simple enough to implement client and servers in other languages. For example `rust` hence we also have [rbus](https://github.com/threefoldtech/rbus) which is 100% compatible with the go implementation. Hence it's possible for services built in rust to be called from golang and the vise versa.


# Installation
Installing the zbus compiler `zbusc`

```bash
go install github.com/threefoldtech/zbus/zbusc
```

The `zbusc` is only needed to generate `stub` code.

# Walk-through
Let's build a service from scratch say a `calculator` service.
First we create a project and init it
```bash
mkdir calc
cd calc
go mod init github.com/example/calc
go get github.com/threefoldtech/zbus
```

this initialize the directory to be a go project (module)
> please use a proper module name when doing `mod init`

> All new files are created under the calc directory
let's create the service file
create new file `api.go`
```go
package calc

type Calculator interface {
	Add(a, b float64) float64
	Multiply(n ...float64) float64
	Divide(a, b float64) (float64, error)
}
```

while it's very simple, it shows that implementation supports variadic arguments, and also returning multiple arguments

> zbus can also return channels for event streams but let's leave that for another example

The next step is simple is to actually implement this interface and start our `zbus` server

we create file `server/server.go`

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/example/calc"
	"github.com/threefoldtech/zbus"
)

// the service implementation structure
type myCalculator struct{}

// this is just to verify that that the myCalculator actually
// implements calc.Calculator interface ! if the interface
// changes this should give a compile error
var _ calc.Calculator = (*myCalculator)(nil)

func (c *myCalculator) Add(a, b float64) float64 {
	return a + b
}

func (c *myCalculator) Multiply(n ...float64) float64 {
	var v float64
	if len(n) > 0 {
		v = n[0]
	}

	for _, x := range n[1:] {
		v *= x
	}

	return v
}

func (c *myCalculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("cannot divide by zero")
	}

	return a / b, nil
}

func app() error {
	const module = "calc"
	const address = "tcp://localhost:6379"
	server, err := zbus.NewRedisServer(module, address, 10)
	if err != nil {
		return err
	}

	impl := &myCalculator{}

	// a single module (in this case calc) can serve multiple objects
	// also it can serve multiple objects with same name but different version
	// hence it's important when u register an object to give it a name and a version
	// it's not possible to register the same name@version twice.
	server.Register(zbus.ObjectIDFromString("calculator@1.0.0"), impl)

	// once you are done registering ALL your objects it's time to start your server

	return server.Run(context.Background())
}

func main() {
	if err := app(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}
```

You can now run the zbus server simply by doing
```bash
go run server/server.go
```

## Generating a stub client and using it
While you still inside the `calc` project create a stubs directory
```bash
mkdir stubs
```
Then run the following command
```bash
zbusc -module calc -name calculator -version 1.0.0 -package stubs github.com/example/calc+Calculator stubs/calculator_stub.go
```
The command line is simple it takes the module name, object name, object version, and the package name to use in the generated code. The it needs to know which interface to generate code for (in that case it's the `Calculator` interface) but it requires to know the full path hence it's provided as `github.com/example/calc+Calculator`. Finally where to output the generated code. We output the generated stub to `stubs/calculator_stubs.go`

To avoid typing this command every time you change the interface or you add new methods, instead edit the `api.go` file by adding this line above the Calculator interface

```go

//go:generate mkdir -p stubs
//go:generate zbusc -module calc -name calculator -version 1.0.0 -package stubs github.com/example/calc+Calculator stubs/calculator_stub.go

type Calculator interface {
```

Now each time you want to regenerate the stubs do
```bash
go generate ./...
```

### Testing the generated stub
while under the `calc` project create a client directory
```bash
mkdir client
```
then create file `client/client.go`

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/example/calc/stubs"
	"github.com/threefoldtech/zbus"
)

func app(ctx context.Context) error {
	const address = "tcp://localhost:6379"

	client, err := zbus.NewRedisClient(address)
	if err != nil {
		return err
	}

	stub := stubs.NewCalculatorStub(client)

	// calling the add function
	fmt.Printf("adding 2 numbers: %f \n", stub.Add(ctx, 20, 30))

	_, err = stub.Divide(ctx, 100, 0)
	if err != nil {
		fmt.Println("got error: ", err)
	}

	return nil
}

func main() {
	if err := app(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}
```

You notice the following:
- You create a generic low level client to zbus, then you can use that to create as many stubs (to other services and modules) as you want
- The client does not have to know about the interface, just the stub and then it can do calls normally like any other service.
- Generated stubs calls always take ctx as first argument which allows you to control timeouts and cancellation if call is taking to long (service down?!)

To test this first to this
```bash
go run server/server.go
```

Then in another terminal do
```bash
go run client/client.go
```
this should output this
```
adding 2 numbers: 50.000000
got error:  cannot divide by zero
```

# Specs
Please check [specs](specs/readme.md) here

# Usage
It's very simple, check the [examples](examples)

The [api.go](examples/server/api/api.go) have some `go generate` lines that runs the zbusc tool

# Projects using zbus
- [ZOS](https://github.com/threefoldtech/zos)
