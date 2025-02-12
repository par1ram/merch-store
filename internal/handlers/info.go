package handlers

import (
	"net/http"

	"github.com/par1ram/merch-store/internal/middleware"
	"github.com/par1ram/merch-store/internal/service"
	"github.com/par1ram/merch-store/internal/utils"
)

type InfoHandler struct {
	InfoService service.InfoService
}

func NewInfoHandler(infoService service.InfoService) *InfoHandler {
	return &InfoHandler{InfoService: infoService}
}

// GET /api/info.
func (h *InfoHandler) HandleInfo(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == 0 {
		utils.JSONErrorResponse(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	info, err := h.InfoService.GetInfo(r.Context(), userID)
	if err != nil {
		utils.JSONErrorResponse(w, http.StatusInternalServerError, "failed to retrieve info")
		return
	}

	utils.JSONResponse(w, http.StatusOK, info)
}
