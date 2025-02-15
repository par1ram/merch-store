package e2e_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:8080"

func TestE2EScenario(t *testing.T) {
	client := &http.Client{}

	// Шаг 1: Аутентификация (POST /api/auth)
	authURL := baseURL + "/api/auth"
	authPayload := map[string]string{
		"username": "testuser",
		"password": "testpassword",
	}
	authBody, err := json.Marshal(authPayload)
	assert.NoError(t, err, "Ошибка маршалинга запроса аутентификации")

	authResp, err := client.Post(authURL, "application/json", bytes.NewBuffer(authBody))
	assert.NoError(t, err, "Ошибка выполнения запроса аутентификации")
	assert.Equal(t, http.StatusOK, authResp.StatusCode, "Аутентификация должна вернуть 200")

	var authData struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(authResp.Body).Decode(&authData)
	authResp.Body.Close()
	assert.NoError(t, err, "Ошибка декодирования ответа аутентификации")
	assert.NotEmpty(t, authData.Token, "JWT-токен не должен быть пустым")
	token := authData.Token

	// Шаг 2: Покупка мерча (GET /api/buy/T-Shirt)
	buyURL := baseURL + "/api/buy/t-shirt"
	reqBuy, err := http.NewRequest(http.MethodGet, buyURL, nil)
	assert.NoError(t, err, "Ошибка создания запроса на покупку мерча")
	reqBuy.Header.Set("Authorization", "Bearer "+token)
	buyResp, err := client.Do(reqBuy)
	assert.NoError(t, err, "Ошибка выполнения запроса на покупку мерча")
	assert.Equal(t, http.StatusOK, buyResp.StatusCode, "Покупка мерча должна вернуть 200")
	buyResp.Body.Close()

	// Шаг 3: Передача монет (POST /api/send-coin)
	sendCoinURL := baseURL + "/api/send-coin"
	sendCoinPayload := map[string]interface{}{
		"to_user": "alice",
		"amount":  50,
	}
	sendCoinBody, err := json.Marshal(sendCoinPayload)
	assert.NoError(t, err, "Ошибка маршалинга запроса передачи монет")

	reqSend, err := http.NewRequest(http.MethodPost, sendCoinURL, bytes.NewBuffer(sendCoinBody))
	assert.NoError(t, err, "Ошибка создания запроса передачи монет")
	reqSend.Header.Set("Authorization", "Bearer "+token)
	reqSend.Header.Set("Content-Type", "application/json")
	sendResp, err := client.Do(reqSend)
	assert.NoError(t, err, "Ошибка выполнения запроса передачи монет")
	assert.Equal(t, http.StatusOK, sendResp.StatusCode, "Передача монет должна вернуть 200")
	sendResp.Body.Close()

	// Шаг 4: Получение информации (GET /api/info)
	infoURL := baseURL + "/api/info"
	reqInfo, err := http.NewRequest(http.MethodGet, infoURL, nil)
	assert.NoError(t, err, "Ошибка создания запроса получения информации")
	reqInfo.Header.Set("Authorization", "Bearer "+token)
	infoResp, err := client.Do(reqInfo)
	assert.NoError(t, err, "Ошибка выполнения запроса получения информации")
	assert.Equal(t, http.StatusOK, infoResp.StatusCode, "Запрос информации должен вернуть 200")

	var infoData map[string]interface{}
	err = json.NewDecoder(infoResp.Body).Decode(&infoData)
	infoResp.Body.Close()
	assert.NoError(t, err, "Ошибка декодирования ответа информации")
	// Проверяем наличие ключей согласно swagger-описанию.
	assert.Contains(t, infoData, "coins", "Ответ должен содержать поле 'coins'")
	assert.Contains(t, infoData, "inventory", "Ответ должен содержать поле 'inventory'")
	assert.Contains(t, infoData, "coinHistory", "Ответ должен содержать поле 'coinHistory'")
}
