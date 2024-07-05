package cache

import (
	"errors"
	"time"

	prefix2 "github.com/khgame/memstore/prefix"

	"github.com/redis/go-redis/v9"
)

type (
	H map[string]any

	Cache struct {
		*redis.Client
	}
)

var (
	MaxRetries       = 3
	DialTimeout      = 100 * time.Millisecond
	ReadWriteTimeout = 100 * time.Millisecond
	PoolSize         = 200
	PoolTimeout      = 100 * time.Millisecond
	IdleTimeout      = 60 * time.Minute

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

		PoolSize:        PoolSize,
		PoolTimeout:     PoolTimeout,
		ConnMaxIdleTime: IdleTimeout,
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
	KeyPrefix := prefix2.Prefix(prefix2.Prefix(prefix).ColonStr())
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
