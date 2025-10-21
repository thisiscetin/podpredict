package linreg

import (
	"testing"
	"time"

	"github.com/sajari/regression"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thisiscetin/podpredict/internal/model"

	"github.com/thisiscetin/podpredict/internal/metrics"
)

// makeDay constructs a Daily with both FE/BE pods set.
func makeDay(date time.Time, gmv float64, users int, mc float64, fe, be int) metrics.Daily {
	d, err := metrics.NewDaily(date, gmv, users, mc, &fe, &be)
	if err != nil {
		panic(err)
	}
	return d
}

func TestLinearModel_TrainPredict_PerfectFit(t *testing.T) {
	// Ground-truth linear rules (library adds intercept internally):
	// FE =  5 + 0.50*GMV + 2.0*Users + 3.0*MC
	// BE = -2 + 0.20*GMV + 1.0*Users + 4.0*MC
	trueFE := func(gmv float64, users float64, mc float64) float64 {
		return 5 + 0.5*gmv + 2*users + 3*mc
	}
	trueBE := func(gmv float64, users float64, mc float64) float64 {
		return -2 + 0.2*gmv + 1*users + 4*mc
	}

	// Build training rows (full-rank, n > p).
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	rows := []metrics.Daily{
		makeDay(base.AddDate(0, 0, 0), 100, 10, 5, int(trueFE(100, 10, 5)), int(trueBE(100, 10, 5))),
		makeDay(base.AddDate(0, 0, 1), 150, 12, 6, int(trueFE(150, 12, 6)), int(trueBE(150, 12, 6))),
		makeDay(base.AddDate(0, 0, 2), 200, 20, 8, int(trueFE(200, 20, 8)), int(trueBE(200, 20, 8))),
		makeDay(base.AddDate(0, 0, 3), 250, 25, 9, int(trueFE(250, 25, 9)), int(trueBE(250, 25, 9))),
		makeDay(base.AddDate(0, 0, 4), 300, 30, 10, int(trueFE(300, 30, 10)), int(trueBE(300, 30, 10))),
	}

	m := NewModel()
	require.NoError(t, m.Train(rows))

	// In-sample prediction: exact integers → equal after rounding.
	fe, be, err := m.Predict(&model.Features{GMV: 200, Users: 20, MarketingCost: 8})
	require.NoError(t, err)
	assert.Equal(t, 169, int(fe)) // 5+100+40+24
	assert.Equal(t, 90, int(be))  // -2+40+20+32 (rounds to 90; clamp not triggered)

	// New point (still inside training manifold)
	fe2, be2, err := m.Predict(&model.Features{GMV: 180, Users: 18, MarketingCost: 7})
	require.NoError(t, err)
	assert.Equal(t, 152, int(fe2)) // 5+90+36+21
	assert.Equal(t, 80, int(be2))  // -2+36+18+28 → 80
}

func TestLinearModel_Train_ErrorOnNoPods(t *testing.T) {
	m := NewModel()
	base := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	// Both pods nil
	r1, err := metrics.NewDaily(base, 100, 10, 5, nil, nil)
	require.NoError(t, err)
	// Only FE present
	feOnly := 10
	r2, err := metrics.NewDaily(base.AddDate(0, 0, 1), 200, 20, 8, &feOnly, nil)
	require.NoError(t, err)

	err = m.Train([]metrics.Daily{r1, r2})
	require.Error(t, err)
}

func TestLinearModel_RealData(t *testing.T) {
	rows := []metrics.Daily{
		//              GMV  Users  MC    FE   BE
		makeDay(time.Now().AddDate(0, 0, 0), 8182928.00, 89988, 179530, 10, 4),
		makeDay(time.Now().AddDate(0, 0, 1), 8181814.00, 72896, 186044, 10, 4),
		makeDay(time.Now().AddDate(0, 0, 2), 19364132.00, 74540, 148377, 20, 9),
		makeDay(time.Now().AddDate(0, 0, 3), 19904689.00, 79045, 147847, 20, 9),
	}

	m := NewModel()
	require.NoError(t, m.Train(rows))

	fe, be, err := m.Predict(&model.Features{GMV: 9928743.00, Users: 76955, MarketingCost: 187234})
	require.NoError(t, err)
	assert.Equal(t, 10, int(fe))
	assert.Equal(t, 4, int(be))
}

