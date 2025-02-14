package repository

import (
	"context"

	"github.com/par1ram/merch-store/internal/db"
	"github.com/stretchr/testify/mock"
)

type MockBuyRepository struct {
	mock.Mock
}

func (m *MockBuyRepository) ExecTx(ctx context.Context, fn func(BuyRepository) error) error {
	args := m.Called(ctx, fn)
	if err := args.Error(0); err != nil {
		return err
	}
	return fn(m)
}

func (m *MockBuyRepository) GetMerch(ctx context.Context, merchName string) (db.Merch, error) {
	args := m.Called(ctx, merchName)
	return args.Get(0).(db.Merch), args.Error(1)
}

func (m *MockBuyRepository) GetBalance(ctx context.Context, userID int32) (int32, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int32), args.Error(1)
}

func (m *MockBuyRepository) DeductCoins(ctx context.Context, userID, amount int32) (int64, error) {
	args := m.Called(ctx, userID, amount)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBuyRepository) UpsertInventory(ctx context.Context, params db.UpsertInventoryParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockBuyRepository) CreatePurchaseTransaction(ctx context.Context, params db.CreateCoinTransactionPurchaseParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}
