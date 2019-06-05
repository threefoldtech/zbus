use crate::{Client, Event, Id, ObjectID, Request, Response, ZBusError};
use futures::{future, prelude::*};
use redis_async::{client, resp::FromResp};
use rmp_serde::{decode, encode};
use serde_bytes::ByteBuf;
use std::net::SocketAddr;

pub struct RedisClient {
    addr: std::net::SocketAddr,
    con: client::paired::PairedConnection,
}

impl RedisClient {
    pub fn new(addr: SocketAddr) -> impl Future<Item = RedisClient, Error = ZBusError> {
        // clone address and move into a future so we can join and have the value later
        let addr_fut = future::ok(addr.clone());
        client::paired_connect(&addr)
            .map_err(|e| e.into())
            .join(addr_fut)
            .map(|(con, addr)| RedisClient { addr, con })
    }
}

fn get_response(
    con: client::PairedConnection,
    id: Id,
) -> impl Future<Item = Response, Error = ZBusError> {
    trace!("Getting response for Id {:?}", id);
    con.send(resp_array!["BLPOP", id.0, 0])
        .map_err(|e| e.into())
        .and_then(|resp: Vec<u8>| decode::from_slice(&resp).map_err(|e| e.into()))
}

impl Client for RedisClient {
    fn request(
        self,
        module: &str,
        object: ObjectID,
        method: &str,
        args: Vec<ByteBuf>,
    ) -> Box<dyn Future<Item = Response, Error = ZBusError>> {
        let queue = format!("{}.{}", module, object.to_string());
        trace!("Pushing to queue {}", queue);

        let id = Id::new();

        let payload = match encode::to_vec(&Request::new(
            id.clone(),
            id.clone(),
            object,
            method.into(),
            args,
        )) {
            Ok(p) => p,
            Err(e) => return Box::new(future::err(e.into())),
        };

        let con = self.con.clone();

        let push = con
            .send(resp_array!["RPUSH", queue, payload])
            .map_err(|e| e.into());

        Box::new(push.and_then(|_: u64| get_response(con, id)))
    }

    fn stream<'a>(
        &mut self,
        module: &str,
        object_id: ObjectID,
        event: &str,
    ) -> Box<dyn Future<Item = Box<dyn Stream<Item = Event, Error = ZBusError>>, Error = ZBusError>>
    {
        let channel = future::ok(format!("{}.{}.{}", module, object_id.to_string(), event));

        let pubsub = client::pubsub_connect(&self.addr)
            .join(channel)
            .and_then(|(chan, channel_name)| chan.subscribe(&channel_name))
            .map_err(|e| e.into())
            .and_then(|stream| future::ok(RedisEventSubscriber::new(stream)));

        Box::new(pubsub)
    }
}

impl std::fmt::Debug for RedisClient {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        write!(f, "Redis PubSub")
    }
}

pub struct RedisEventSubscriber {
    chan: client::pubsub::PubsubStream,
}

impl RedisEventSubscriber {
    fn new(chan: client::pubsub::PubsubStream) -> Box<dyn Stream<Item = Event, Error = ZBusError>> {
        Box::new(RedisEventSubscriber { chan })
    }
}

impl Stream for RedisEventSubscriber {
    type Item = Event;
    type Error = ZBusError;

    fn poll(&mut self) -> Poll<Option<Self::Item>, Self::Error> {
        let value: Vec<u8> = match try_ready!(self.chan.poll()) {
            None => return Ok(Async::Ready(None)),
            Some(v) => Vec::<u8>::from_resp(v)?,
        };

        let event: Event = decode::from_slice(&value)?;

        Ok(Async::Ready(Some(event)))
    }
}

impl std::fmt::Debug for RedisEventSubscriber {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        write!(f, "Redis event subscriber")
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::net::{IpAddr, Ipv6Addr, SocketAddr};

    #[test]
    fn test_client_connection() {
        let cl = RedisClient::new(SocketAddr::new(IpAddr::V6(Ipv6Addr::LOCALHOST), 6379));

        // ...
    }
}
