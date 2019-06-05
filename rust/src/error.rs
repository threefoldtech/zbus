use redis_async::error::Error as RedisError;
use rmp_serde::{decode, encode};

#[derive(Debug)]
pub enum ZBusError {
    RedisError(RedisError),
    EncodingError(encode::Error),
    DecodingError(decode::Error),
    Void,
}

impl std::fmt::Display for ZBusError {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        match self {
            ZBusError::RedisError(e) => write!(f, "Redis error: {}", e),
            ZBusError::EncodingError(e) => write!(f, "Encoding error: {}", e),
            ZBusError::DecodingError(e) => write!(f, "Decoding error: {}", e),
            ZBusError::Void => write!(f, "Void error"),
        }
    }
}

impl std::error::Error for ZBusError {}

impl From<RedisError> for ZBusError {
    fn from(e: RedisError) -> Self {
        ZBusError::RedisError(e)
    }
}

impl From<encode::Error> for ZBusError {
    fn from(e: encode::Error) -> Self {
        ZBusError::EncodingError(e)
    }
}

impl From<decode::Error> for ZBusError {
    fn from(e: decode::Error) -> Self {
        ZBusError::DecodingError(e)
    }
}

impl From<()> for ZBusError {
    fn from(e: ()) -> Self {
        ZBusError::Void
    }
}
