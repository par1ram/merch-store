package service

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockBuyService struct {
	mock.Mock
}

func (m *MockBuyService) Purchase(ctx context.Context, item string) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}
