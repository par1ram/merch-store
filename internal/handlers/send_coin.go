package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/par1ram/merch-store/internal/service"
	"github.com/par1ram/merch-store/internal/utils"
)

type SendCoinRequest struct {
	ToUser string `json:"to_user"`
	Amount int32  `json:"amount"`
}

type SendCoinHandler struct {
	SendCoinService service.SendCoinService
}

func NewSendCoinHandler(sendCoinService service.SendCoinService) *SendCoinHandler {
	return &SendCoinHandler{
		SendCoinService: sendCoinService,
	}
}

// POST /api/send-coin
func (h *SendCoinHandler) HandleSendCoin(w http.ResponseWriter, r *http.Request) {
	var req SendCoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Базовая валидация формата
	if req.ToUser == "" {
		utils.JSONErrorResponse(w, http.StatusBadRequest, "to_user is required")
		return
	}

	if req.Amount <= 0 {
		utils.JSONErrorResponse(w, http.StatusBadRequest, "amount must be positive")
		return
	}

	err := h.SendCoinService.SendCoin(r.Context(), req.ToUser, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBusinessValidation):
			utils.JSONErrorResponse(w, http.StatusBadRequest, err.Error())
		default:
			utils.JSONErrorResponse(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}
