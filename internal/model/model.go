package model

// Model is an interface for predicting FEPods and BEPods.
type Model interface {
	// Predict predicts FEPods and BEPods for given features.
	Predict(features []float64) (int, int)
}
