package model

import "time"

type User struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

type City struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
}

type WeatherHistory struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	City        string    `json:"city"`
	Temperature float64   `json:"temperature"`
	Description string    `json:"description"`
	RequestedAt time.Time `json:"requested_at"`
}

// DTO for weather history response
type WeatherHistoryResponse struct {
	UserID  int64          `json:"user_id"`
	City    string         `json:"city"`
	History []WeatherEntry `json:"history"`
}

type WeatherEntry struct {
	Temperature float64   `json:"temperature"`
	Description string    `json:"description"`
	RequestedAt time.Time `json:"requested_at"`
}

// DTO for current weather (aggregated response)
type CityWeather struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Description string  `json:"description"`
}
