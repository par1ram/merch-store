package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/par1ram/merch-store/internal/handlers"
	"github.com/par1ram/merch-store/internal/middleware"
	"github.com/par1ram/merch-store/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInfoHandler_HandleInfo_Success(t *testing.T) {
	// Создаем мок-сервис InfoService.
	mockInfoService := new(service.MockInfoService)

	// Ожидаемые данные, которые должен вернуть сервис.
	expectedInfo := service.InfoResponse{
		Coins: 100,
		Inventory: []service.Inventory{
			{Type: "T-Shirt", Quantity: 2},
		},
		CoinHistory: service.CoinHistory{
			Received: []service.ReceivedTransaction{
				{FromUser: "Alice", Amount: 50},
			},
			Sent: []service.SentTransaction{
				{ToUser: "Bob", Amount: 30},
			},
		},
	}

	// Настраиваем ожидание: при вызове GetInfo с любым контекстом и userID равным 123 возвращаются expectedInfo и nil.
	mockInfoService.
		On("GetInfo", mock.Anything, int64(123)).
		Return(expectedInfo, nil).
		Once()

	// Создаем экземпляр обработчика InfoHandler с моковым сервисом.
	infoHandler := handlers.NewInfoHandler(mockInfoService)

	// Создаем тестовый HTTP-запрос к /api/info.
	req := httptest.NewRequest("GET", "/api/info", nil)
	// В контекст помещаем jwt.MapClaims, где ключ "user_id" имеет значение 123 (как float64).
	claims := jwt.MapClaims{"user_id": 123.0}
	ctx := context.WithValue(req.Context(), middleware.UserCtxKey, claims)
	req = req.WithContext(ctx)

	// Создаем ResponseRecorder для захвата ответа.
	w := httptest.NewRecorder()

	// Вызываем обработчик.
	infoHandler.HandleInfo(w, req)

	// Проверяем, что статус ответа — 200 OK.
	assert.Equal(t, http.StatusOK, w.Code)

	// Разбираем JSON-ответ.
	var response service.InfoResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	// Сравниваем полученный ответ с ожидаемыми данными.
	assert.Equal(t, expectedInfo, response)

	// Проверяем, что все ожидания мока выполнены.
	mockInfoService.AssertExpectations(t)
}

func TestInfoHandler_HandleInfo_Error(t *testing.T) {
	// Моковый сервис, возвращающий ошибку.
	mockInfoService := new(service.MockInfoService)
	mockInfoService.
		On("GetInfo", mock.Anything, int64(123)).
		Return(service.InfoResponse{}, errors.New("service error")).
		Once()

	infoHandler := handlers.NewInfoHandler(mockInfoService)

	req := httptest.NewRequest("GET", "/api/info", nil)
	claims := jwt.MapClaims{"user_id": 123.0}
	ctx := context.WithValue(req.Context(), middleware.UserCtxKey, claims)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	infoHandler.HandleInfo(w, req)

	// Ожидаем статус 500 Internal Server Error.
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Проверяем, что тело ответа содержит сообщение об ошибке.
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "failed to retrieve info", resp["error"])

	mockInfoService.AssertExpectations(t)
}

func TestInfoHandler_HandleInfo_Unauthorized(t *testing.T) {
	// Если в контексте отсутствует userID, обработчик должен вернуть 401.
	mockInfoService := new(service.MockInfoService)
	// В этом сценарии метод GetInfo не должен вызываться.
	infoHandler := handlers.NewInfoHandler(mockInfoService)

	// Создаем запрос без установки идентификатора пользователя в контекст.
	req := httptest.NewRequest("GET", "/api/info", nil)
	w := httptest.NewRecorder()

	infoHandler.HandleInfo(w, req)

	// Ожидаем статус 401 Unauthorized.
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "user not authenticated", resp["error"])
}
