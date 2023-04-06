package memstore

import "context"

type (
	// Storage is an interface that all storage implementations must implement
	Storage[DataType StorableType] interface {
		// Get retrieves a resource for a given user
		Get(user string, out *DataType) error
		// List retrieves all resources' StoreName() for a given user
		List(user string) ([]string, error)

		// Set sets a resource for a given user
		Set(user string, in *DataType) error
		// Delete deletes a resource for a given user
		Delete(user string, storeName string) error

		// IsDirty returns true if the storage has been modified since
		IsDirty() bool

		// Save persists the storage to permanent storage
		Save(ctx context.Context) error

		// Load loads the storage from permanent storage
		Load(ctx context.Context) error
	}

	// StorableType is an interface that all types that can be stored must implement
	StorableType interface {
		StoreName() string
	}
)
