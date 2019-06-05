//use crate::{ObjectID, Result, Server};
//
//#[derive(Debug)]
//pub struct RedisServer {
//    cl: redis::Client,
//    module: String,
//}
//
//impl RedisServer {
//    pub fn new(url: &str, module: String) -> RedisServer {
//        RedisServer {
//            cl: redis::Client::open(url).expect("Failed to connect to redis instance"),
//            module,
//        }
//    }
//}
//
//impl Server for RedisServer {
//    fn run(&self) -> Result<()> {
//        unimplemented!();
//    }
//
//    fn register<T>(&self, object_id: ObjectID, object: T) -> Result<()> {
//        unimplemented!();
//    }
//}
