package regressor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOLS_FitPredict_SingleFeature(t *testing.T) {
	X := [][]float64{
		{1, 1},
		{1, 2},
		{1, 3},
	}
	y := []float64{2, 3, 4}

	ols := &ols{}
	err := ols.Fit(X, y)
	assert.NoError(t, err)

	pred := ols.Predict([]float64{1, 4})
	assert.InDelta(t, pred, 8.71, 0.01)
}

func TestOLS_FitPredict_MultipleFeatures(t *testing.T) {
	X := [][]float64{
		{1, 1, 2},
		{1, 2, 1},
		{1, 3, 3},
	}
	y := []float64{5, 6, 10}

	ols := &ols{}
	err := ols.Fit(X, y)
	assert.NoError(t, err)

	pred := ols.Predict([]float64{5, 1, 2})
	assert.InDelta(t, pred, 44.92, 0.01)
}

func TestOLS_Fit_InvalidData(t *testing.T) {
	ols := &ols{}

	// zero-length X
	err := ols.Fit([][]float64{}, []float64{})
	assert.Error(t, err)

	// mismatched lengths
	X := [][]float64{{1, 2}, {3, 4}}
	y := []float64{1}
	err = ols.Fit(X, y)
	assert.Error(t, err)
}

func TestOLS_Predict_NotTrained(t *testing.T) {
	ols := &ols{}
	// predict without training
	pred := ols.Predict([]float64{1, 2})
	assert.Equal(t, 0.0, pred)
}

func TestOLS_Predict_AfterTraining(t *testing.T) {
	X := [][]float64{
		{1, 1},
		{1, 2},
		{1, 3},
	}
	y := []float64{2, 3, 4}

	ols := &ols{}
	err := ols.Fit(X, y)
	assert.NoError(t, err)

	preds := []float64{
		ols.Predict([]float64{1, 0}),
		ols.Predict([]float64{1, 1}),
		ols.Predict([]float64{1, 2}),
	}
	for _, p := range preds {
		assert.True(t, p > 0)
	}
}

func TestOLS_Predict_MultiDimensionalConsistency(t *testing.T) {
	X := [][]float64{
		{1, 1, 2},
		{1, 2, 1},
		{1, 3, 3},
	}
	y := []float64{4, 5, 8}

	ols := &ols{}
	err := ols.Fit(X, y)
	assert.NoError(t, err)

	pred1 := ols.Predict([]float64{1, 2, 2})
	pred2 := ols.Predict([]float64{1, 2, 2})
	assert.Equal(t, pred1, pred2)
}
