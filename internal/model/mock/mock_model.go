package mock

import (
	"github.com/stretchr/testify/mock"
	"github.com/thisiscetin/podpredict/internal/model"
)

// MockModel is a mock implementation of the Model interface.
type MockModel struct {
	mock.Mock
}

// Predict mocks the Predict method.
func (m *MockModel) Predict(features *model.Features) (model.FEPods, model.BEPods, error) {
	args := m.Called(features)

	return model.FEPods(args.Int(0)), model.BEPods(args.Int(1)), nil
}
