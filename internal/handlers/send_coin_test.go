package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/par1ram/merch-store/internal/handlers"
	"github.com/par1ram/merch-store/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendCoinHandler_HandleSendCoin_Success(t *testing.T) {
	// Формируем корректный запрос.
	reqBody := handlers.SendCoinRequest{
		ToUser: "Bob",
		Amount: 50,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/send-coin", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Настраиваем моковый сервис: при вызове SendCoin с аргументами ("Bob", 50) возвращаем nil.
	mockService := new(service.MockSendCoinService)
	mockService.
		On("SendCoin", mock.Anything, "Bob", int32(50)).
		Return(nil).
		Once()

	handler := handlers.NewSendCoinHandler(mockService)
	handler.HandleSendCoin(rr, req)

	// Проверяем статус ответа.
	assert.Equal(t, http.StatusOK, rr.Code)

	// Проверяем тело ответа: ожидается JSON {"status": "success"}
	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	expectedResp := map[string]string{"status": "success"}
	assert.Equal(t, expectedResp, resp)

	mockService.AssertExpectations(t)
}

func TestSendCoinHandler_HandleSendCoin_InvalidJSON(t *testing.T) {
	// Передаем некорректный JSON.
	req := httptest.NewRequest("POST", "/api/send-coin", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mockService := new(service.MockSendCoinService)
	// В этом сценарии метод SendCoin не должен вызываться.
	handler := handlers.NewSendCoinHandler(mockService)
	handler.HandleSendCoin(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	// Предполагаем, что utils.JSONErrorResponse формирует JSON с ключом "error".
	assert.Equal(t, "invalid request body", resp["error"])
}

func TestSendCoinHandler_HandleSendCoin_MissingToUser(t *testing.T) {
	// to_user отсутствует.
	reqBody := handlers.SendCoinRequest{
		ToUser: "",
		Amount: 50,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/send-coin", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mockService := new(service.MockSendCoinService)
	handler := handlers.NewSendCoinHandler(mockService)
	handler.HandleSendCoin(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "to_user is required", resp["error"])
}

func TestSendCoinHandler_HandleSendCoin_NonPositiveAmount(t *testing.T) {
	// amount <= 0.
	reqBody := handlers.SendCoinRequest{
		ToUser: "Bob",
		Amount: 0,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/send-coin", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mockService := new(service.MockSendCoinService)
	handler := handlers.NewSendCoinHandler(mockService)
	handler.HandleSendCoin(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "amount must be positive", resp["error"])
}

func TestSendCoinHandler_HandleSendCoin_BusinessValidationError(t *testing.T) {
	// Создаем корректный запрос.
	reqBody := handlers.SendCoinRequest{
		ToUser: "Bob",
		Amount: 50,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/send-coin", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Настраиваем моковый сервис так, чтобы он возвращал бизнес-валидационную ошибку.
	mockService := new(service.MockSendCoinService)
	mockService.
		On("SendCoin", mock.Anything, "Bob", int32(50)).
		Return(service.ErrBusinessValidation).
		Once()

	handler := handlers.NewSendCoinHandler(mockService)
	handler.HandleSendCoin(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	// Ожидаем, что текст ошибки совпадает с service.ErrBusinessValidation.Error()
	assert.Equal(t, service.ErrBusinessValidation.Error(), resp["error"])

	mockService.AssertExpectations(t)
}

func TestSendCoinHandler_HandleSendCoin_InternalError(t *testing.T) {
	// Создаем корректный запрос.
	reqBody := handlers.SendCoinRequest{
		ToUser: "Bob",
		Amount: 50,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/send-coin", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Настраиваем моковый сервис так, чтобы он возвращал общую ошибку.
	mockService := new(service.MockSendCoinService)
	mockService.
		On("SendCoin", mock.Anything, "Bob", int32(50)).
		Return(errors.New("some internal error")).
		Once()

	handler := handlers.NewSendCoinHandler(mockService)
	handler.HandleSendCoin(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "internal error", resp["error"])

	mockService.AssertExpectations(t)
}
