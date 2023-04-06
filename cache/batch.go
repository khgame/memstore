package cache

import (
	"context"
	"sync"
	"time"

	"github.com/bagaking/goulp/jsonex"
	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/errgroup"
)

func (cli *Cache) PipeGet(ctx context.Context, keys ...string) (valuesHit []string, keysMiss []string, err error) {
	if len(keys) == 0 {
		return nil, nil, nil
	}

	pipe := cli.Client.Pipeline()
	defer pipe.Close()

	cmds := make(map[string]*redis.StringCmd, len(keys))
	for _, key := range keys {
		cmds[key] = pipe.Get(ctx, key)
	}

	if _, err = pipe.Exec(ctx); err != nil {
		return nil, nil, err
	}

	hit, miss := make([]string, 0, len(keys)), make([]string, 0, len(keys))
	for k, cmd := range cmds {
		value, errCmd := cmd.Result()
		if errCmd != nil {
			if IsRedisNil(errCmd) {
				miss = append(miss, k)
				continue
			}
			return nil, nil, err
		}
		hit = append(hit, value)
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

func (cli *Cache) BatchSave(ctx context.Context, forEachReceiver func(fn func(key string, v any) error) error, expiration time.Duration) error {
	// create pipeline
	p := cli.Pipeline()
	dataLen := 0
	err := forEachReceiver(func(key string, v any) error {
		ret, err := jsonex.Marshal(v)
		if err != nil {
			return err
		}
		p.Set(ctx, key, string(ret), expiration)
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
