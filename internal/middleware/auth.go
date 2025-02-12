package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const UserCtxKey = contextKey("user")

// JWTMiddleware проверяет валидность JWT-токена и добавляет данные из него в контекст.
func JWTMiddleware(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}
			tokenStr := parts[1]

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return jwtSecret, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				ctx := context.WithValue(r.Context(), UserCtxKey, claims)
				r = r.WithContext(ctx)
			} else {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext — вспомогательная функция для извлечения userID из контекста.
func GetUserIDFromContext(ctx context.Context) int64 {
	claims, ok := ctx.Value(UserCtxKey).(jwt.MapClaims)
	if !ok {
		return 0
	}
	if id, ok := claims["user_id"].(float64); ok {
		return int64(id)
	}
	return 0
}
