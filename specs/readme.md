# Line Protocol
## Request

```json
{
    // ID is the request ID, usually a guid
    "ID": "request-id",
    // Arguments is a list of the arguments where each element
    // is a msgpack serialized bytes of the argument
    "Arguments": [], 
    "Object": {
        "Name": "object-name",
        "Version": "object-version",
    },
    "ReplyTo": "response id",
    "Method": "actual method to call"
}
```
- The full object is again serialized as another msgpack bytes. before it's pushed to the msg broker.
- The request is pushed to a `<module>.<object>@<version>` queue
- Once the request is handled, a response is pushed back to the `ReplyTo` queue.

## Response 

```json
{
    // ID is the request ID, usually a guid. same as the the request id
    // that initiated this response
    "ID": "request-id",
    // Arguments is a list of the returns where each element
    // is a msgpack serialized bytes of the argument
    "Arguments": [], 
    "Error": "protocol error message"
}
```

# Events Stream 
- Objects can publish events to listeners an event can be any chunk of bytes that is published to certain key
- In redis implementation, we use the `PUBSUB` feature.
- Events are pushed to `<module>.<object>@<version>.<event>` channel.

## Note on Events
In Go/Redis implementation, the Server interface can define a method that accept a single argument `context.Context` and return 
a `chan T` where T is msgpack serializeable.

The redis server will understand that this is a stream/events method and will run it once the `server.Run` method is called. Your
event data (served over the `chan`) will be published to the write redis channel.

The generated Stub will have a stream stub method that u can call, to get another channel that subscribe and serve the published data
in the correct type.