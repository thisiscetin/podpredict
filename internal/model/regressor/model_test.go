package regressor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thisiscetin/podpredict/internal/data/metrics"
	"github.com/thisiscetin/podpredict/internal/model"
)

// helper to create a Daily with pods
func makeDaily(gmv float64, users int, cost float64, fe, be int) metrics.Daily {
	d, _ := metrics.NewDaily(time.Now(), gmv, users, cost, &fe, &be)
	return d
}

func TestModel_TrainAndPredict(t *testing.T) {
	data := []metrics.Daily{
		makeDaily(1000, 10, 50, 1, 2),
		makeDaily(2000, 20, 80, 2, 3),
		makeDaily(3000, 30, 120, 3, 4),
	}

	m := NewModel()
	err := m.Train(data)
	assert.NoError(t, err)
	assert.True(t, m.trained)

	fePred, bePred, err := m.Predict(&model.Features{GMV: 1500, Users: 15, MarketingCost: 60})
	assert.NoError(t, err)
	assert.True(t, fePred > 0)
	assert.True(t, bePred > 0)
}

func TestModel_Predict_NotTrained(t *testing.T) {
	m := NewModel()
	_, _, err := m.Predict(&model.Features{GMV: 1000, Users: 10, MarketingCost: 50})
	assert.Error(t, err)
	assert.Equal(t, "model not trained", err.Error())
}

func TestModel_Train_EmptyData(t *testing.T) {
	m := NewModel()
	err := m.Train([]metrics.Daily{})
	assert.Error(t, err)
	assert.Equal(t, "no valid rows with FE/BE pods", err.Error())
}

func TestModel_PredictBatch(t *testing.T) {
	data := []metrics.Daily{
		makeDaily(1000, 10, 50, 1, 2),
		makeDaily(2000, 20, 80, 2, 3),
		makeDaily(3000, 30, 120, 3, 4),
	}

	m := NewModel()
	assert.NoError(t, m.Train(data))

	fePreds := make([]model.FEPods, len(data))
	bePreds := make([]model.BEPods, len(data))

	for i, d := range data {
		fe, be, err := m.Predict(&model.Features{GMV: d.GMV, Users: d.Users, MarketingCost: d.MarketingCost})
		assert.NoError(t, err)
		fePreds[i] = fe
		bePreds[i] = be
	}

	for i := range data {
		assert.True(t, fePreds[i] > 0)
		assert.True(t, bePreds[i] > 0)
	}
}

func TestModel_TransformData_MissingPods(t *testing.T) {
	dailyWithPods := makeDaily(1000, 10, 50, 1, 2)
	dailyWithoutPods := metrics.Daily{
		Date:          time.Now(),
		GMV:           1000,
		Users:         10,
		MarketingCost: 50,
		FEPods:        nil,
		BEPods:        nil,
	}

	X, feY, beY := transformData([]metrics.Daily{dailyWithPods, dailyWithoutPods})
	assert.Len(t, X, 1)
	assert.Len(t, feY, 1)
	assert.Len(t, beY, 1)
}
