package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Fixsbreaker/weather_with_db/internal/dto"
	"github.com/Fixsbreaker/weather_with_db/internal/model"
	"github.com/Fixsbreaker/weather_with_db/internal/service"
)

type authService interface {
	Register(ctx context.Context, in dto.RegisterRequest) (*model.User, error)
	Login(ctx context.Context, in dto.LoginRequest) (string, error)
}

type AuthHandler struct {
	svc authService
}

func NewAuthHandler(svc authService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.svc.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "failed to register user", http.StatusInternalServerError)
		}
		return
	}

	resp := dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		DeletedAt: user.DeletedAt,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.svc.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
		} else {
			http.Error(w, "failed to login", http.StatusInternalServerError)
		}
		return
	}

	resp := dto.AuthResponse{AccessToken: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
