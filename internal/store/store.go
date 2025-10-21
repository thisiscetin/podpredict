package store

import (
	"context"
	"time"

	"github.com/thisiscetin/podpredict/internal/model"
)

type Prediction struct {
	Timestamp time.Time      `json:"timestamp"`
	Input     model.Features `json:"input"`
	FEPods    int            `json:"fe_pods"`
	BEPods    int            `json:"be_pods"`
}

type Store interface {
	Append(ctx context.Context, r Prediction) error
	List(ctx context.Context) ([]Prediction, error)
}
