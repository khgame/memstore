package dumper_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/khgame/memstore"
	"github.com/khgame/memstore/dumper"
	"github.com/stretchr/testify/assert"
)

type (
	// TestDataType is a test type that implements StorableType
	TestDataType struct {
		Name     string
		Quantity int64
	}
)

func createCacheDumper() memstore.Dumper[TestDataType] {
	// create mini redis server
	mini, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	return dumper.CreateCacheDumperByAddr[TestDataType](mini.Addr())
}

// Test_Dump tests the Dump method of CacheDumper with testify
func Test_Dump(t *testing.T) {
	dp := createCacheDumper()
	ctx := context.Background()
	err := dp.Dump(ctx, "test_storage", map[memstore.UID]memstore.DataMap[TestDataType]{
		"uid001": {
			"res001": {
				Name:     "res001",
				Quantity: 1,
			},
			"res002": {
				Name:     "res002",
				Quantity: 200,
			},
		},
	})

	assert.NoError(t, err)
	// get
	cmd := dp.(*dumper.CacheDumper[TestDataType]).Cache.Get(ctx, dumper.SchemeMemStoreSaving.Make("test_storage", "uid001"))
	assert.NoError(t, cmd.Err())
	assert.Equal(t, `{"res001":{"Name":"res001","Quantity":1},"res002":{"Name":"res002","Quantity":200}}`, cmd.Val())
}
