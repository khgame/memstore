package memstore_test

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

func (t TestDataType) StoreName() string {
	return t.Name
}

func createCacheDumper() memstore.Dumper[TestDataType] {
	// create mini redis server
	mini, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	return dumper.CreateCacheDumperByAddr[TestDataType](mini.Addr())
}

// Test_InMemStorage_SetGetListDelete tests the Get / Set / List / Delete method of InMemStorage with testify
func Test_InMemStorage_SetGetListDelete(t *testing.T) {
	storage := memstore.NewInMemoryStorage[TestDataType]("test_storage")
	// test Set
	storage.Set("uid001", &TestDataType{
		Name:     "res001",
		Quantity: 1,
	})
	storage.Set("uid001", &TestDataType{
		Name:     "res002",
		Quantity: 200,
	})

	// test Get
	data := TestDataType{
		Name: "res001",
	}
	err := storage.Get("uid001", &data)
	assert.NoError(t, err)
	assert.Equal(t, "res001", data.Name)
	assert.Equal(t, int64(1), data.Quantity)

	// test List
	resources, err := storage.List("uid001")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(resources))
	assert.Equal(t, "res001", resources[0])
	assert.Equal(t, "res002", resources[1])

	// test Delete
	err = storage.Delete("uid001", "res001")
	assert.NoError(t, err)
	resources, err = storage.List("uid001")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(resources))
	assert.Equal(t, "res002", resources[0])
}

// Test_InMemStorage_SaveLoad tests the Save & Load method of InMemStorage with testify
func Test_InMemStorage_SaveLoad(t *testing.T) {
	storage := memstore.NewInMemoryStorage[TestDataType]("test_storage")
	storage.Dumper = createCacheDumper()
	// test Set
	storage.Set("uid001", &TestDataType{
		Name:     "res001",
		Quantity: 1,
	})
	storage.Set("uid001", &TestDataType{
		Name:     "res002",
		Quantity: 200,
	})

	// test Get
	data := TestDataType{
		Name: "res001",
	}
	err := storage.Get("uid001", &data)
	assert.NoError(t, err)
	assert.Equal(t, "res001", data.Name)
	assert.Equal(t, int64(1), data.Quantity)

	ctx := context.Background()

	// test Save
	err = storage.Save(ctx)
	assert.NoError(t, err)

	storage2 := memstore.NewInMemoryStorage[TestDataType]("test_storage")
	// test Load
	err = storage2.Load(ctx)
	assert.Error(t, err)
	storage2.Dumper = storage.Dumper

	err = storage2.Load(ctx)
	assert.NoError(t, err)

	// test Get from storage2
	data = TestDataType{
		Name: "res001",
	}
	err = storage2.Get("uid001", &data)
	assert.NoError(t, err)
	assert.Equal(t, "res001", data.Name)
	assert.Equal(t, int64(1), data.Quantity)

	// test List from storage2
	resources, err := storage2.List("uid001")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(resources))
	assert.Equal(t, "res001", resources[0])
	assert.Equal(t, "res002", resources[1])
}
