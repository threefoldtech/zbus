#![deny(missing_debug_implementations)]

#[macro_use]
extern crate log;
#[macro_use]
extern crate redis_async;
#[macro_use]
extern crate futures;

use futures::Future;

use redis_async::error::Error;

use serde::{Deserialize, Serialize};
use serde_bytes::ByteBuf;

use std::convert::TryFrom;

pub mod error;
pub mod red;

pub use self::error::ZBusError;

#[derive(Debug, PartialEq, Eq, Clone, Deserialize, Serialize)]
pub struct ObjectID {
    name: String,
    version: String,
}

impl ObjectID {
    pub fn new(name: &str, version: &str) -> ObjectID {
        ObjectID {
            name: name.into(),
            version: version.into(),
        }
    }
}

impl ToString for ObjectID {
    fn to_string(&self) -> String {
        let mut s = self.name.clone();
        if self.version != "" {
            s += &format!("@{}", self.version);
        }

        s
    }
}

impl From<ObjectID> for String {
    fn from(o: ObjectID) -> Self {
        o.to_string()
    }
}

impl TryFrom<String> for ObjectID {
    type Error = &'static str;

    fn try_from(s: String) -> std::result::Result<Self, Self::Error> {
        let mut parts = s.split("@");

        // get name
        let name = match parts.next() {
            Some(n) => n,
            None => return Err("No name found"),
        };

        // get a possible version, default to "" if nothing is left
        let version = parts.next().unwrap_or_default();

        // since there should only be 2 parts, the iterator must be empty here
        if let Some(_) = parts.next() {
            return Err("Unrecognized part for objectID");
        }

        Ok(ObjectID::new(name, version))
    }
}

#[derive(Debug, PartialEq, Eq, Clone, Deserialize, Serialize)]
#[serde(from = "String")]
#[serde(into = "String")]
pub struct Id(String);

impl Id {
    fn new() -> Self {
        Id(uuid::Uuid::new_v4().to_string())
    }
}

impl From<String> for Id {
    fn from(s: String) -> Self {
        Id(s)
    }
}

impl From<Id> for String {
    fn from(id: Id) -> Self {
        id.0
    }
}

#[derive(Debug, PartialEq, Eq, Deserialize, Serialize)]
pub struct Response {
    id: Id,
    arguments: Vec<Vec<u8>>,
    error: String,
}

#[derive(Debug, PartialEq, Eq, Deserialize, Serialize)]
#[serde(rename_all = "PascalCase")]
pub struct Request {
    #[serde(rename = "ID")]
    id: Id,
    arguments: Vec<ByteBuf>,
    object: ObjectID,
    reply_to: Id,
    method: String,
}

impl Request {
    fn new(
        id: Id,
        reply_to: Id,
        object: ObjectID,
        method: String,
        arguments: Vec<ByteBuf>,
    ) -> Self {
        Request {
            object,
            reply_to,
            method,
            id,
            arguments,
        }
    }
}

#[derive(Debug, Deserialize, Serialize)]
pub struct Event {}

pub trait Client {
    fn request<'a>(
        &'a self,
        module: &str,
        object: ObjectID,
        method: &str,
        args: Vec<ByteBuf>,
    ) -> Box<dyn Future<Item = Response, Error = ZBusError> + 'a>;
    fn stream<'a>(
        &mut self,
        module: &str,
        object_id: ObjectID,
        event: &str,
    ) -> Box<dyn Future<Item = Iterator<Item = Event>, Error = ZBusError>>;
}

pub trait Server {
    fn run(&self) -> Box<dyn Future<Item = (), Error = ZBusError>>;
    fn register<T>(
        &self,
        object_id: ObjectID,
        object: T,
    ) -> Box<dyn Future<Item = (), Error = ZBusError>>;
}

pub type Result<T> = std::result::Result<T, ZBusError>;

#[cfg(test)]
mod tests {
    use super::*;
    use rmp_serde::{decode, encode};

    #[test]
    fn test_request_encoding_roundtrip() {
        let id = Id::new();
        let request = Request::new(
            id.clone(),
            id.clone(),
            ObjectID::new("test", "0.0"),
            "testmethod".into(),
            vec![],
        );

        eprintln!("Raw request: {:#?}", request);
        let encoded = encode::to_vec(&request).expect("Failed to encode");
        eprintln!("Encoded request: {:?}", encoded);
        let decoded: Request = decode::from_slice(&encoded).expect("Failed to decode");

        assert_eq!(decoded, request);
    }

