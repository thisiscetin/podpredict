package inmemory

import (
	"context"
	"sync"

	"github.com/thisiscetin/podpredict/internal/store"
)

type impl struct {
	predictions []store.Prediction
	sync.RWMutex
}

func (i *impl) Append(ctx context.Context, r store.Prediction) error {
	i.Lock()
	defer i.Unlock()

	i.predictions = append(i.predictions, r)
	return nil
}

func (i *impl) List(_ context.Context) ([]store.Prediction, error) {
	i.RLock()
	defer i.RUnlock()

	out := make([]store.Prediction, len(i.predictions))
	copy(out, i.predictions)
	return out, nil
}

func NewStore() store.Store {
	return &impl{
		predictions: make([]store.Prediction, 0),
	}
}
