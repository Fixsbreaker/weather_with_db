package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Fixsbreaker/weather_with_db/internal/dto"
	"github.com/Fixsbreaker/weather_with_db/internal/model"
	"github.com/Fixsbreaker/weather_with_db/internal/repository"
	"github.com/Fixsbreaker/weather_with_db/internal/repository/mocks"
	"github.com/Fixsbreaker/weather_with_db/internal/weather"
)

type mockWeatherRepo struct {
	mock.Mock
}

func (m *mockWeatherRepo) Save(ctx context.Context, userID int64, city string, temp float64, desc string) error {
	args := m.Called(ctx, userID, city, temp, desc)
	return args.Error(0)
}

func (m *mockWeatherRepo) GetHistory(ctx context.Context, userID int64, f repository.HistoryFilter) ([]model.WeatherHistory, error) {
	args := m.Called(ctx, userID, f)
	var history []model.WeatherHistory
	if args.Get(0) != nil {
		history = args.Get(0).([]model.WeatherHistory)
	}
	return history, args.Error(1)
}

type mockWeatherClient struct {
	mock.Mock
}

func (m *mockWeatherClient) GetWeather(ctx context.Context, city string) (*weather.WeatherData, error) {
	args := m.Called(ctx, city)
	var data *weather.WeatherData
	if args.Get(0) != nil {
		data = args.Get(0).(*weather.WeatherData)
	}
	return data, args.Error(1)
}

func TestWeatherService_GetCurrentWeather(t *testing.T) {
	mockWeatherRepo := new(mockWeatherRepo)
	mockCityRepo := new(mockCityRepo) // uses the one defined in cities_test.go
	mockUserRepo := new(mocks.MockUserRepository)
	mockClient := new(mockWeatherClient)

	svc := NewWeatherService(mockWeatherRepo, mockCityRepo, mockUserRepo, mockClient)
	ctx := context.Background()

	t.Run("Happy path", func(t *testing.T) {
		mockUserRepo.On("GetByID", ctx, int64(1)).Return(&model.User{ID: 1}, nil).Once()
		mockCityRepo.On("GetCityNames", ctx, int64(1)).Return([]string{"London"}, nil).Once()
		
		mockClient.On("GetWeather", ctx, "London").Return(&weather.WeatherData{Temperature: 15.0, Description: "Cloudy"}, nil).Once()
		mockWeatherRepo.On("Save", ctx, int64(1), "London", 15.0, "Cloudy").Return(nil).Once()

		result, err := svc.GetCurrentWeather(ctx, 1)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "London", result[0].City)
		assert.Equal(t, 15.0, result[0].Temperature)
	})

	t.Run("No cities", func(t *testing.T) {
		mockUserRepo.On("GetByID", ctx, int64(1)).Return(&model.User{ID: 1}, nil).Once()
		mockCityRepo.On("GetCityNames", ctx, int64(1)).Return([]string{}, nil).Once()

		result, err := svc.GetCurrentWeather(ctx, 1)

		require.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestWeatherService_GetHistory(t *testing.T) {
	mockWeatherRepo := new(mockWeatherRepo)
	mockCityRepo := new(mockCityRepo)
	mockUserRepo := new(mocks.MockUserRepository)
	mockClient := new(mockWeatherClient)

	svc := NewWeatherService(mockWeatherRepo, mockCityRepo, mockUserRepo, mockClient)
	ctx := context.Background()

	t.Run("Happy path", func(t *testing.T) {
		req := dto.HistoryQuery{City: "London", Limit: 10}
		
		now := time.Now()
		expectedHistory := []model.WeatherHistory{
			{Temperature: 15.0, Description: "Cloudy", RequestedAt: now},
		}

		mockWeatherRepo.On("GetHistory", ctx, int64(1), repository.HistoryFilter{City: "London", Limit: 10}).Return(expectedHistory, nil).Once()

		result, err := svc.GetHistory(ctx, 1, req)

		require.NoError(t, err)
		assert.Equal(t, "London", result.City)
		assert.Len(t, result.History, 1)
		assert.Equal(t, 15.0, result.History[0].Temperature)
	})

	t.Run("Empty city error", func(t *testing.T) {
		req := dto.HistoryQuery{City: ""}
		result, err := svc.GetHistory(ctx, 1, req)

		require.ErrorIs(t, err, ErrValidation)
		assert.Nil(t, result)
	})
}
