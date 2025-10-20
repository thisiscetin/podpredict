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

	fePods := model.FEPods(args.Int(0))
	bePods := model.BEPods(args.Int(1))
	err, _ := args.Get(2).(error) // Safely retrieve the error value (if any)

	return fePods, bePods, err
}
