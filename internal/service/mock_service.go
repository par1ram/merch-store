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

type MockInfoService struct {
	mock.Mock
}

func (m *MockInfoService) GetInfo(ctx context.Context, userID int64) (InfoResponse, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(InfoResponse), args.Error(1)
}

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Authenticate(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

type MockSendCoinService struct {
	mock.Mock
}

func (m *MockSendCoinService) SendCoin(ctx context.Context, toUser string, amount int32) error {
	args := m.Called(ctx, toUser, amount)
	return args.Error(0)
}
