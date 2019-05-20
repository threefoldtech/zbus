![travis](https://travis-ci.com/threefoldtech/zbus.svg?branch=master) [![codecov](https://codecov.io/gh/threefoldtech/zbus/branch/master/graph/badge.svg)](https://codecov.io/gh/threefoldtech/zbus) [![GoDoc](https://godoc.org/github.com/threefoldtech/zbus?status.svg)](https://godoc.org/github.com/threefoldtech/zbus)

# Motivation
A light weight bus replacement for local inter-process communication. The main goal is to decouple separate
components from each other, by using a light-weight message bus (current implemented redis), to queue and
send message to the separate component that can serve it.

# Goal
- Each module has a name, a single module can host one or more objects
- While it's not required an object can implement one or more interfaces
- Each object must have a name and a version
- Interfaces are mainly used to generate client stubs, but it's totally fine to not have one. In that case the client
  must know precisely the method signature (name and arguments number and types). Same for the return value.
- A consumer who has connection to the message broker can call methods on the remote objects, knowing only the module name, object name, method name, and argument list. The current implementation of the client supports only synchronous calls. In that matter it's similar to RPC.
- A consumer of the component can use a stub to abstract the calls to the remote module
- Support for events where clients can listen to announcements from different components

# Installation
Installing the zbus compiler `zbusc`

```bash
go install github.com/threefoldtech/zbus/zbusc
```
# Specs
Please check [specs](specs/readme.md) here

# Usage
It's very simple, check the [examples](examples)

The [api.go](examples/server/api/api.go) have some `go generate` lines that runs the zbusc tool