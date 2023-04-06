package memstore

import (
	"testing"

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

// Test_InMemStorage_SetGetListDelete tests the Get & Set method of InMemStorage with testify
func Test_InMemStorage_SetGetListDelete(t *testing.T) {
	storage := NewInMemoryStorage[TestDataType]("test_storage")
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
