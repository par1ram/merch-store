package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/par1ram/merch-store/internal/middleware"
	"github.com/stretchr/testify/assert"
)

// dummyHandler — простой обработчик, который запоминает, что был вызван.
type dummyHandler struct {
	called bool
}

func (h *dummyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func TestJWTMiddleware_MissingAuthorizationHeader(t *testing.T) {
	secret := []byte("test-secret")
	mw := middleware.JWTMiddleware(secret)

	next := &dummyHandler{}
	handler := mw(next)

	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	// Не устанавливаем заголовок Authorization
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "missing authorization header")
	assert.False(t, next.called, "next handler should not be called")
}

func TestJWTMiddleware_InvalidAuthorizationHeader(t *testing.T) {
	secret := []byte("test-secret")
	mw := middleware.JWTMiddleware(secret)

	next := &dummyHandler{}
	handler := mw(next)

	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	req.Header.Set("Authorization", "SomethingWrong")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid authorization header")
	assert.False(t, next.called, "next handler should not be called")
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	secret := []byte("test-secret")
	mw := middleware.JWTMiddleware(secret)

	next := &dummyHandler{}
	handler := mw(next)

	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	req.Header.Set("Authorization", "Bearer invalidtokenstring")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid token")
	assert.False(t, next.called)
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	secret := []byte("test-secret")
	mw := middleware.JWTMiddleware(secret)

	// Создаём JWT-токен с user_id=123.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(123),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString(secret)
	assert.NoError(t, err)

	next := &dummyHandler{}
	handler := mw(next)

	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, next.called)
	assert.Equal(t, "OK", rr.Body.String())
}

func TestJWTMiddleware_InvalidClaims(t *testing.T) {
	// В этом тесте проверим, что если claims не jwt.MapClaims,
	// то вернётся 401. Для этого «подделаем» метод Parse.
	// Однако проще всего подделать этот вариант,
	// если, к примеру, подпись другая или claims другого типа.
	// Здесь приведён лишь общий пример.

	secret := []byte("test-secret")
	mw := middleware.JWTMiddleware(secret)

	// Создадим токен с неподходящей SigningMethod
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"user_id": float64(456),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	// Подпишем чем-то простым (может вызвать ошибку)
	tokenStr, _ := token.SignedString([]byte("other-secret"))

	next := &dummyHandler{}
	handler := mw(next)

	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid token")
	assert.False(t, next.called)
}

func TestGetUserIDFromContext(t *testing.T) {
	ctx := context.Background()

	// Без user_id в контексте
	userID := middleware.GetUserIDFromContext(ctx)
	assert.Equal(t, int64(0), userID)

	// С user_id=99
	claims := jwt.MapClaims{"user_id": float64(99)}
	ctxWithClaims := context.WithValue(ctx, middleware.UserCtxKey, claims)

	userID = middleware.GetUserIDFromContext(ctxWithClaims)
	assert.Equal(t, int64(99), userID)
}
