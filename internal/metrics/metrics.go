package metrics

import (
	"errors"
	"time"
)

var (
	// ErrInvalidDate is returned when a zero date is provided.
	ErrInvalidDate = errors.New("invalid date")
	// ErrGmvNegative is returned when GMV is negative.
	ErrGmvNegative = errors.New("gmv cannot be negative")
	// ErrUsersNegative is returned when Users is negative.
	ErrUsersNegative = errors.New("users cannot be negative")
	// ErrMarketingCostNegative is returned when MarketingCost is negative.
	ErrMarketingCostNegative = errors.New("marketing cost cannot be negative")
)

// Daily represents the business KPIs and optional pod counts for a single day.
// This will be used an input to models.
type Daily struct {
	Date          time.Time
	GMV           float64
	Users         float64
	MarketingCost float64
	FEPods        *int
	BEPods        *int
}

// NewDaily creates a new Daily instance after validating input values.
// Returns an error if any validation fails.
func NewDaily(date time.Time, gmv float64, users int, marketingCost float64, fePods, bePods *int) (Daily, error) {
	if date.IsZero() {
		return Daily{}, ErrInvalidDate
	}
	if gmv < 0 {
		return Daily{}, ErrGmvNegative
	}
	if users < 0 {
		return Daily{}, ErrUsersNegative
	}
	if marketingCost < 0 {
		return Daily{}, ErrMarketingCostNegative
	}
	return Daily{
		Date:          date,
		GMV:           gmv,
		Users:         float64(users),
		MarketingCost: marketingCost,
		FEPods:        fePods,
		BEPods:        bePods,
	}, nil
}

// HasFePods returns true if FEPods is not nil.
func (d Daily) HasFePods() bool { return d.FEPods != nil }

// HasBePods returns true if BEPods is not nil.
func (d Daily) HasBePods() bool { return d.BEPods != nil }

// HasPods returns true if both FEPods and BEPods are present.
func (d Daily) HasPods() bool { return d.HasFePods() && d.HasBePods() }

// Features returns the numeric feature slice for ML models.
func (d Daily) Features() []float64 {
	return []float64{
		d.GMV,
		d.Users,
		d.MarketingCost,
	}
}

// Pods returns FE and BE pods values along with a boolean indicating if both are present.
func (d Daily) Pods() (fe int, be int, ok bool) {
	if d.FEPods == nil || d.BEPods == nil {
		return 0, 0, false
	}
	return *d.FEPods, *d.BEPods, true
}
