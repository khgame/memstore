package memstore

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrResourceNotFound is returned when a resource is not found
	ErrResourceNotFound = fmt.Errorf("resource not found")
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = fmt.Errorf("user not found")

	// ErrInvalidInput is returned when the input is invalid
	ErrInvalidInput = fmt.Errorf("invalid input")
	// ErrInvalidUser is returned when the user is invalid
	ErrInvalidUser = fmt.Errorf("invalid user")
	// ErrStatusError is returned when the status is invalid
	ErrStatusError = fmt.Errorf("status error")
)

type (
	UID = string

	// DataMap is a map that maps a resource's saving name to the resource
	DataMap[T any] map[string]T

	// SavingFunc is a function that saves the storage to permanent storage
	SavingFunc[T any] func(storageName string, data DataMap[T]) error

	// Dumper is a function that dumps memory data to a permanent storage,
	// or loads data from a permanent storage to memory
	Dumper[T any] interface {
		// Dump dumps memory data to a permanent storage
		Dump(ctx context.Context, storageName string, data map[UID]DataMap[T]) error
		// Load loads data from a permanent storage to memory
		Load(ctx context.Context, storageName string, out *map[UID]DataMap[T]) error
	}

	// InMemoryStorage is an in-memory implementation of Storage
	InMemoryStorage[TData StorableType] struct {
		StorageName string

		// mu is a mutex that protects the data map
		mu sync.RWMutex
		// data is the actual data map
		data map[UID]DataMap[TData]

		// dirty is a flag that indicates if the storage has been modified since
		dirty bool
		// saveTime is the last time the storage was saved
		saveTime int64

		// Dumper is a function that dumps memory data to a permanent storage,
		Dumper Dumper[TData]
	}
)

// NewInMemoryStorage creates a new instance of InMemoryResourceStorage
func NewInMemoryStorage[DataType StorableType](storageName string) *InMemoryStorage[DataType] {
	return &InMemoryStorage[DataType]{
		StorageName: storageName,
		data:        make(map[UID]DataMap[DataType]),
	}
}

// Get retrieves a resource for a given user
func (s *InMemoryStorage[DataType]) Get(user string, out *DataType) error {
	// validate input
	if user == "" {
		return fmt.Errorf("%w, user cannot be empty", ErrInvalidUser)
	}
	if out == nil {
		return fmt.Errorf("%w, output cannot be nil", ErrInvalidInput)
	}
	// lock the mutex
	s.mu.RLock()
	defer s.mu.RUnlock()

	storeName := (*out).StoreName()
	// get the resources of the user
	r, ok := s.data[user]
	if !ok {
		return fmt.Errorf("%w, user: %s", ErrUserNotFound, user)
	}

	// get the resource
	*out = r[storeName]

	return nil
}

// List retrieves all resources' StoreName() for a given user
func (s *InMemoryStorage[DataType]) List(user UID) ([]string, error) {
	// validate input
	if user == "" {
		return nil, fmt.Errorf("%w, user cannot be empty", ErrInvalidUser)
	}
	// lock the mutex
	s.mu.RLock()
	defer s.mu.RUnlock()

	// get the resources of the user
	res, ok := s.data[user]
	if !ok {
		return nil, fmt.Errorf("%w, user: %s", ErrResourceNotFound, user)
	}

	// get the resource names
	ret := make([]string, 0, len(res))
	for k := range res {
		ret = append(ret, k)
	}

	return ret, nil
}

// Set stores a resource for a given user
func (s *InMemoryStorage[DataType]) Set(user string, in *DataType) error {
	// validate input
	if user == "" {
		return fmt.Errorf("%w, user cannot be empty", ErrInvalidUser)
	}
	if in == nil {
		return fmt.Errorf("%w, output cannot be nil", ErrInvalidInput)
	}
	// lock the mutex
	s.mu.Lock()
	defer s.mu.Unlock()

	// mark the storage as dirty
	s.dirty = true

	storeName := (*in).StoreName()

	// get the resources of the user
	r, ok := s.data[user]
	if !ok {
		r = make(DataMap[DataType])
		s.data[user] = r
	}

	// store the resource
	r[storeName] = *in

	return nil
}

// Delete deletes a resource for a given user
func (s *InMemoryStorage[DataType]) Delete(user string, storeName string) error {
	// validate input
	if user == "" {
		return fmt.Errorf("%w, user cannot be empty", ErrInvalidUser)
	}
	if storeName == "" {
		return fmt.Errorf("%w, storeName cannot be empty", ErrInvalidInput)
	}
	// lock the mutex
	s.mu.Lock()
	defer s.mu.Unlock()

	// mark the storage as dirty
	s.dirty = true

	// get the resources of the user
	r, ok := s.data[user]
	if !ok {
		return nil
	}

	// delete the resource
	delete(r, storeName)

	// :: if the user has no more resources, do not delete the user
	return nil
}

// IsDirty returns true if the storage has been modified since
func (s *InMemoryStorage[DataType]) IsDirty() bool {
	// lock the mutex
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.dirty
}

// Save persists the storage to permanent storage
// if the storage is not dirty, this function does nothing
func (s *InMemoryStorage[DataType]) Save(ctx context.Context) error {
	// lock the mutex
	s.mu.Lock()
	defer s.mu.Unlock()

	// if the storage is not dirty, do nothing
	if !s.dirty {
		return nil
	}

	// if the dumper is set, dump the data to permanent storage
	if s.Dumper != nil {
		// dump the data to permanent storage
		if err := s.Dumper.Dump(ctx, s.StorageName, s.data); err != nil {
			return fmt.Errorf("failed to dump data to permanent storage, err: %w", err)
		}
	}

	// mark the storage as clean
	s.dirty = false

	// update the save time
	s.saveTime = time.Now().Unix()
	return nil
}

// Load loads the storage from permanent storage
func (s *InMemoryStorage[DataType]) Load(ctx context.Context) error {
	// lock the mutex
	s.mu.Lock()
	defer s.mu.Unlock()

	// check if the storage is dirty
	if s.dirty {
		return fmt.Errorf("%w, cannot load data when storage is dirty", ErrStatusError)
	}

	// if the dumper is not set, do nothing
	if s.Dumper == nil {
		return nil
	}

	// load the data from permanent storage
	if err := s.Dumper.Load(ctx, s.StorageName, &s.data); err != nil {
		return fmt.Errorf("failed to load data from permanent storage, err: %w", err)
	}

	// set the save time, since we are loading from permanent storage
	// we assume the data is clean, so we set the save time to now
	s.saveTime = time.Now().Unix()
	return nil
}
