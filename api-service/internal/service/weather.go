package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/Fixsbreaker/weather_with_db/internal/dto"
	"github.com/Fixsbreaker/weather_with_db/internal/model"
	"github.com/Fixsbreaker/weather_with_db/internal/repository"
	"github.com/Fixsbreaker/weather_with_db/internal/weather"
)

type weatherRepo interface {
	Save(ctx context.Context, userID int64, city string, temp float64, desc string) error
	GetHistory(ctx context.Context, userID int64, f repository.HistoryFilter) ([]model.WeatherHistory, error)
}

type cityRepoForWeather interface {
	GetCityNames(ctx context.Context, userID int64) ([]string, error)
}

type weatherClient interface {
	GetWeather(ctx context.Context, city string) (*weather.WeatherData, error)
}

type WeatherService struct {
	weatherRepo weatherRepo
	cityRepo    cityRepoForWeather
	userRepo    userRepo
	client      weatherClient
}

func NewWeatherService(
	weatherRepo weatherRepo,
	cityRepo cityRepoForWeather,
	userRepo userRepo,
	client weatherClient,
) *WeatherService {
	return &WeatherService{
		weatherRepo: weatherRepo,
		cityRepo:    cityRepo,
		userRepo:    userRepo,
		client:      client,
	}
}

// GetCurrentWeather fetches live weather for all user's cities in parallel,
// saves each result to history, and returns an aggregated response.
func (s *WeatherService) GetCurrentWeather(ctx context.Context, userID int64) ([]dto.CityWeather, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.DeletedAt != nil {
		return nil, fmt.Errorf("%w: user is deleted", ErrValidation)
	}

	cities, err := s.cityRepo.GetCityNames(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(cities) == 0 {
		return []dto.CityWeather{}, nil
	}

	type result struct {
		cw  dto.CityWeather
		err error
	}

	results := make([]result, len(cities))
	var wg sync.WaitGroup

	for i, city := range cities {
		wg.Add(1)
		go func(i int, city string) {
			defer wg.Done()

			data, err := s.client.GetWeather(ctx, city)
			if err != nil {
				results[i] = result{err: err}
				return
			}

			// fire-and-forget save; don't fail the whole request if history write fails
			_ = s.weatherRepo.Save(ctx, userID, city, data.Temperature, data.Description)

			results[i] = result{cw: dto.CityWeather{
				City:        city,
				Temperature: data.Temperature,
				Description: data.Description,
			}}
		}(i, city)
	}

	wg.Wait()

	var out []dto.CityWeather
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		out = append(out, r.cw)
	}
	return out, nil
}

func (s *WeatherService) GetHistory(ctx context.Context, userID int64, q dto.HistoryQuery) (*dto.WeatherHistoryResponse, error) {
	if q.City == "" {
		return nil, fmt.Errorf("%w: city is required", ErrValidation)
	}

	records, err := s.weatherRepo.GetHistory(ctx, userID, repository.HistoryFilter{
		City:   q.City,
		Limit:  q.Limit,
		Offset: q.Offset,
	})
	if err != nil {
		return nil, err
	}

	entries := make([]dto.WeatherEntry, len(records))
	for i, r := range records {
		entries[i] = dto.WeatherEntry{
			Temperature: r.Temperature,
			Description: r.Description,
			RequestedAt: r.RequestedAt,
		}
	}

	return &dto.WeatherHistoryResponse{
		UserID:  userID,
		City:    q.City,
		History: entries,
	}, nil
}
