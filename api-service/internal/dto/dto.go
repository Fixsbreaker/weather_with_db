package dto

import "time"

// --- Auth DTOs ---

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

// --- User DTOs ---

type UserResponse struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type CreateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// --- City DTOs ---

type AddCityInput struct {
	Name string `json:"name"`
}

type CityResponse struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
}

// --- Weather DTOs ---

type CityWeather struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Description string  `json:"description"`
}

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

type HistoryQuery struct {
	City   string
	Limit  int
	Offset int
}
