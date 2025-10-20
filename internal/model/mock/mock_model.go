package mock

import "github.com/stretchr/testify/mock"

// MockModel is a mock implementation of the Model interface.
type MockModel struct {
	mock.Mock
}

// Predict mocks the Predict method.
func (m *MockModel) Predict(input []float64) (int, int) {
	args := m.Called(input)

	return args.Int(0), args.Int(1)
}
