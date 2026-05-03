package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Fixsbreaker/weather_with_db/internal/middleware"
	"github.com/Fixsbreaker/weather_with_db/internal/model"
	"github.com/Fixsbreaker/weather_with_db/internal/service"
)

type weatherService interface {
	GetCurrentWeather(ctx context.Context, userID int64) ([]model.CityWeather, error)
	GetHistory(ctx context.Context, userID int64, q service.HistoryQuery) (*model.WeatherHistoryResponse, error)
}

type WeatherHandler struct {
	svc weatherService
}

func NewWeatherHandler(svc weatherService) *WeatherHandler {
	return &WeatherHandler{svc: svc}
}

func (h *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	data, err := h.svc.GetCurrentWeather(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *WeatherHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	city := r.URL.Query().Get("city")
	if city == "" {
		writeError(w, http.StatusBadRequest, "city query parameter is required")
		return
	}

	limit := parseQueryInt(r, "limit", 0)
	offset := parseQueryInt(r, "offset", 0)

	resp, err := h.svc.GetHistory(r.Context(), userID, service.HistoryQuery{
		City:   city,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func parseQueryInt(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return fallback
	}
	return v
}
