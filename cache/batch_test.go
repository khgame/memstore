package cache_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/khgame/memstore/cache"
	"github.com/stretchr/testify/assert"
)

func createMockCache() *cache.Cache {
	// create mini redis server
	mini, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	return cache.NewClient(mini.Addr())
}

// Test_PipeGet tests the PipeGet method with testify
func Test_PipeGet(t *testing.T) {
	cache := createMockCache()
	defer cache.Close()

	ctx := context.Background()
	// set data
	cache.Set(ctx, "uid001", "res001", 0)
	cache.Set(ctx, "uid002", "res002", 0)
	cache.Set(ctx, "uid003", "res003", 0)
	cache.Set(ctx, "uid004", "res004", 0)
	cache.Set(ctx, "uid005", "res005", 0)

	// test PipeGet

	valuesHit, keysMiss, err := cache.PipeGet(ctx, "uid001")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(valuesHit))
	assert.Equal(t, 0, len(keysMiss))
	assert.Equal(t, "res001", valuesHit[0])

	valuesHit, keysMiss, err = cache.PipeGet(ctx, "uid001", "uid002", "uid003", "uid004", "uid005")
	assert.NoError(t, err)
	assert.Equal(t, 5, len(valuesHit))
	assert.Equal(t, 0, len(keysMiss))
	assert.Equal(t, "res001", valuesHit[0])
	assert.Equal(t, "res002", valuesHit[1])
	assert.Equal(t, "res003", valuesHit[2])
	assert.Equal(t, "res004", valuesHit[3])
	assert.Equal(t, "res005", valuesHit[4])

	valuesHit, keysMiss, err = cache.PipeGet(ctx, "uid001", "uidAAA")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(valuesHit))
	assert.Equal(t, 1, len(keysMiss))
	assert.Equal(t, "res001", valuesHit[0])
	assert.Equal(t, "uidAAA", keysMiss[0])
}

// Test_BatchSave tests the BatchSave method with testify
func Test_BatchSave(t *testing.T) {
	cache := createMockCache()
	defer cache.Close()

	ctx := context.Background()

	testMap := map[string]interface{}{
		"uid001": "res001",
		"uid002": "res002",
		"uid003": "res003",
		"uid004": 1,
		"uid005": struct {
			Name string
		}{
			Name: "res005",
		},
	}

	// test BatchSave
	err := cache.BatchSave(ctx, func(fn func(key string, v any) error) error {
		for k, v := range testMap {
			err := fn(k, v)
			if err != nil {
				return err
			}
		}
		return nil
	}, 0)
	assert.NoError(t, err)

	// test PipeGet
	valuesHit, keysMiss, err := cache.PipeGet(ctx, "uid001", "uid002", "uid003", "uid004", "uid005")
	assert.NoError(t, err)
	assert.Equal(t, 5, len(valuesHit))
	assert.Equal(t, 0, len(keysMiss))
	assert.Equal(t, "res001", valuesHit[0])
	assert.Equal(t, "res002", valuesHit[1])
	assert.Equal(t, "res003", valuesHit[2])
	assert.Equal(t, "1", valuesHit[3])
	assert.Equal(t, "{\"Name\":\"res005\"}", valuesHit[4])
}
