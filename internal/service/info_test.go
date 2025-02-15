package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/par1ram/merch-store/internal/repository"
	"github.com/par1ram/merch-store/internal/service"
	"github.com/par1ram/merch-store/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockInfoRepository struct {
	mock.Mock
}

func (m *MockInfoRepository) GetCoins(ctx context.Context, userID int64) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockInfoRepository) GetInventory(ctx context.Context, userID int64) ([]repository.InventoryItem, error) {
	args := m.Called(ctx, userID)
	// Чтобы безопасно получить результат в виде []InventoryItem, используем args.Get(0) и делаем приведение
	if res, ok := args.Get(0).([]repository.InventoryItem); ok {
		return res, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInfoRepository) GetReceivedTransfers(ctx context.Context, userID int64) ([]repository.ReceivedTransaction, error) {
	args := m.Called(ctx, userID)
	if res, ok := args.Get(0).([]repository.ReceivedTransaction); ok {
		return res, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInfoRepository) GetSentTransfers(ctx context.Context, userID int64) ([]repository.SentTransaction, error) {
	args := m.Called(ctx, userID)
	if res, ok := args.Get(0).([]repository.SentTransaction); ok {
		return res, args.Error(1)
	}
	return nil, args.Error(1)
}

// TestInfoService_GetInfo_Success проверяет успешное получение информации.
func TestInfoService_GetInfo_Success(t *testing.T) {
	// Создаем мок-репозиторий.
	mockRepo := new(MockInfoRepository)

	userID := int64(123)

	// Настраиваем ответы для каждого метода, который вызовет сервис.
	mockRepo.On("GetCoins", mock.Anything, userID).Return(150, nil).Once()
	mockRepo.On("GetInventory", mock.Anything, userID).
		Return([]repository.InventoryItem{
			{Type: "T-Shirt", Quantity: 2},
			{Type: "Hoodie", Quantity: 1},
		}, nil).Once()
	mockRepo.On("GetReceivedTransfers", mock.Anything, userID).
		Return([]repository.ReceivedTransaction{
			{FromUser: "Alice", Amount: 50},
		}, nil).Once()
	mockRepo.On("GetSentTransfers", mock.Anything, userID).
		Return([]repository.SentTransaction{
			{ToUser: "Bob", Amount: 30},
		}, nil).Once()

	// Создаем InfoService с моковым репо.
	logger := utils.NewLogger() // или utils.NewLogger(), если у вас есть такая
	infoSvc := service.NewInfoService(mockRepo, logger)

	// Вызываем тестируемый метод
	resp, err := infoSvc.GetInfo(context.Background(), userID)

	// Проверяем, что нет ошибки
	assert.NoError(t, err)

	// Проверяем, что респонс сформирован правильно
	assert.Equal(t, 150, resp.Coins)
	assert.Len(t, resp.Inventory, 2)
	assert.Equal(t, "T-Shirt", resp.Inventory[0].Type)
	assert.Equal(t, 2, resp.Inventory[0].Quantity)
	assert.Equal(t, "Hoodie", resp.Inventory[1].Type)
	assert.Equal(t, 1, resp.Inventory[1].Quantity)

	assert.Len(t, resp.CoinHistory.Received, 1)
	assert.Equal(t, "Alice", resp.CoinHistory.Received[0].FromUser)
	assert.Equal(t, 50, resp.CoinHistory.Received[0].Amount)

	assert.Len(t, resp.CoinHistory.Sent, 1)
	assert.Equal(t, "Bob", resp.CoinHistory.Sent[0].ToUser)
	assert.Equal(t, 30, resp.CoinHistory.Sent[0].Amount)

	// Проверяем, что все ожидания выполнились
	mockRepo.AssertExpectations(t)
}

// TestInfoService_GetInfo_ErrorOnCoins проверяет ошибку на этапе GetCoins.
func TestInfoService_GetInfo_ErrorOnCoins(t *testing.T) {
	mockRepo := new(MockInfoRepository)
	userID := int64(123)

	mockRepo.On("GetCoins", mock.Anything, userID).Return(0, errors.New("coin error")).Once()

	// Сервис не должен вызывать остальные методы, если GetCoins вернет ошибку.
	// Поэтому не настраиваем GetInventory, GetReceivedTransfers, GetSentTransfers.

	infoSvc := service.NewInfoService(mockRepo, utils.NewLogger())

	resp, err := infoSvc.GetInfo(context.Background(), userID)
	assert.Error(t, err)
	assert.Equal(t, "coin error", err.Error())
	assert.Equal(t, service.InfoResponse{}, resp) // Пустой результат

	mockRepo.AssertExpectations(t)
}

// TestInfoService_GetInfo_ErrorOnInventory проверяет ошибку на этапе GetInventory.
func TestInfoService_GetInfo_ErrorOnInventory(t *testing.T) {
	mockRepo := new(MockInfoRepository)
	userID := int64(123)

	mockRepo.On("GetCoins", mock.Anything, userID).Return(150, nil).Once()
	mockRepo.On("GetInventory", mock.Anything, userID).Return(nil, errors.New("inventory error")).Once()

	// Остальные методы не должны быть вызваны после ошибки
	infoSvc := service.NewInfoService(mockRepo, utils.NewLogger())

	resp, err := infoSvc.GetInfo(context.Background(), userID)
	assert.Error(t, err)
	assert.Equal(t, "inventory error", err.Error())
	assert.Equal(t, service.InfoResponse{}, resp)

	mockRepo.AssertExpectations(t)
}

// TestInfoService_GetInfo_ErrorOnReceivedTransfers проверяет ошибку на этапе GetReceivedTransfers.
func TestInfoService_GetInfo_ErrorOnReceivedTransfers(t *testing.T) {
	mockRepo := new(MockInfoRepository)
	userID := int64(123)

	mockRepo.On("GetCoins", mock.Anything, userID).Return(150, nil).Once()
	mockRepo.On("GetInventory", mock.Anything, userID).
		Return([]repository.InventoryItem{
			{Type: "T-Shirt", Quantity: 2},
		}, nil).Once()
	mockRepo.On("GetReceivedTransfers", mock.Anything, userID).
		Return(nil, errors.New("received error")).Once()

	// Метод GetSentTransfers не вызывается, так как код должен вернуть ошибку после GetReceivedTransfers.
	infoSvc := service.NewInfoService(mockRepo, utils.NewLogger())

	resp, err := infoSvc.GetInfo(context.Background(), userID)
	assert.Error(t, err)
	assert.Equal(t, "received error", err.Error())
	assert.Equal(t, service.InfoResponse{}, resp)

	mockRepo.AssertExpectations(t)
}

// TestInfoService_GetInfo_ErrorOnSentTransfers проверяет ошибку на этапе GetSentTransfers.
func TestInfoService_GetInfo_ErrorOnSentTransfers(t *testing.T) {
	mockRepo := new(MockInfoRepository)
	userID := int64(123)

	mockRepo.On("GetCoins", mock.Anything, userID).Return(150, nil).Once()
	mockRepo.On("GetInventory", mock.Anything, userID).
		Return([]repository.InventoryItem{
			{Type: "T-Shirt", Quantity: 2},
		}, nil).Once()
	mockRepo.On("GetReceivedTransfers", mock.Anything, userID).
		Return([]repository.ReceivedTransaction{
			{FromUser: "Alice", Amount: 50},
		}, nil).Once()
	mockRepo.On("GetSentTransfers", mock.Anything, userID).
		Return(nil, errors.New("sent error")).Once()

	infoSvc := service.NewInfoService(mockRepo, utils.NewLogger())

	resp, err := infoSvc.GetInfo(context.Background(), userID)
	assert.Error(t, err)
	assert.Equal(t, "sent error", err.Error())
	assert.Equal(t, service.InfoResponse{}, resp)

	mockRepo.AssertExpectations(t)
}
