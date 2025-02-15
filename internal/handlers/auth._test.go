package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/par1ram/merch-store/internal/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService — моковая реализация AuthService для тестов.
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Authenticate(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

func TestAuthHandler_HandleAuth_Success(t *testing.T) {
	// Подготавливаем корректный JSON-тело запроса.
	reqBody := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Настраиваем моковый сервис аутентификации.
	mockAuthService := new(MockAuthService)
	mockAuthService.
		On("Authenticate", mock.Anything, "testuser", "testpass").
		Return("valid_token", nil).
		Once()

	// Создаем AuthHandler с использованием мока.
	authHandler := handlers.NewAuthHandler(mockAuthService)

	// Вызываем обработчик.
	authHandler.HandleAuth(rr, req)

	// Проверяем, что статус ответа 200 OK.
	assert.Equal(t, http.StatusOK, rr.Code)

	// Разбираем JSON-ответ.
	var resp handlers.AuthResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "valid_token", resp.Token)

	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_HandleAuth_InvalidJSON(t *testing.T) {
	// Передаем некорректный JSON.
	req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// В этом тесте AuthService не должен вызываться.
	mockAuthService := new(MockAuthService)

	authHandler := handlers.NewAuthHandler(mockAuthService)
	authHandler.HandleAuth(rr, req)

	// Ожидаем статус 400 Bad Request.
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// Проверяем, что в теле ответа содержится "invalid request".
	assert.Contains(t, rr.Body.String(), "invalid request")
}

func TestAuthHandler_HandleAuth_MissingFields(t *testing.T) {
	// Создаем запрос с отсутствующими полями.
	reqBody := map[string]string{
		"username": "",
		"password": "",
	}
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// AuthService не должен вызываться, так как поля отсутствуют.
	mockAuthService := new(MockAuthService)

	authHandler := handlers.NewAuthHandler(mockAuthService)
	authHandler.HandleAuth(rr, req)

	// Ожидаем статус 400 Bad Request.
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// Проверяем, что в теле ответа содержится "username and password are required".
	assert.Contains(t, rr.Body.String(), "username and password are required")
}

func TestAuthHandler_HandleAuth_AuthError(t *testing.T) {
	// Создаем корректный JSON-тело запроса.
	reqBody := map[string]string{
		"username": "testuser",
		"password": "wrongpass",
	}
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Настраиваем моковый сервис так, чтобы он возвращал ошибку аутентификации.
	mockAuthService := new(MockAuthService)
	mockAuthService.
		On("Authenticate", mock.Anything, "testuser", "wrongpass").
		Return("", errors.New("authentication failed")).
		Once()

	authHandler := handlers.NewAuthHandler(mockAuthService)
	authHandler.HandleAuth(rr, req)

	// Ожидаем статус 401 Unauthorized.
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	// Проверяем, что в теле ответа содержится сообщение об ошибке.
	assert.Contains(t, rr.Body.String(), "authentication failed")

	mockAuthService.AssertExpectations(t)
}
