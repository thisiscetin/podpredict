package store

import (
	"context"
	"time"

	"github.com/thisiscetin/podpredict/internal/model"
)

// Prediction represents a single model output and its metadata.
// Each prediction contains a unique ID, a UTC timestamp indicating when
// it was generated, the input feature values used to compute it, and
// the resulting predicted number of front-end (FE) and back-end (BE) pods.
// Prediction values are immutable after creation and can be safely copied.
type Prediction struct {
	// ID uniquely identifies this prediction. It is typically a UUID string.
	ID string `json:"id"`

	// Timestamp records when the prediction was created, in UTC.
	Timestamp time.Time `json:"timestamp"`

	// Input contains the model feature values (e.g., GMV, Users, MarketingCost)
	// that produced this prediction.
	Input model.Features `json:"input"`

	// FEPods is the predicted number of front-end pods required.
	FEPods int `json:"fe_pods"`

	// BEPods is the predicted number of back-end pods required.
	BEPods int `json:"be_pods"`
}

// Store defines the interface for persisting and retrieving predictions.
// Implementations must be safe for concurrent use by multiple goroutines.
// The interface is intentionally minimal to allow flexible backends such
// as in-memory, database, or remote stores.
type Store interface {
	// Append adds a new prediction record to the store.
	// Implementations should ensure the operation is atomic and safe
	// for concurrent writers. The provided context can be used to
	// cancel the operation early if supported by the backend.
	Append(ctx context.Context, r Prediction) error

	// List returns all stored predictions.
	// Implementations should return a defensive copy so that callers
	// can modify the result without affecting internal state. The
	// provided context can be used to cancel the operation early
	// if supported by the backend.
	List(ctx context.Context) ([]Prediction, error)
}
