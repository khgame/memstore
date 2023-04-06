package dumper

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/khgame/memstore"
	"github.com/khgame/memstore/cache"
)

type (
	// CacheDumper - a memory store saving algorithm
	// should implement the memstore.Dumper[T any] interface
	CacheDumper[T any] struct {
		c *cache.Cache
	}
)

const (
	SchemeMemStoreSaving memstore.KeyScheme = "store:%s:%s"
)

var _ memstore.Dumper[any] = (*CacheDumper[any])(nil)

// CreateCacheDumperByAddr - create a CacheDumper algorithm instance of given type T
func CreateCacheDumperByAddr[T any](addr string) *CacheDumper[T] {
	return CreateCacheDumperByCacheInstance[T](cache.NewClient(addr))
}

// CreateCacheDumperByCacheInstance - create a CacheDumper algorithm instance of given type T
func CreateCacheDumperByCacheInstance[T any](c *cache.Cache) *CacheDumper[T] {
	return &CacheDumper[T]{
		c: c,
	}
}

// CreateCacheDumperByRedisInstance - create a CacheDumper algorithm instance of given type T
func CreateCacheDumperByRedisInstance[T any](c *redis.Client) *CacheDumper[T] {
	return CreateCacheDumperByCacheInstance[T](cache.NewClientByRedisCli(c))
}

// Dump - dump the data to the cache
func (m *CacheDumper[T]) Dump(ctx context.Context, storageName string, data map[memstore.UID]memstore.DataMap[T]) error {
	makeKey := SchemeMemStoreSaving.Partial(storageName)

	keysLst := make([]string, 0, len(data))

	// use pipeline to save data, group by data length
	// set expire time to forever
	err := m.c.BatchSave(ctx,
		func(fn func(key string, v any) error) error {
			for uid, v := range data {
				if err := fn(makeKey(uid), v); err != nil {
					return err
				}
				keysLst = append(keysLst, uid)
			}
			return nil
		}, 0)
	if err != nil {
		return err
	}

	return m.c.Set(ctx, makeKey("__index"), keysLst, 0).Err()
}

// Load - load the data from the cache
func (m *CacheDumper[T]) Load(ctx context.Context, storageName string, data *map[memstore.UID]memstore.DataMap[T]) error {
	makeKey := SchemeMemStoreSaving.Partial(storageName)

	// load index
	var keys []string
	m.c.Get(ctx, makeKey("__index")).Scan(&keys)

	// load data
	for _, uid := range keys {
		get := m.c.Get(ctx, makeKey(uid))
		if err := get.Err(); err != nil {
			return err
		}
		var v memstore.DataMap[T]
		if err := get.Scan(&v); err != nil {
			return err
		}
		(*data)[uid] = v
	}
	return nil
}
