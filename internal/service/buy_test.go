package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/par1ram/merch-store/internal/db"
	"github.com/par1ram/merch-store/internal/middleware"
	"github.com/par1ram/merch-store/internal/repository"
	"github.com/par1ram/merch-store/internal/service"
	"github.com/par1ram/merch-store/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPurchase_Success(t *testing.T) {
	// 1) Создаём контекст с user_id
	claims := jwt.MapClaims{"user_id": 123.0}
	ctx := context.WithValue(context.Background(), middleware.UserCtxKey, claims)

	// 2) Создаём мок
	repoMock := new(repository.MockBuyRepository)

	// 3) Настраиваем «вне транзакции»
	merchData := db.Merch{
		ID:    int32(1),
		Name:  "T-Shirt",
		Price: int32(100),
	}
	repoMock.On("GetMerch", ctx, "T-Shirt").Return(merchData, nil).Once()
	repoMock.On("GetBalance", ctx, int32(123)).Return(int32(200), nil).Once()

	// 4) Настраиваем ExecTx (не делаем Run(func(...){...}) — достаточно Return(nil)),
	//    потому что в MockBuyRepository.ExecTx уже есть “return fn(m)”.
	repoMock.On("ExecTx", ctx, mock.AnythingOfType("func(repository.BuyRepository) error")).
		Return(nil).Once()

	// 5) Настраиваем методы, которые будут вызваны внутри транзакции (на том же repoMock!):
	repoMock.On("DeductCoins", mock.Anything, int32(123), int32(100)).
		Return(int64(1), nil).
		Once()
	repoMock.On("UpsertInventory", mock.Anything, mock.Anything).
		Return(nil).
		Once()
	repoMock.On("CreatePurchaseTransaction", mock.Anything, mock.Anything).
		Return(nil).
		Once()

	// 6) Вызываем сервис
	logger := utils.NewLogger()
	buySvc := service.NewBuyService(repoMock, logger)
	err := buySvc.Purchase(ctx, "T-Shirt")
	assert.NoError(t, err)

	// 7) Проверяем ожидания
	repoMock.AssertExpectations(t)
}

// TestPurchase_UserNotAuthenticated проверяет случай, когда пользователь не аутентифицирован.
func TestPurchase_UserNotAuthenticated(t *testing.T) {
	// Контекст без данных пользователя.
	ctx := context.Background()

	repoMock := new(repository.MockBuyRepository)
	logger := utils.NewLogger()
	buySvc := service.NewBuyService(repoMock, logger)

	err := buySvc.Purchase(ctx, "T-Shirt")
	assert.Error(t, err)
	assert.Equal(t, "user not authenticated", err.Error())
}

// TestPurchase_InsufficientFunds проверяет, что при недостатке средств возвращается соответствующая ошибка.
func TestPurchase_InsufficientFunds(t *testing.T) {
	claims := jwt.MapClaims{"user_id": 123.0}
	ctx := context.WithValue(context.Background(), middleware.UserCtxKey, claims)

	repoMock := new(repository.MockBuyRepository)
	merchItem := "T-Shirt"
	merchData := db.Merch{
		ID:    1,
		Name:  merchItem,
		Price: 150,
	}
	repoMock.On("GetMerch", ctx, merchItem).Return(merchData, nil).Once()
	// Возвращаем баланс меньше цены товара.
	repoMock.On("GetBalance", ctx, int32(123)).Return(int32(100), nil).Once()

	logger := utils.NewLogger()
	buySvc := service.NewBuyService(repoMock, logger)

	err := buySvc.Purchase(ctx, merchItem)
	assert.Error(t, err)
	assert.Equal(t, "insufficient funds", err.Error())

	repoMock.AssertExpectations(t)
}

// TestPurchase_GetMerchError проверяет ошибку при получении товара.
func TestPurchase_GetMerchError(t *testing.T) {
	claims := jwt.MapClaims{"user_id": 123.0}
	ctx := context.WithValue(context.Background(), middleware.UserCtxKey, claims)

	repoMock := new(repository.MockBuyRepository)
	merchItem := "NonExistentItem"
	getMerchErr := errors.New("merch not found")
	repoMock.On("GetMerch", ctx, merchItem).Return(db.Merch{}, getMerchErr).Once()

	logger := utils.NewLogger()
	buySvc := service.NewBuyService(repoMock, logger)

	err := buySvc.Purchase(ctx, merchItem)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merch not found")

	repoMock.AssertExpectations(t)
}

func TestPurchase_DeductCoinsFailure(t *testing.T) {
	claims := jwt.MapClaims{"user_id": 123.0}
	ctx := context.WithValue(context.Background(), middleware.UserCtxKey, claims)

	repoMock := new(repository.MockBuyRepository)
	merchItem := "T-Shirt"
	merchData := db.Merch{
		ID:    int32(1),
		Name:  merchItem,
		Price: int32(100),
	}

	// "До транзакции"
	repoMock.On("GetMerch", ctx, merchItem).Return(merchData, nil).Once()
	repoMock.On("GetBalance", ctx, int32(123)).Return(int32(200), nil).Once()

	// ExecTx
	repoMock.On("ExecTx", ctx, mock.AnythingOfType("func(repository.BuyRepository) error")).
		Return(nil).Once()

	// "Внутри транзакции"
	// Возвращаем (int64(0), deductErr)
	deductErr := errors.New("failed to deduct coins")
	repoMock.On("DeductCoins", mock.Anything, int32(123), int32(100)).
		Return(int64(0), deductErr).
		Once()

	buySvc := service.NewBuyService(repoMock, utils.NewLogger())
	err := buySvc.Purchase(ctx, merchItem)
	assert.Error(t, err)
	assert.Equal(t, deductErr, err)

	repoMock.AssertExpectations(t)
}
