package mock

import (
	"github.com/stretchr/testify/mock"
	"github.com/thisiscetin/podpredict/internal/data/metrics"
)

// MockFetcher is a mock implementation of the Fetcher interface.
type MockFetcher struct {
	mock.Mock
}

// Fetch mocks the Fetch method to return predefined data.
func (m *MockFetcher) Fetch() ([]metrics.Daily, error) {
	args := m.Called()
	return args.Get(0).([]metrics.Daily), args.Error(1)
}