func TestLinearModel_MinClamp_ZeroishPredictionsBecomeOne(t *testing.T) {
	// Train on zero targets to force ~0 predictions; clamp should lift to 1.
	rows := []metrics.Daily{
		makeDay(time.Now().AddDate(0, 0, 0), 10, 5, 1, 0, 0),
		makeDay(time.Now().AddDate(0, 0, 1), 20, 6, 2, 0, 0),
		makeDay(time.Now().AddDate(0, 0, 2), 30, 7, 3, 0, 0),
		makeDay(time.Now().AddDate(0, 0, 3), 40, 8, 4, 0, 0),
		makeDay(time.Now().AddDate(0, 0, 4), 55, 12, 5, 0, 0),
	}

	m := NewModel()
	require.NoError(t, m.Train(rows))

	fe, be, err := m.Predict(&model.Features{GMV: 15.4, Users: 16, MarketingCost: 2})
	require.NoError(t, err)
	assert.Equal(t, 1, int(fe), "FE should be clamped to minimum 1")
	assert.Equal(t, 1, int(be), "BE should be clamped to minimum 1")
}

func TestLinearModel_MinClamp_NegativePredictionsBecomeOne(t *testing.T) {
	// Downward trends so extrapolation can go negative; clamp must lift to 1.
	rows := []metrics.Daily{
		makeDay(time.Now().AddDate(0, 0, 0), 0, 0, 0, 5, 4),
		makeDay(time.Now().AddDate(0, 0, 1), 2, 5, 1, 4, 3),
		makeDay(time.Now().AddDate(0, 0, 2), 5, 10, 2, 3, 2),
		makeDay(time.Now().AddDate(0, 0, 3), 8, 15, 3, 1, 1),
		makeDay(time.Now().AddDate(0, 0, 4), 12, 18, 5, 1, 1),
	}

	m := NewModel()
	require.NoError(t, m.Train(rows))

	fe, be, err := m.Predict(&model.Features{GMV: 20, Users: 40, MarketingCost: 5}) // likely negative raw
	require.NoError(t, err)
	assert.Equal(t, 1, int(fe), "FE negative raw prediction must clamp to 1")
	assert.Equal(t, 1, int(be), "BE negative raw prediction must clamp to 1")
}

func TestLinearModel_InternalCoefficients_AreInExpectedOrder(t *testing.T) {
	// Inspect FE coefficients directly through sajari/regression.
	base := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)

	// FE = 1 + 2*GMV + 3*Users + 4*MC
	rows := []metrics.Daily{
		makeDay(base, 1, 1, 1, 1+2*1+3*1+4*1, 0),
		makeDay(base.AddDate(0, 0, 1), 2, 3, 4, 1+2*2+3*3+4*4, 0),
		makeDay(base.AddDate(0, 0, 2), 5, 6, 7, 1+2*5+3*6+4*7, 0),
		makeDay(base.AddDate(0, 0, 3), 8, 2, 9, 1+2*8+3*2+4*9, 0),
	}

	// Train FE only; BE ignored.
	m := &linearModel{fe: &regression.Regression{}, be: &regression.Regression{}}
	m.fe.SetObserved("FEPods")
	m.fe.SetVar(0, "GMV")
	m.fe.SetVar(1, "Users")
	m.fe.SetVar(2, "MarketingCost")

	for _, d := range rows {
		fe, _, _ := d.Pods()
		m.fe.Train(regression.DataPoint(float64(fe), d.Features()))
	}
	require.NoError(t, m.fe.Run())

	coeffs := m.fe.GetCoeffs()
	require.Len(t, coeffs, 1+3) // [intercept, GMV, Users, MC]
	assert.InDelta(t, 1.0, coeffs[0], 1e-9)
	assert.InDelta(t, 2.0, coeffs[1], 1e-9)
	assert.InDelta(t, 3.0, coeffs[2], 1e-9)
	assert.InDelta(t, 4.0, coeffs[3], 1e-9)
}
