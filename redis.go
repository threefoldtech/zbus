package zbus

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"

	"github.com/garyburd/redigo/redis"
)

func newRedisPool(address string) (*redis.Pool, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}
	var host string
	switch u.Scheme {
	case "tcp":
		host = u.Host
	case "unix":
		host = u.Path
	default:
		return nil, fmt.Errorf("unknown scheme '%s' expecting tcp or unix", u.Scheme)
	}
	var opts []redis.DialOption

	if u.User != nil {
		opts = append(
			opts,
			redis.DialPassword(u.User.Username()),
		)
	}

	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial(u.Scheme, host, opts...)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) > 10*time.Second {
				//only check connection if more than 10 second of inactivity
				_, err := c.Do("PING")
				return err
			}

			return nil
		},
		MaxActive:   10,
		IdleTimeout: 1 * time.Minute,
		Wait:        true,
	}, nil
}

// RedisServer implementation for Redis
type RedisServer struct {
	BaseServer
	queue   string
	pool    *redis.Pool
	workers uint
}

// NewRedisServer builds a new ZBus server that uses disque as message broker
func NewRedisServer(queue, address string, workers uint) (Server, error) {
	if workers == 0 {
		return nil, fmt.Errorf("invalid number of workers")
	}

	pool, err := newRedisPool(address)
	if err != nil {
		return nil, err
	}

	con := pool.Get()
	defer con.Close()

	if _, err := con.Do("PING"); err != nil {
		return nil, fmt.Errorf("could not establish connection: %s", err)
	}

	return &RedisServer{queue: queue, pool: pool, workers: workers}, nil
}

func (s *RedisServer) cb(request *Request, response *Response) {
	con := s.pool.Get()
	defer con.Close()
	payload, err := response.Encode()
	if err != nil {
		log.WithError(err).Error("failed to encode response")
	}

	if err := con.Send("RPUSH", request.ReplyTo, payload); err != nil {
		log.WithError(err).Error("failed to send response")
	}
}

func (s *RedisServer) getNext() ([]byte, error) {
	con := s.pool.Get()
	defer con.Close()

	payload, err := redis.ByteSlices(con.Do("BLPOP", s.queue, 10))
	if err != nil {
		return nil, err
	}

	if payload == nil || len(payload) < 2 {
		return nil, redis.ErrNil
	}

	return payload[1], nil
}

// Run starts the ZBus server
func (s *RedisServer) Run(ctx context.Context) error {
	ch := s.Start(ctx, s.workers, s.cb)
	for {
		payload, err := s.getNext()
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err == redis.ErrNil {
			continue
		} else if err != nil {
			log.WithError(err).Error("failed to get next job. Retrying in 1 second")
			<-time.After(1 * time.Second)
			continue
		}

		request, err := LoadRequest(payload)
		if err != nil {
			log.WithError(err).Error("failed to load request object")
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- request:
		}
	}
}

// RedisClient is client implementation for redis broker
type RedisClient struct {
	pool *redis.Pool
}

// NewRedisClient creates a new redis client
func NewRedisClient(address string) (Client, error) {
	pool, err := newRedisPool(address)
	if err != nil {
		return nil, err
	}

	return &RedisClient{pool}, nil
}

// Request makes a request to object.Method hosted by module. A module name is the queue name used in the server part.
func (c *RedisClient) Request(module string, object ObjectID, method string, args ...interface{}) (*Response, error) {
	id := uuid.New().String()
	request, err := NewRequest(id, id, object, method, args...)
	if err != nil {
		return nil, err
	}

	payload, err := request.Encode()
	if err != nil {
		return nil, err
	}

	con := c.pool.Get()
	defer con.Close()

	if err := con.Send("RPUSH", module, payload); err != nil {
		return nil, err
	}

	// wait for response
	return c.getResponse(con, id)
}

func (c *RedisClient) getResponse(con redis.Conn, id string) (*Response, error) {
	//TODO: a timeout or an exit strategy is required here in case
	//the response never came back

	payload, err := redis.ByteSlices(con.Do("BLPOP", id, 0))
	if err != nil {
		return nil, err
	}

	if payload == nil || len(payload) < 2 {
		return nil, redis.ErrNil
	}

	response, err := LoadResponse(payload[1])
	if err != nil {
		return nil, err
	}

	if len(response.Error) != 0 {
		return nil, fmt.Errorf(response.Error)
	}

	return response, nil
}
