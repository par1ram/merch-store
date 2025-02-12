package handlers

import (
	"net/http"
	"strings"

	"github.com/par1ram/merch-store/internal/service"
	"github.com/par1ram/merch-store/internal/utils"
)

type BuyHandler struct {
	BuyService service.BuyService
}

func NewBuyHandler(buyService service.BuyService) *BuyHandler {
	return &BuyHandler{
		BuyService: buyService,
	}
}

func (h *BuyHandler) HandleBuy(w http.ResponseWriter, r *http.Request) {
	// Удаляем префикс "/api/buy/" из URL, чтобы получить название товара.
	item := strings.TrimPrefix(r.URL.Path, "/api/buy/")
	if item == "" {
		utils.JSONErrorResponse(w, http.StatusBadRequest, "item is required in URL")
		return
	}

	if err := h.BuyService.Purchase(r.Context(), item); err != nil {
		utils.JSONErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{"status": "purchase successful"})
}
