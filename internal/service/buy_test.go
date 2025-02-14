// buy_service_test.go
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

// TestPurchase_Success проверяет успешную покупку.
func TestPurchase_Success(t *testing.T) {
	// Создаем контекст с валидными данными пользователя.
	claims := jwt.MapClaims{"user_id": 123.0}
	ctx := context.WithValue(context.Background(), middleware.UserCtxKey, claims)

	// Создаем мок репозитория.
	repoMock := new(repository.MockBuyRepository)

	// Тестовые данные товара.
	merchItem := "T-Shirt"
	merchData := db.Merch{
		ID:    1,
		Name:  merchItem,
		Price: 100,
	}

	// Ожидаем вызов GetMerch и возвращаем информацию о товаре.
	repoMock.On("GetMerch", ctx, merchItem).Return(merchData, nil).Once()

	// Ожидаем вызов GetBalance и возвращаем баланс, достаточный для покупки.
	repoMock.On("GetBalance", ctx, int32(123)).Return(int32(200), nil).Once()

	// В ExecTx внутри транзакции должны вызываться DeductCoins, UpsertInventory и CreatePurchaseTransaction.
	repoMock.On("ExecTx", ctx, mock.AnythingOfType("func(repository.BuyRepository) error")).
		Run(func(args mock.Arguments) {
			// Извлекаем функцию транзакции.
			txFunc := args.Get(1).(func(repository.BuyRepository) error)

			// Устанавливаем ожидания для методов, вызываемых внутри транзакции.
			repoMock.On("DeductCoins", ctx, int32(123), merchData.Price).Return(int64(1), nil).Once()
			repoMock.On("UpsertInventory", ctx, mock.Anything).Return(nil).Once()
			repoMock.On("CreatePurchaseTransaction", ctx, mock.Anything).Return(nil).Once()

			// Вызываем транзакционную функцию.
			err := txFunc(repoMock)
			assert.NoError(t, err)
		}).
		Return(nil).Once()

	// Используем тестовый логгер (реализация может быть простенькой, см. пример ниже).
	logger := utils.NewLogger()
	buySvc := service.NewBuyService(repoMock, logger)

	// Вызываем метод Purchase.
	err := buySvc.Purchase(ctx, merchItem)
	assert.NoError(t, err)

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

// TestPurchase_DeductCoinsFailure проверяет ошибку в транзакционной части (например, не удалось списать монеты).
func TestPurchase_DeductCoinsFailure(t *testing.T) {
	claims := jwt.MapClaims{"user_id": 123.0}
	ctx := context.WithValue(context.Background(), middleware.UserCtxKey, claims)

	repoMock := new(repository.MockBuyRepository)
	merchItem := "T-Shirt"
	merchData := db.Merch{
		ID:    1,
		Name:  merchItem,
		Price: 100,
	}
	repoMock.On("GetMerch", ctx, merchItem).Return(merchData, nil).Once()
	repoMock.On("GetBalance", ctx, int32(123)).Return(int32(200), nil).Once()

	deductErr := errors.New("failed to deduct coins")
	repoMock.On("ExecTx", ctx, mock.AnythingOfType("func(repository.BuyRepository) error")).
		Run(func(args mock.Arguments) {
			txFunc := args.Get(1).(func(repository.BuyRepository) error)
			repoMock.On("DeductCoins", ctx, int32(123), merchData.Price).Return(int64(0), deductErr).Once()
			err := txFunc(repoMock)
			assert.Equal(t, deductErr, err)
		}).
		Return(nil).Once()

	logger := utils.NewLogger()
	buySvc := service.NewBuyService(repoMock, logger)

	err := buySvc.Purchase(ctx, merchItem)
	assert.Error(t, err)
	assert.Equal(t, deductErr, err)

	repoMock.AssertExpectations(t)
}
