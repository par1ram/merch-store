package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/par1ram/merch-store/internal/service"

	"github.com/sirupsen/logrus"
)

// AuthHandler обрабатывает запросы на аутентификацию.
type AuthHandler struct {
	AuthService service.AuthService
}

// NewAuthHandler создаёт новый AuthHandler.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

// AuthRequest – структура входящего запроса.
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse – структура ответа с JWT-токеном.
type AuthResponse struct {
	Token string `json:"token"`
}

// HandleAuth обрабатывает POST-запрос на аутентификацию.
func (h *AuthHandler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	// Вызываем сервис для аутентификации.
	token, err := h.AuthService.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	resp := AuthResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logrus.WithError(err).Error("failed to encode auth response")
	}
}
