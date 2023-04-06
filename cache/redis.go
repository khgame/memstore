package cache

import (
	"errors"
	"github.com/go-redis/redis/v8"
	"time"
)

type (
	H map[string]any

	Cache struct {
		*redis.Client
	}
)

var (
	MaxRetries         = 3
	DialTimeout        = 100 * time.Millisecond
	ReadWriteTimeout   = 100 * time.Millisecond
	PoolSize           = 200
	PoolTimeout        = 100 * time.Millisecond
	IdleTimeout        = 60 * time.Minute
	IdleCheckFrequency = time.Minute

	defaultAddr   string
	defaultClient *Cache
)

func NewClient(addr string) *Cache {
	opts := &redis.Options{
		Addr:         addr,
		MaxRetries:   MaxRetries,
		DialTimeout:  DialTimeout,
		ReadTimeout:  ReadWriteTimeout,
		WriteTimeout: ReadWriteTimeout,

		PoolSize:           PoolSize,
		PoolTimeout:        PoolTimeout,
		IdleTimeout:        IdleTimeout,
		IdleCheckFrequency: IdleCheckFrequency,
	}

	return NewClientByRedisCli(redis.NewClient(opts))
}

func NewClientByRedisCli(cli *redis.Client) *Cache {
	return &Cache{
		Client: cli,
	}
}

func NewPrefixedCli(prefix string) *Cache {
	if prefix == "" {
		return defaultClient
	}
	c := NewClient(defaultAddr)
	KeyPrefix := Prefix(Prefix(prefix).ColonStr())
	c.AddHook(KeyPrefix)
	return c
}

func Client() *Cache {
	return defaultClient
}

func Init(addr string) {
	defaultAddr = addr
	defaultClient = NewClient(addr)
}

func IsRedisNil(err error) bool {
	return errors.Is(err, redis.Nil)
}

func SingleFlightAPI() SingleFlight {
	return defaultClient.SingleFlightAPI()
}
