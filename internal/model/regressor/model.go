package regressor

import (
	"errors"
	"math"

	"github.com/thisiscetin/podpredict/internal/data/metrics"
	"github.com/thisiscetin/podpredict/internal/model"
)

type Model struct {
	feModel *ols
	beModel *ols
	trained bool
}

// New creates a new Model instance.
func NewModel() *Model {
	return &Model{
		feModel: &ols{},
		beModel: &ols{},
		trained: false,
	}
}

// Train fits the model using daily metrics.
func (m *Model) Train(data []metrics.Daily) error {
	X, feY, beY := transformData(data)
	if len(X) == 0 {
		return errors.New("no valid rows with FE/BE pods")
	}

	if err := m.feModel.Fit(X, feY); err != nil {
		return err
	}
	if err := m.beModel.Fit(X, beY); err != nil {
		return err
	}
	m.trained = true
	return nil
}

// Predict predicts FE and BE pods for a single feature.
func (m *Model) Predict(f *model.Features) (model.FEPods, model.BEPods, error) {
	if !m.trained {
		return 0, 0, errors.New("model not trained")
	}
	x := []float64{1, f.GMV, f.Users, f.MarketingCost} // intercept
	return model.FEPods(round(m.feModel.Predict(x))), model.BEPods(round(m.beModel.Predict(x))), nil
}

func transformData(data []metrics.Daily) (X [][]float64, feY []float64, beY []float64) {
	for _, d := range data {
		if !d.HasPods() {
			continue
		}
		fe, be, _ := d.Pods()
		features := d.Features()
		xRow := make([]float64, len(features)+1)
		xRow[0] = 1.0 // intercept
		copy(xRow[1:], features)
		X = append(X, xRow)
		feY = append(feY, float64(fe))
		beY = append(beY, float64(be))
	}
	return
}

func round(f float64) int {
	return int(math.Round(f))
}
