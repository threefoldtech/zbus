package zbus

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack"

	log "github.com/rs/zerolog/log"

	"github.com/gomodule/redigo/redis"
)

const (
	redisPullTimeout = 10
	redisResponseTTL = 5 * 60 // 5 minutes
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
		MaxActive:   100,
		IdleTimeout: 1 * time.Minute,
		Wait:        true,
	}, nil
}

// RedisServer implementation for Redis
type RedisServer struct {
	BaseServer
	module  string
	pool    *redis.Pool
	workers uint
	running bool
	state   sync.Mutex
}

// NewRedisServer builds a new ZBus server that uses disque as message broker
func NewRedisServer(module, address string, workers uint) (Server, error) {
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

	return &RedisServer{module: module, pool: pool, workers: workers}, nil
}

func (s *RedisServer) cb(request *Request, response *Response) {
	con := s.pool.Get()
	defer con.Close()
	payload, err := response.Encode()
	if err != nil {
		log.Error().Err(err).Msg("failed to encode response")
		return
	}

	if err := con.Send("RPUSH", request.ReplyTo, payload); err != nil {
		log.Error().Err(err).Msg("failed to send response")
		return
	}

	con.Send("EXPIRE", request.ReplyTo, redisResponseTTL)
}

// ecb event callback
func (s *RedisServer) ecb(key string, o interface{}) {
	con := s.pool.Get()
	defer con.Close()
	data, err := msgpack.Marshal(o)
	if err != nil {
		log.Error().Err(err).Msg("failed to encode event")
		return
	}

	key = fmt.Sprintf("%s.%s", s.module, key)

	if err := con.Send("PUBLISH", key, data); err != nil {
		log.Error().Err(err).Msg("failed to send event")
	}
}

func (s *RedisServer) getNext(pullArgs []interface{}) ([]byte, error) {
	con := s.pool.Get()
	defer con.Close()

	payload, err := redis.ByteSlices(con.Do("BLPOP", pullArgs...))
	if err != nil {
		return nil, err
	}

	if payload == nil || len(payload) < 2 {
		return nil, redis.ErrNil
	}

	return payload[1], nil
}

func (s *RedisServer) statusHandler(ctx context.Context) error {
	pullArgs := []interface{}{
		fmt.Sprintf("%s.%s", s.module, statusObjectID),
		redisPullTimeout,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		payload, err := s.getNext(pullArgs)

		if err == redis.ErrNil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			continue
		} else if err != nil {
			log.Error().Err(err).Msg("failed to get next job. Retrying in 1 second")
			<-time.After(1 * time.Second)
			continue
		}

		request, err := LoadRequest(payload)
		if err != nil {
			log.Error().Err(err).Msg("failed to load request object")
			continue
		}

		status, err := returnFromObjects(nil, s.Status())
		if err != nil {
			log.Error().Err(err).Msg("failed to create response")
			continue
		}
		response := NewResponse(request.ID, status, "")

		// send response back
		s.cb(request, response)
	}
}

// Run starts the ZBus server
func (s *RedisServer) Run(ctx context.Context) error {
	//don't run multiple instances at the same time
	s.state.Lock()
	if s.running {
		s.state.Unlock()
		return fmt.Errorf("server is already running")
	}
	var pullArgs []interface{}
	//fill in the queues to pull from, we have a queue per object
	for id := range s.objects {
		pullArgs = append(
			pullArgs,
			fmt.Sprintf("%s.%s", s.module, id),
		)
	}

	s.running = true
	s.state.Unlock()

	//status handler runs in its own worker.
	go s.statusHandler(ctx)

	//start event workers
	s.StartStreams(ctx, s.ecb)

	// now start request/response workers and proxy calls and responses
	pullArgs = append(pullArgs, redisPullTimeout) //the pull timeout
	workerCtx, shutdown := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	ch := s.Start(workerCtx, &wg, s.workers, s.cb)

	defer func() {
		shutdown()
		wg.Wait()
		close(ch)
	}()

	for {
		// wait for free worker before we poll for jobs
		select {
		case ch <- &NoOP:
		case <-ctx.Done():
			return ctx.Err()
		}

		payload, err := s.getNext(pullArgs)

		if err == redis.ErrNil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			continue
		} else if err != nil {
			log.Error().Err(err).Msg("failed to get next job. Retrying in 1 second")
			<-time.After(1 * time.Second)
			continue
		}

		request, err := LoadRequest(payload)
		if err != nil {
			log.Error().Err(err).Msg("failed to load request object")
			continue
		}

		// force wait for a worker to poll
		// the job (since we sure there is one free)
		// we don't allow shutting the workers down here.
		ch <- request
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
	return c.RequestContext(context.Background(), module, object, method, args...)
}

// RequestContext makes a request to object.Method hosted by module. A module name is the queue name used in the server part.
func (c *RedisClient) RequestContext(ctx context.Context, module string, object ObjectID, method string, args ...interface{}) (*Response, error) {
	id := uuid.New().String()
	request, err := NewRequest(id, id, object, method, args...)
	if err != nil {
		return nil, err
	}

	payload, err := request.Encode()
	if err != nil {
		return nil, err
	}

	con, err := c.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer con.Close()
	queue := fmt.Sprintf("%s.%s", module, object)
	if err := con.Send("RPUSH", queue, payload); err != nil {
		return nil, err
	}

	// wait for response
	for {
		response, err := c.getResponse(con, id)
		if err == redis.ErrNil {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				continue
			}
		}

		return response, nil
	}
}

func (c *RedisClient) getResponse(con redis.Conn, id string) (*Response, error) {
	//TODO: a timeout or an exit strategy is required here in case
	//the response never came back

	payload, err := redis.ByteSlices(con.Do("BLPOP", id, 1))
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

	if response.Error != nil {
		return nil, fmt.Errorf(*response.Error)
	}

	return response, nil
}

// Status return module status
func (c *RedisClient) Status(ctx context.Context, module string) (Status, error) {
	response, err := c.RequestContext(ctx, module, statusObjectID, "")
	if err != nil {
		return Status{}, err
	}

	var status Status
	loader := Loader{
		&status,
	}
	if err := response.Unmarshal(&loader); err != nil {
		return status, err
	}

	return status, nil
}

// Stream listens to a stream of events from the server
func (c *RedisClient) Stream(ctx context.Context, module string, object ObjectID, event string) (<-chan Event, error) {
	con, err := c.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s.%s.%s", module, object, event)
	_, err = con.Do("SUBSCRIBE", key)

	if err != nil {
		con.Close()
		return nil, err
	}

	ch := make(chan Event)
	go func(con redis.Conn) {
		defer func() {
			close(ch)
			con.Send("UNSUBSCRIBE")
			con.Close()
		}()

		for {
			message, err := redis.ByteSlices(con.Receive())
			if err != nil {
				log.Error().Err(err).Msgf("failed to get next event for '%s'", key)
				return
			}

			if len(message) != 3 {
				log.Debug().Str("key", key).Msgf("message was of len (%d)", len(message))
				continue
			}
			// problem with cancellation here is that
			// it won't actually happen unless we received
			// a message on the subscribe channel
			select {
			case ch <- Event(message[2]):
			case <-ctx.Done():
				return
			}
		}
	}(con)

	return ch, nil
}
