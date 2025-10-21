package model

import "github.com/thisiscetin/podpredict/internal/metrics"

// Features represents the input features used for prediction.
type Features struct {
	GMV           float64 `json:"gmv"`            // Gross Merchandise Value
	Users         float64 `json:"users"`          // Number of users
	MarketingCost float64 `json:"marketing_cost"` // Marketing expenditure
}

// BEPods represents the predicted number of Back-End Pods.
type BEPods int

// FEPods represents the predicted number of Front-End Pods.
type FEPods int

// Model defines an interface for predicting FEPods and BEPods based on given features.
type Model interface {
	// Train trains the model using the provided daily metrics.
	Train([]metrics.Daily) error
	// Predict takes Features as input and returns the predicted FEPods, BEPods, and any potential error.
	Predict(features *Features) (FEPods, BEPods, error)
}
