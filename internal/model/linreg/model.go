package linreg

import (
	"errors"
	"math"

	"github.com/sajari/regression"
	"github.com/thisiscetin/podpredict/internal/metrics"
	"github.com/thisiscetin/podpredict/internal/model"
)

// linearModel implements Model using github.com/sajari/regression.
type linearModel struct {
	fe *regression.Regression
	be *regression.Regression
	// trained ensures Predict() is only available after a successful Train()
	trained bool
}

// NewModel returns a Model backed by sajari/regression.
func NewModel() model.Model {
	return &linearModel{
		fe: &regression.Regression{},
		be: &regression.Regression{},
	}
}

// Train builds two independent linear regressors for FE and BE pods.
func (m *linearModel) Train(rows []metrics.Daily) error {
	fe := &regression.Regression{}
	be := &regression.Regression{}

	fe.SetObserved("FEPods")
	be.SetObserved("BEPods")

	// Variable order matches Features() {GMV, Users, MarketingCost}
	fe.SetVar(0, "GMV")
	fe.SetVar(1, "Users")
	fe.SetVar(2, "MarketingCost")

	be.SetVar(0, "GMV")
	be.SetVar(1, "Users")
	be.SetVar(2, "MarketingCost")

	n := 0
	for _, d := range rows {
		if !d.HasPods() {
			continue
		}
		fePods, bePods, _ := d.Pods()
		x := d.Features() // []float64{GMV, Users, MarketingCost}

		fe.Train(regression.DataPoint(float64(fePods), x))
		be.Train(regression.DataPoint(float64(bePods), x))
		n++
	}
	if n == 0 {
		return errors.New("no valid rows with FE/BE pods")
	}

	if err := fe.Run(); err != nil {
		return err
	}
	if err := be.Run(); err != nil {
		return err
	}

	m.fe = fe
	m.be = be
	m.trained = true
	return nil
}

// Predict returns rounded FE/BE pod counts for the supplied features.
// Guarantees a minimum of 1 pod for both FE and BE.
func (m *linearModel) Predict(f *model.Features) (model.FEPods, model.BEPods, error) {
	if !m.trained {
		return 0, 0, errors.New("model not trained")
	}
	in := []float64{f.GMV, f.Users, f.MarketingCost}

	fe, err := m.fe.Predict(in)
	if err != nil {
		return 0, 0, err
	}
	be, err := m.be.Predict(in)
	if err != nil {
		return 0, 0, err
	}

	feInt := clampMinInt(safeRound(fe), 1)
	beInt := clampMinInt(safeRound(be), 1)
	return model.FEPods(feInt), model.BEPods(beInt), nil
}

// safeRound handles NaN/Inf defensively before rounding.
func safeRound(v float64) int {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 1 // fallback to minimum if solver returns invalid
	}
	return int(math.Round(v))
}

func clampMinInt(v, min int) int {
	if v < min {
		return min
	}
	return v
}
