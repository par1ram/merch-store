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

type MockSendCoinRepository struct {
	mock.Mock
}

// ExecTx
func (m *MockSendCoinRepository) ExecTx(ctx context.Context, fn func(repository.SendCoinRepository) error) error {
	args := m.Called(ctx, fn)
	if err := args.Error(0); err != nil {
		return err
	}
	// Вызываем колбэк на том же самом объекте (m),
	// чтобы методы TransferCoins, GetBalance и т.п. вызывались на нем же.
	return fn(m)
}

// GetRecipient — возвращает *Employee или ошибку
func (m *MockSendCoinRepository) GetRecipient(ctx context.Context, username string) (*db.Employee, error) {
	args := m.Called(ctx, username)
	// Если args.Get(0) — nil, значит возвращаем (nil, err).
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	return val.(*db.Employee), args.Error(1)
}

// GetBalance — возвращает int32 (баланс) и ошибку
func (m *MockSendCoinRepository) GetBalance(ctx context.Context, userID int32) (int32, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int32), args.Error(1)
}

// TransferCoins — имитируем успешный или неуспешный перевод
func (m *MockSendCoinRepository) TransferCoins(ctx context.Context, fromUserID, toUserID, amount int32) error {
	args := m.Called(ctx, fromUserID, toUserID, amount)
	return args.Error(0)
}

