package regressor

import (
	"errors"
)

// ols implements a simple ordinary least squares linear regression.
type ols struct {
	coeffs []float64
}

// Fit trains the OLS model on features X and targets y.
func (o *ols) Fit(X [][]float64, y []float64) error {
	if len(X) == 0 || len(y) == 0 || len(X) != len(y) {
		return errors.New("invalid training data")
	}

	cols := len(X[0])
	o.coeffs = make([]float64, cols)

	// naive diagonal approximation
	for j, _ := range o.coeffs {
		sumXY := 0.0
		sumXX := 0.0
		for i, row := range X {
			sumXY += row[j] * y[i]
			sumXX += row[j] * row[j]
		}
		o.coeffs[j] = sumXY / sumXX
	}
	return nil
}

// Predict returns the predicted value for a feature vector x.
func (o *ols) Predict(x []float64) float64 {
	if o.coeffs == nil || len(x) != len(o.coeffs) {
		return 0
	}
	sum := 0.0
	for i, val := range x {
		sum += val * o.coeffs[i]
	}
	return sum
}
