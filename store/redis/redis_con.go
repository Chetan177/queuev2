package redis

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

//Connection - address to pass in newPool Connection
type Connection struct {
	address           string
	port              string
	pool              *redis.Pool
	connectionChannel chan bool
	isRedisConnected  bool
	expireTime        string
}

var (
	once      sync.Once
	redisConn *Connection
)

// NewStore - to configure the redis connection
func NewStore(address, port string) *Connection {
	once.Do(func() {
		redisConn = &Connection{
			address:           address,
			port:              port,
			connectionChannel: make(chan bool),
		}

		redisConn.expireTime = "86400"
		go redisConn.reconnectListener()
		err := redisConn.connect()
		if err != nil {
			redisConn.connectionChannel <- false
		}

	})
	return redisConn
}

func (c *Connection) reconnectListener() {
	// default keep-alive config
	duration := 20
	missCount := 3
	maxRetryCount := 8

	d := time.Duration(duration) * time.Second
	t := time.NewTimer(d)
	count := 0

	for {
		select {
		case c.isRedisConnected = <-c.connectionChannel:
			if c.isRedisConnected {
				count = 0
			} else {
				count += missCount //fast reconnect attempt
			}
		case <-t.C:
			count++
		}

		t.Reset(d)
		if count >= missCount {
			c.connectWithExitOnError(maxRetryCount)

			count = 0
			t.Reset(d)
		}
	}
}

func (c *Connection) isConnected() bool {
	conn, err := c.getConnFromPool()
	if err != nil {
		return false
	}

	if _, err = conn.Do("ping"); err != nil {
		return false
	}

	return true
}

func (c *Connection) connectWithExitOnError(maxRetryCount int) {
	if c.isRedisConnected {
		return
	}

	for i := 1; i <= maxRetryCount; i++ {
		duration := c.delayInMS(i)
		time.Sleep(duration)
		err := c.connect()
		if err == nil && c.isConnected() {
			return
		}
	}

	os.Exit(1)
}

//connect - connect to redis
func (c *Connection) connect() error {
	c.pool = c.newPool()

	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}

func (c *Connection) getConnFromPool() (redis.Conn, error) {
	if c.pool == nil {
		return nil, errors.New("connection pool is nil")
	}

	conn := c.pool.Get()
	err := conn.Err()
	c.notifyReconnectListener(err == nil)
	if err != nil {
		return nil, errors.New("failed to get redis conn from pool: " + err.Error() + "\n")
	}
	return conn, nil
}

func (c *Connection) notifyReconnectListener(isErrorNil bool) {
	select {
	case c.connectionChannel <- isErrorNil:
		if !isErrorNil {
		}
	default:
	}
}

func (c *Connection) newPool() *redis.Pool {
	return &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: 80,
		// max number of connections
		MaxActive: 12000,
		// Dial is an application supplied function for creating and
		// configuring a connection.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", c.address+":"+c.port)
			return c, err
		},
	}
}

func (c *Connection) delayInMS(attempt int) time.Duration {
	newDelay := (1 << attempt) * 100
	return time.Duration(newDelay) * time.Millisecond
}
