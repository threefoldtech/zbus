# Z-Bus compatibility tests

In this directory you will find some test you can use when you implement Z-Bus in another language and you want to test it against the reference golang implementation

The interface you need to implement is the one presented in the example: [See interface](../examples/server/api/api.go)

## Server implementation test

To test your server implementation, you need to implement example interface and run the tests in `server_test.go`

Make sure you have :

- redis-running on `localhost:6379`
- your server running and connected to the redis server

Then run the go test in the server directory:

```
go test -v
```

You should have a full passed results:
```
 go test -v
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
=== RUN   TestPow
--- PASS: TestPow (0.00s)
=== RUN   TestDivide
--- PASS: TestDivide (0.00s)
=== RUN   TestAvg
--- PASS: TestAvg (0.00s)
=== RUN   TestCapitalize
--- PASS: TestCapitalize (0.00s)
=== RUN   TestTickTock
--- PASS: TestTickTock (5.59s)
PASS
ok  	github.com/threefoldtech/zbus/compatibility/server	5.600s
```

## Client implementation test

To test your client implementation, the easiest things to do is to implement some test in your language and use the example server to test against.

You can use the [server_test.go](server/server_test.go) file as inspiration for your test implementation.

Make sure you have :

- redis-running on `localhost:6379`
- the example server is running and connected to the redis server

Then run your tests.