    #[test]
    fn test_encoding_port() {
        #[derive(Debug, Serialize, Deserialize)]
        struct T {
            name: String,
            age: f64,
        }

        let arg = T {
            name: "Azmy".into(),
            age: 36.,
        };
        let obj_id = ObjectID::new("object", "1.0");
        let args = vec![
            ByteBuf::from(encode::to_vec_named(&"arg1").unwrap()),
            //ByteBuf::from(encode::to_vec_named(&2u64).unwrap()),
            ByteBuf::from(encode::to_vec_named(&arg).unwrap()),
        ];
        let request = Request::new(
            Id("my-id".into()),
            Id("".into()),
            obj_id,
            "DoSomething".into(),
            args,
        );

        let encoded = encode::to_vec_named(&request).expect("Failed to encode request");
        assert_eq!(
            encoded,
            //vec![
            //    133, 162, 73, 68, 165, 109, 121, 45, 105, 100, 169, 65, 114, 103, 117, 109, 101,
            //    110, 116, 115, 147, 196, 5, 164, 97, 114, 103, 49, 196, 9, 211, 0, 0, 0, 0, 0, 0,
            //    0, 2, 196, 24, 130, 164, 78, 97, 109, 101, 164, 65, 122, 109, 121, 163, 65, 103,
            //    101, 203, 64, 66, 0, 0, 0, 0, 0, 0, 166, 79, 98, 106, 101, 99, 116, 130, 164, 78,
            //    97, 109, 101, 166, 111, 98, 106, 101, 99, 116, 167, 86, 101, 114, 115, 105, 111,
            //    110, 163, 49, 46, 48, 167, 82, 101, 112, 108, 121, 84, 111, 160, 166, 77, 101, 116,
            //    104, 111, 100, 171, 68, 111, 83, 111, 109, 101, 116, 104, 105, 110, 103
            //]
            vec![
                133, 162, 73, 68, 165, 109, 121, 45, 105, 100, 169, 65, 114, 103, 117, 109, 101,
                110, 116, 115, 146, 196, 5, 164, 97, 114, 103, 49, 196, 24, 130, 164, 78, 97, 109,
                101, 164, 65, 122, 109, 121, 163, 65, 103, 101, 203, 64, 66, 0, 0, 0, 0, 0, 0, 166,
                79, 98, 106, 101, 99, 116, 130, 164, 78, 97, 109, 101, 166, 111, 98, 106, 101, 99,
                116, 167, 86, 101, 114, 115, 105, 111, 110, 163, 49, 46, 48, 167, 82, 101, 112,
                108, 121, 84, 111, 160, 166, 77, 101, 116, 104, 111, 100, 171, 68, 111, 83, 111,
                109, 101, 116, 104, 105, 110, 103,
            ]
        );
    }

    #[test]
    fn test_decode_go() {
        #[derive(Debug, Deserialize, Serialize)]
        struct T {
            name: String,
            age: f64,
        }

        //let data = vec![
        //    133, 162, 73, 68, 165, 109, 121, 45, 105, 100, 169, 65, 114, 103, 117, 109, 101, 110,
        //    116, 115, 147, 196, 5, 164, 97, 114, 103, 49, 196, 9, 211, 0, 0, 0, 0, 0, 0, 0, 2, 196,
        //    24, 130, 164, 78, 97, 109, 101, 164, 65, 122, 109, 121, 163, 65, 103, 101, 203, 64, 66,
        //    0, 0, 0, 0, 0, 0, 166, 79, 98, 106, 101, 99, 116, 130, 164, 78, 97, 109, 101, 166, 111,
        //    98, 106, 101, 99, 116, 167, 86, 101, 114, 115, 105, 111, 110, 163, 49, 46, 48, 167, 82,
        //    101, 112, 108, 121, 84, 111, 160, 166, 77, 101, 116, 104, 111, 100, 171, 68, 111, 83,
        //    111, 109, 101, 116, 104, 105, 110, 103,
        //];=
        let data = vec![
            133, 162, 73, 68, 165, 109, 121, 45, 105, 100, 169, 65, 114, 103, 117, 109, 101, 110,
            116, 115, 146, 196, 5, 164, 97, 114, 103, 49, 196, 24, 130, 164, 78, 97, 109, 101, 164,
            65, 122, 109, 121, 163, 65, 103, 101, 203, 64, 66, 0, 0, 0, 0, 0, 0, 166, 79, 98, 106,
            101, 99, 116, 130, 164, 78, 97, 109, 101, 166, 111, 98, 106, 101, 99, 116, 167, 86,
            101, 114, 115, 105, 111, 110, 163, 49, 46, 48, 167, 82, 101, 112, 108, 121, 84, 111,
            160, 166, 77, 101, 116, 104, 111, 100, 171, 68, 111, 83, 111, 109, 101, 116, 104, 105,
            110, 103,
        ];

        let request: Request =
            decode::from_slice(&data).expect("Failed to decode go-encoded request");
        eprintln!("Decoded request: {:#?}", request);
        panic!();
    }

}
