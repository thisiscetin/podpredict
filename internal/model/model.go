package model

type Features struct {
	GMV           float64
	Users         float64
	MarketingCost float64
}

type BEPods int
type FEPods int

// Model is an interface for predicting FEPods and BEPods.
type Model interface {
	// Predict predicts FEPods and BEPods for given features.
	Predict(features *Features) (FEPods, BEPods, error)
}
