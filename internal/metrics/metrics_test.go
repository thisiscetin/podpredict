package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDaily_Valid(t *testing.T) {
	d, err := NewDaily(time.Date(2025, 10, 20, 0, 0, 0, 0, time.UTC), 1000, 10, 500, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1000.0, d.GMV)
}

func TestNewDaily_InvalidDate(t *testing.T) {
	_, err := NewDaily(time.Time{}, 1000, 10, 500, nil, nil)
	assert.ErrorIs(t, err, ErrInvalidDate)
}

func TestNewDaily_NegativeGMV(t *testing.T) {
	_, err := NewDaily(time.Now(), -1, 10, 500, nil, nil)
	assert.ErrorIs(t, err, ErrGmvNegative)
}

func TestNewDaily_NegativeUsers(t *testing.T) {
	_, err := NewDaily(time.Now(), 1000, -1, 500, nil, nil)
	assert.ErrorIs(t, err, ErrUsersNegative)
}

func TestNewDaily_NegativeMarketing(t *testing.T) {
	_, err := NewDaily(time.Now(), 1000, 10, -5, nil, nil)
	assert.ErrorIs(t, err, ErrMarketingCostNegative)
}

func TestHasFePods(t *testing.T) {
	fe := 5
	d, _ := NewDaily(time.Now(), 1000, 10, 500, &fe, nil)
	assert.True(t, d.HasFePods())
}

func TestHasBePods(t *testing.T) {
	be := 5
	d, _ := NewDaily(time.Now(), 1000, 10, 500, nil, &be)
	assert.True(t, d.HasBePods())
}

func TestHasPods(t *testing.T) {
	fe := 1
	be := 2
	d, _ := NewDaily(time.Now(), 1000, 10, 500, &fe, &be)
	assert.True(t, d.HasPods())
}

func TestFeatures(t *testing.T) {
	d, _ := NewDaily(time.Now(), 1000, 10, 500, nil, nil)
	assert.Equal(t, []float64{1000, 10, 500}, d.Features())
}

func TestPods_Present(t *testing.T) {
	fe := 3
	be := 4
	d, _ := NewDaily(time.Now(), 1000, 10, 500, &fe, &be)
	f, b, ok := d.Pods()
	assert.True(t, ok)
	assert.Equal(t, 3, f)
	assert.Equal(t, 4, b)
}

func TestPods_Missing(t *testing.T) {
	d, _ := NewDaily(time.Now(), 1000, 10, 500, nil, nil)
	_, _, ok := d.Pods()
	assert.False(t, ok)
}
