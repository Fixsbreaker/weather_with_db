package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Fixsbreaker/weather_with_db/internal/dto"
	"github.com/Fixsbreaker/weather_with_db/internal/middleware"
	"github.com/Fixsbreaker/weather_with_db/internal/model"
)

type cityService interface {
	Add(ctx context.Context, userID int64, in dto.AddCityInput) (*model.City, error)
	List(ctx context.Context, userID int64) ([]*model.City, error)
	Delete(ctx context.Context, userID, cityID int64) error
}

type CityHandler struct {
	svc cityService
}

func NewCityHandler(svc cityService) *CityHandler {
	return &CityHandler{svc: svc}
}

func (h *CityHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var in dto.AddCityInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	city, err := h.svc.Add(r.Context(), userID, in)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapCityToResponse(city))
}

func (h *CityHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cities, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	var resp []dto.CityResponse
	for _, c := range cities {
		resp = append(resp, mapCityToResponse(c))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *CityHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	rawCityID := chi.URLParam(r, "city_id")
	cityID, err := strconv.ParseInt(rawCityID, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid city_id")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, cityID); err != nil {
		handleServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

func mapCityToResponse(c *model.City) dto.CityResponse {
	return dto.CityResponse{
		ID:     c.ID,
		UserID: c.UserID,
		Name:   c.Name,
	}
}