// TestSendCoinService_Success — проверяем успешный сценарий перевода.
func TestSendCoinService_Success(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := new(MockSendCoinRepository)
	logger := utils.NewLogger()

	senderID := int64(123)
	toUser := "alice"
	amount := int32(50)

	// Контекст с user_id
	claims := jwt.MapClaims{"user_id": float64(senderID)}
	ctx := context.WithValue(context.Background(), middleware.UserCtxKey, claims)

	// Настраиваем ожидания
	// 1) ExecTx не возвращает ошибку (Return(nil)) — внутри вызов fn(m)
	mockRepo.On("ExecTx", ctx, mock.AnythingOfType("func(repository.SendCoinRepository) error")).
		Return(nil).Once()

	// 2) GetRecipient
	mockRepo.On("GetRecipient", mock.Anything, toUser).
		Return(&db.Employee{ID: 999, Username: "alice"}, nil).
		Once()

	// 3) GetBalance
	mockRepo.On("GetBalance", mock.Anything, int32(senderID)).
		Return(int32(200), nil).
		Once()

	// 4) TransferCoins
	mockRepo.On("TransferCoins", mock.Anything, int32(senderID), int32(999), amount).
		Return(nil).
		Once()

	svc := service.NewSendCoinService(mockRepo, logger)
	err := svc.SendCoin(ctx, toUser, amount)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// TestSendCoinService_AuthRequired — ошибка, если пользователь не аутентифицирован.
func TestSendCoinService_AuthRequired(t *testing.T) {
	mockRepo := new(MockSendCoinRepository)
	logger := utils.NewLogger()

	// Контекст без user_id
	ctx := context.Background()

	svc := service.NewSendCoinService(mockRepo, logger)
	err := svc.SendCoin(ctx, "alice", 50)
	assert.Error(t, err)
	assert.Equal(t, "authentication required", err.Error())

	// Проверим, что ExecTx и др. методы не вызывались
	mockRepo.AssertNotCalled(t, "ExecTx", mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "GetRecipient", mock.Anything, mock.Anything)
}

// TestSendCoinService_RecipientNotFound — если repo.GetRecipient вернёт ошибку,
// сервис должен вернуть бизнес-ошибку c ErrRecipientNotFound.
func TestSendCoinService_RecipientNotFound(t *testing.T) {
	mockRepo := new(MockSendCoinRepository)
	logger := utils.NewLogger()

	mockRepo.On("ExecTx", mock.Anything, mock.AnythingOfType("func(repository.SendCoinRepository) error")).
		Return(nil).Once()

	// GetRecipient вернет ошибку
	mockRepo.On("GetRecipient", mock.Anything, "bob").
		Return((*db.Employee)(nil), errors.New("no rows found")).Once()

	claims := jwt.MapClaims{"user_id": float64(111)}
	ctx := context.WithValue(context.Background(), middleware.UserCtxKey, claims)

	svc := service.NewSendCoinService(mockRepo, logger)
	err := svc.SendCoin(ctx, "bob", 30)
	assert.Error(t, err)

	// У нас код делает:
	//    return fmt.Errorf("%w: %v", ErrBusinessValidation, ErrRecipientNotFound)
	// Проверим, что сообщение содержит "recipient not found"
	assert.Contains(t, err.Error(), "recipient not found")

	mockRepo.AssertExpectations(t)
}

// TestSendCoinService_SelfTransfer — если recipient.ID == senderID, возвращаем ошибку.
func TestSendCoinService_SelfTransfer(t *testing.T) {
	mockRepo := new(MockSendCoinRepository)
	logger := utils.NewLogger()

	mockRepo.On("ExecTx", mock.Anything, mock.AnythingOfType("func(repository.SendCoinRepository) error")).
		Return(nil).Once()

	// GetRecipient возвращает Employee c таким же ID.
	mockRepo.On("GetRecipient", mock.Anything, "john").
		Return(&db.Employee{ID: 444}, nil).Once()

	claims := jwt.MapClaims{"user_id": float64(444)} // senderID = 444
	ctx := context.WithValue(context.Background(), middleware.UserCtxKey, claims)

	svc := service.NewSendCoinService(mockRepo, logger)
	err := svc.SendCoin(ctx, "john", 10)
	assert.Error(t, err)
	// Сервис выдаёт "...: self-transfer prohibited"
	assert.Contains(t, err.Error(), "self-transfer")

	mockRepo.AssertExpectations(t)
}

// TestSendCoinService_GetBalanceError — если repo.GetBalance вернёт ошибку,
// сервис возвращает "internal server error".
func TestSendCoinService_GetBalanceError(t *testing.T) {
	mockRepo := new(MockSendCoinRepository)
	logger := utils.NewLogger()

	mockRepo.On("ExecTx", mock.Anything, mock.AnythingOfType("func(repository.SendCoinRepository) error")).
		Return(nil).Once()

	mockRepo.On("GetRecipient", mock.Anything, "alice").
		Return(&db.Employee{ID: 999, Username: "alice"}, nil).Once()
	// GetBalance вернёт ошибку
	mockRepo.On("GetBalance", mock.Anything, int32(123)).
		Return(int32(0), errors.New("some DB error")).Once()

	claims := jwt.MapClaims{"user_id": float64(123)}
	ctx := context.WithValue(context.Background(), middleware.UserCtxKey, claims)

	svc := service.NewSendCoinService(mockRepo, logger)
	err := svc.SendCoin(ctx, "alice", 50)
	assert.Error(t, err)
	assert.Equal(t, "internal server error", err.Error())

	mockRepo.AssertExpectations(t)
}

// TestSendCoinService_InsufficientFunds — если balance < amount,
// сервис должен вернуть ErrInsufficientFunds в составе бизнес-ошибки.
func TestSendCoinService_InsufficientFunds(t *testing.T) {
	mockRepo := new(MockSendCoinRepository)
	logger := utils.NewLogger()

	mockRepo.On("ExecTx", mock.Anything, mock.AnythingOfType("func(repository.SendCoinRepository) error")).
		Return(nil).Once()

	mockRepo.On("GetRecipient", mock.Anything, "bob").
		Return(&db.Employee{ID: 999, Username: "bob"}, nil).
		Once()

	mockRepo.On("GetBalance", mock.Anything, int32(123)).
		Return(int32(40), nil).Once()

	claims := jwt.MapClaims{"user_id": float64(123)}
	ctx := context.WithValue(context.Background(), middleware.UserCtxKey, claims)

	svc := service.NewSendCoinService(mockRepo, logger)
	err := svc.SendCoin(ctx, "bob", 50)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds")

	mockRepo.AssertExpectations(t)
}

// TestSendCoinService_TransferError — если TransferCoins вернёт ошибку, сервис возвращает её.
func TestSendCoinService_TransferError(t *testing.T) {
	mockRepo := new(MockSendCoinRepository)
	logger := utils.NewLogger()

	mockRepo.On("ExecTx", mock.Anything, mock.AnythingOfType("func(repository.SendCoinRepository) error")).
		Return(nil).Once()

	mockRepo.On("GetRecipient", mock.Anything, "bob").
		Return(&db.Employee{ID: 999, Username: "bob"}, nil).Once()
	mockRepo.On("GetBalance", mock.Anything, int32(123)).
		Return(int32(100), nil).Once()

	mockRepo.On("TransferCoins", mock.Anything, int32(123), int32(999), int32(50)).
		Return(errors.New("updateEmployeeCoins failed")).
		Once()

	claims := jwt.MapClaims{"user_id": float64(123)}
	ctx := context.WithValue(context.Background(), middleware.UserCtxKey, claims)

	svc := service.NewSendCoinService(mockRepo, logger)
	err := svc.SendCoin(ctx, "bob", 50)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updateEmployeeCoins failed")

	mockRepo.AssertExpectations(t)
}
