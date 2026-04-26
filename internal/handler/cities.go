package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Fixsbreaker/weather_with_db/internal/model"
	"github.com/Fixsbreaker/weather_with_db/internal/service"
)

type cityService interface {
	Add(ctx context.Context, userID int64, in service.AddCityInput) (*model.City, error)
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
	userID, ok := parseID(w, r)
	if !ok {
		return
	}

	var in service.AddCityInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	city, err := h.svc.Add(r.Context(), userID, in)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, city)
}

func (h *CityHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseID(w, r)
	if !ok {
		return
	}

	cities, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, cities)
}

func (h *CityHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseID(w, r)
	if !ok {
		return
	}

	rawCityID := chi.URLParam(r, "city_id")
	cityID, err := strconv.ParseInt(rawCityID, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid city_id")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, cityID); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
