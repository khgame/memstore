package cache

import (
	"context"
	"sync"
	"time"

	"github.com/bagaking/goulp/jsonex"
	"github.com/redis/go-redis/v9"

	"golang.org/x/sync/errgroup"
)

// PipeGet get values from redis by pipeline
func (cli *Cache) PipeGet(ctx context.Context, keys ...string) (valuesHit []string, keysMiss []string, err error) {
	lKeys := len(keys)
	if lKeys == 0 {
		return nil, nil, nil
	}

	cmders := make([]*redis.StringCmd, 0, lKeys)
	pipe := cli.Client.Pipeline()
	for _, key := range keys {
		cmders = append(cmders, pipe.Get(ctx, key))
	}

	// exec pipeline and get results of all commands
	if _, err = pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, nil, err
	}

	hit, miss := make([]string, 0, lKeys), make([]string, 0)
	for i, cmd := range cmders {
		// check error, if error is redis.Nil, then record the key as miss
		if errCmd := cmd.Err(); errCmd != nil {
			if IsRedisNil(errCmd) {
				miss = append(miss, keys[i])
				continue
			}
			return nil, nil, err
		}
		// record the value as hit
		hit = append(hit, cmd.Val())
	}

	return hit, miss, nil
}

func (cli *Cache) BatchGet(ctx context.Context, shotLimit int, keys ...string) (valuesHit []string, keysMiss []string, err error) {
	if shotLimit < 1 {
		return nil, nil, ErrShotLimits
	}

	l := len(keys)
	if l < shotLimit {
		return cli.PipeGet(ctx, keys...)
	}

	var g errgroup.Group
	var mu sync.Mutex

	valuesHit, keysMiss = make([]string, 0, len(keys)), make([]string, 0, len(keys))

	for from, to := 0, shotLimit; from < l; from, to = from+shotLimit, to+shotLimit {
		if to > l {
			to = l
		}
		ks := keys[from:to]
		g.Go(func() error {
			h, m, errGet := cli.PipeGet(ctx, ks...)
			if errGet != nil {
				return errGet
			}
			mu.Lock()
			valuesHit, keysMiss = append(valuesHit, h...), append(keysMiss, m...)
			mu.Unlock()
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		return nil, nil, err
	}
	return
}

// BatchSave save data to redis in batch
func (cli *Cache) BatchSave(ctx context.Context, forEachReceiver func(fn func(key string, v any) error) error, expiration time.Duration) error {
	// create pipeline
	p := cli.Pipeline()
	dataLen := 0
	err := forEachReceiver(func(key string, v any) error {
		var ret string
		if str, ok := v.(string); ok {
			ret = str
		} else {
			bytes, err := jsonex.Marshal(v)
			if err != nil {
				return err
			}
			ret = string(bytes)
		}

		p.Set(ctx, key, ret, expiration)
		dataLen += len(key) + len(ret)
		// send data to redis, if dataLen > 500k
		if dataLen > 512*1024 {
			_, e2 := p.Exec(ctx)
			if e2 != nil {
				return e2
			}
			dataLen = 0
			p = cli.Pipeline()
		}
		return nil
	})
	if err != nil {
		return err
	}
	_, err = p.Exec(ctx)
	return err
}
