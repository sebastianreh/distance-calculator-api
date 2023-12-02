package mocks

import (
	"context"
	"github.com/sebastianreh/distance-calculator-api/internal/entities"
	"github.com/stretchr/testify/mock"
)

type CalculatorServiceMock struct {
	mock.Mock
}

func NewCalculatorServiceMock() *CalculatorServiceMock {
	return new(CalculatorServiceMock)
}

func (m *CalculatorServiceMock) CalculateDeliveryRange(ctx context.Context,
	request entities.CalculationRequest) (entities.CalculationResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(entities.CalculationResponse), args.Error(1)
}

func (m *CalculatorServiceMock) PreprocessRestaurants(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
