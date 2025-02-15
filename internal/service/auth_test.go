package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/par1ram/merch-store/internal/repository"
	"github.com/par1ram/merch-store/internal/service"
	"github.com/par1ram/merch-store/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository — мок для репозитория пользователей.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (repository.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(repository.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, username, passwordHash string) (repository.User, error) {
	args := m.Called(ctx, username, passwordHash)
	return args.Get(0).(repository.User), args.Error(1)
}

// userStub — упрощённая заготовка для пользователя.
func userStub(id int32, username, passwordHash string) repository.User {
	return repository.User{
		ID:           int64(id),
		Username:     username,
		PasswordHash: passwordHash,
	}
}

func TestAuthService_UserNotFound_CreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("secret")
	logger := utils.NewLogger()
	authSvc := service.NewAuthService(mockRepo, jwtSecret, logger)

	ctx := context.Background()
	username := "newuser"
	password := "12345"

	// 1) GetByUsername возвращает sql.ErrNoRows => пользователь не найден
	mockRepo.
		On("GetByUsername", ctx, username).
		Return(repository.User{}, sql.ErrNoRows).
		Once()

	// 2) При создании пользователя возвращаем успешный результат
	mockRepo.
		On("Create", ctx, username, mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) {
			// Можно проверить, что "passwordHash" это bcrypt-хеш (но обычно строку не сравнивают напрямую).
		}).
		Return(userStub(10, username, "someHash"), nil).
		Once()

	token, err := authSvc.Authenticate(ctx, username, password)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	mockRepo.AssertExpectations(t)

	// Дополнительно можно декодировать токен и проверить claims.
	claims := extractClaims(t, token, jwtSecret)
	assert.Equal(t, float64(10), claims["user_id"])
	assert.Equal(t, username, claims["username"])
}

func TestAuthService_UserFound_CorrectPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("secret")
	logger := utils.NewLogger()
	authSvc := service.NewAuthService(mockRepo, jwtSecret, logger)

	ctx := context.Background()
	username := "existing"
	password := "mypassword"

	// Генерируем bcrypt-хеш, чтобы проверка CompareHashAndPassword прошла.
	hashBytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashStr := string(hashBytes)

	// 1) GetByUsername возвращает пользователя
	mockRepo.
		On("GetByUsername", ctx, username).
		Return(userStub(5, username, hashStr), nil).
		Once()

	token, err := authSvc.Authenticate(ctx, username, password)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	mockRepo.AssertExpectations(t)

	claims := extractClaims(t, token, jwtSecret)
	assert.Equal(t, float64(5), claims["user_id"])
	assert.Equal(t, username, claims["username"])
}

func TestAuthService_UserFound_InvalidPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("secret")
	logger := utils.NewLogger()
	authSvc := service.NewAuthService(mockRepo, jwtSecret, logger)

	ctx := context.Background()
	username := "existing"
	password := "wrongpass"

	// Создадим хеш "correctpass"
	correctPass := "correctpass"
	hashBytes, _ := bcrypt.GenerateFromPassword([]byte(correctPass), bcrypt.DefaultCost)
	hashStr := string(hashBytes)

	// GetByUsername возвращает пользователя с паролем correctPass
	mockRepo.
		On("GetByUsername", ctx, username).
		Return(userStub(5, username, hashStr), nil).
		Once()

	token, err := authSvc.Authenticate(ctx, username, password)
	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Empty(t, token)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_GetUserError(t *testing.T) {
	// Проверяем, что если GetByUsername возвращает не ErrNoRows, а другую ошибку,
	// сервис возвращает эту ошибку.
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("secret")
	logger := utils.NewLogger()
	authSvc := service.NewAuthService(mockRepo, jwtSecret, logger)

	ctx := context.Background()
	username := "someuser"
	password := "somepass"

	mockRepo.
		On("GetByUsername", ctx, username).
		Return(repository.User{}, errors.New("db error")).
		Once()

	token, err := authSvc.Authenticate(ctx, username, password)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	assert.Empty(t, token)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_CreateUserError(t *testing.T) {
	// Если пользователь не найден, но при создании возникает ошибка.
	mockRepo := new(MockUserRepository)
	jwtSecret := []byte("secret")
	logger := utils.NewLogger()
	authSvc := service.NewAuthService(mockRepo, jwtSecret, logger)

	ctx := context.Background()
	username := "newuser"
	password := "12345"

	// GetByUsername возвращает sql.ErrNoRows
	mockRepo.
		On("GetByUsername", ctx, username).
		Return(repository.User{}, sql.ErrNoRows).
		Once()

	// Create возвращает ошибку
	mockRepo.
		On("Create", ctx, username, mock.AnythingOfType("string")).
		Return(repository.User{}, errors.New("create failed")).
		Once()

	token, err := authSvc.Authenticate(ctx, username, password)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create failed")
	assert.Empty(t, token)

	mockRepo.AssertExpectations(t)
}

//
// Дополнительные вспомогательные функции
//

// extractClaims — для удобства проверяем содержимое JWT, используя jwtSecret.
func extractClaims(t *testing.T, tokenStr string, jwtSecret []byte) jwt.MapClaims {
	t.Helper()
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, token)
	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	return claims
}
