package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Fixsbreaker/weather_with_db/internal/config"
	"github.com/Fixsbreaker/weather_with_db/internal/handler"
	customMiddleware "github.com/Fixsbreaker/weather_with_db/internal/middleware"
	"github.com/Fixsbreaker/weather_with_db/internal/repository"
	"github.com/Fixsbreaker/weather_with_db/internal/service"
	"github.com/Fixsbreaker/weather_with_db/internal/weather"
)

func main() {
	cfg := config.Load()

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("connected to database")

	// repositories
	userRepo := repository.NewUserRepository(db)
	cityRepo := repository.NewCityRepository(db)
	weatherRepo := repository.NewWeatherRepository(db)

	// weather client
	weatherClient := weather.NewClient()

	// services
	userSvc := service.NewUserService(userRepo, cfg.JWTSecret)
	citySvc := service.NewCityService(cityRepo, userRepo)
	weatherSvc := service.NewWeatherService(weatherRepo, cityRepo, userRepo, weatherClient)

	// handlers
	authH := handler.NewAuthHandler(userSvc)
	userH := handler.NewUserHandler(userSvc)
	cityH := handler.NewCityHandler(citySvc)
	weatherH := handler.NewWeatherHandler(weatherSvc)

	// router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Public routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware(cfg.JWTSecret))

		r.Get("/users/me", userH.GetMe)

		r.Route("/cities", func(r chi.Router) {
			r.Post("/", cityH.Add)
			r.Get("/", cityH.List)
			r.Delete("/{city_id}", cityH.Delete)
		})

		r.Get("/weather", weatherH.GetWeather)
		r.Get("/weather/history", weatherH.GetHistory)

		// Admin routes
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.RequireRole("admin"))

			r.Route("/users", func(r chi.Router) {
				r.Get("/", userH.List)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", userH.GetByID)
					r.Put("/", userH.Update)
					r.Delete("/", userH.Delete)
				})
			})
		})
	})

	log.Printf("server listening on %s", cfg.Addr())
	if err := http.ListenAndServe(cfg.Addr(), r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
