package inmemory

import (
	"context"
	"sync"

	"github.com/thisiscetin/podpredict/internal/store"
)

// impl is a concrete in-memory implementation of the store.Store interface.
// It holds all predictions in a slice protected by a read/write mutex.
// This type is safe for concurrent access by multiple goroutines.
type impl struct {
	predictions []store.Prediction
	sync.RWMutex
}

// Append adds the given prediction to the store.
// It acquires a write lock to ensure thread safety, appends the prediction
// to the internal slice, and releases the lock. Append always returns nil
// unless the method’s signature changes to support future error handling.
// Append is safe for concurrent use.
func (i *impl) Append(_ context.Context, r store.Prediction) error {
	i.Lock()
	defer i.Unlock()

	i.predictions = append(i.predictions, r)
	return nil
}

// List returns a copy of all stored Predictions.
// The returned slice is a defensive copy of the internal state,
// meaning callers can modify it freely without affecting the
// underlying store. Order of elements corresponds to the order
// in which they were appended.
// The provided context is currently unused but is accepted to
// satisfy the store.Store interface.
func (i *impl) List(_ context.Context) ([]store.Prediction, error) {
	i.RLock()
	defer i.RUnlock()

	out := make([]store.Prediction, len(i.predictions))
	copy(out, i.predictions)
	return out, nil
}

// NewStore creates and returns a new, empty in-memory store.
// The store is thread-safe for concurrent appends and listings.
// Because it keeps all data in memory, its contents are lost when
// the process exits or the store is discarded.
func NewStore() store.Store {
	return &impl{
		predictions: make([]store.Prediction, 0),
	}
}
