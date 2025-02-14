package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/par1ram/merch-store/internal/handlers"
	"github.com/par1ram/merch-store/internal/middleware"
	"github.com/par1ram/merch-store/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBuyHandler_HandleBuy_Success(t *testing.T) {
	// Создаем MockBuyService
	mockBuyService := new(service.MockBuyService)
	mockBuyService.On("Purchase", mock.Anything, "testItem").Return(nil)

	// Создаем BuyHandler
	buyHandler := handlers.NewBuyHandler(mockBuyService)

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "/api/buy/testItem", nil)
	ctx := context.WithValue(req.Context(), middleware.UserCtxKey, int64(123))
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	buyHandler.HandleBuy(w, req)

	// Проверяем ответ
	assert.Equal(t, http.StatusOK, w.Code)

	var responseMap map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &responseMap)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"status": "purchase successful"}, responseMap)

	// Проверяем, что MockBuyService был вызван
	mockBuyService.AssertExpectations(t)
}

func TestBuyHandler_HandleBuy_Error(t *testing.T) {
	// Создаем MockBuyService, который возвращает ошибку
	mockBuyService := new(service.MockBuyService)
	mockBuyService.On("Purchase", mock.Anything, "testItem").Return(errors.New("some error"))

	// Создаем BuyHandler
	buyHandler := handlers.NewBuyHandler(mockBuyService)

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "/api/buy/testItem", nil)
	ctx := context.WithValue(req.Context(), middleware.UserCtxKey, int64(123))
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	buyHandler.HandleBuy(w, req)

	// Проверяем ответ
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	// Проверьте тело ответа на наличие сообщения об ошибке

	// Проверяем, что MockBuyService был вызван
	mockBuyService.AssertExpectations(t)
}
