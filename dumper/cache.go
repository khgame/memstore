package dumper

import (
	"context"
	"fmt"

	"github.com/bagaking/goulp/jsonex"

	"github.com/redis/go-redis/v9"

	"github.com/khgame/memstore"
	"github.com/khgame/memstore/cache"
)

type (
	// CacheDumper - a memory store saving algorithm
	// should implement the memstore.Dumper[T any] interface
	CacheDumper[T any] struct {
		Cache *cache.Cache
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
		Cache: c,
	}
}

// CreateCacheDumperByRedisInstance - create a CacheDumper algorithm instance of given type T
func CreateCacheDumperByRedisInstance[T any](c *redis.Client) *CacheDumper[T] {
	return CreateCacheDumperByCacheInstance[T](cache.NewClientByRedisCli(c))
}

// Dump - dump the data to the cache
func (m *CacheDumper[T]) Dump(ctx context.Context, permanentKey string, data map[memstore.UID]memstore.DataMap[T]) error {
	makeKey := SchemeMemStoreSaving.Partial(permanentKey)

	keysLst := make([]string, 0, len(data))

	// use pipeline to save data, group by data length
	// set expire time to forever
	err := m.Cache.BatchSave(ctx,
		func(fn func(key, v string) error) error {
			for uid, v := range data {
				str, err := jsonex.Marshal(v)
				if err != nil {
					return err
				}
				if err = fn(makeKey(uid), string(str)); err != nil {
					return err
				}
				keysLst = append(keysLst, uid)
			}
			return nil
		}, 0)
	if err != nil {
		return err
	}

	// marshal the key list
	strLst, err := jsonex.Marshal(keysLst)
	if err != nil {
		return err
	}
	return m.Cache.Set(ctx, makeKey("__index"), strLst, 0).Err()
}

// Load - load the data from the cache
func (m *CacheDumper[T]) Load(ctx context.Context, permanentKey string, data *map[memstore.UID]memstore.DataMap[T]) error {
	makeKey := SchemeMemStoreSaving.Partial(permanentKey)

	// load index
	var keys []string
	cmd := m.Cache.Get(ctx, makeKey("__index"))
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("get index of storage %s error: %w", permanentKey, err)
	}
	if err := jsonex.Unmarshal([]byte(cmd.Val()), &keys); err != nil {
		return fmt.Errorf("unmarshal index of storage %s error: %w", permanentKey, err)
	}

	// load data
	for _, uid := range keys {
		get := m.Cache.Get(ctx, makeKey(uid))
		if err := get.Err(); err != nil {
			return err
		}
		var v memstore.DataMap[T]
		if err := jsonex.Unmarshal([]byte(get.Val()), &v); err != nil {
			return err
		}
		(*data)[uid] = v
	}

	return nil
}
