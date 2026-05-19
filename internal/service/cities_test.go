package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Fixsbreaker/weather_with_db/internal/dto"
	"github.com/Fixsbreaker/weather_with_db/internal/model"
	"github.com/Fixsbreaker/weather_with_db/internal/repository/mocks"
)

type mockCityRepo struct {
	mock.Mock
}

func (m *mockCityRepo) Add(ctx context.Context, userID int64, name string) (*model.City, error) {
	args := m.Called(ctx, userID, name)
	var city *model.City
	if args.Get(0) != nil {
		city = args.Get(0).(*model.City)
	}
	return city, args.Error(1)
}

func (m *mockCityRepo) ListByUser(ctx context.Context, userID int64) ([]*model.City, error) {
	args := m.Called(ctx, userID)
	var cities []*model.City
	if args.Get(0) != nil {
		cities = args.Get(0).([]*model.City)
	}
	return cities, args.Error(1)
}

func (m *mockCityRepo) Delete(ctx context.Context, userID, cityID int64) error {
	args := m.Called(ctx, userID, cityID)
	return args.Error(0)
}

func (m *mockCityRepo) GetCityNames(ctx context.Context, userID int64) ([]string, error) {
	args := m.Called(ctx, userID)
	var names []string
	if args.Get(0) != nil {
		names = args.Get(0).([]string)
	}
	return names, args.Error(1)
}

func TestCityService_Add(t *testing.T) {
	mockRepo := new(mockCityRepo)
	mockUserRepo := new(mocks.MockUserRepository)
	svc := NewCityService(mockRepo, mockUserRepo)
	ctx := context.Background()

	t.Run("Happy path", func(t *testing.T) {
		mockUserRepo.On("GetByID", ctx, int64(1)).Return(&model.User{ID: 1}, nil).Once()
		
		expectedCity := &model.City{ID: 1, UserID: 1, Name: "London"}
		mockRepo.On("Add", ctx, int64(1), "London").Return(expectedCity, nil).Once()

		city, err := svc.Add(ctx, 1, dto.AddCityInput{Name: "London"})
		require.NoError(t, err)
		assert.Equal(t, expectedCity, city)
	})

	t.Run("Empty city name", func(t *testing.T) {
		city, err := svc.Add(ctx, 1, dto.AddCityInput{Name: ""})
		require.ErrorIs(t, err, ErrValidation)
		assert.Nil(t, city)
	})
}

func TestCityService_List(t *testing.T) {
	mockRepo := new(mockCityRepo)
	mockUserRepo := new(mocks.MockUserRepository)
	svc := NewCityService(mockRepo, mockUserRepo)
	ctx := context.Background()

	t.Run("Happy path", func(t *testing.T) {
		expectedCities := []*model.City{{ID: 1, Name: "London"}}
		mockRepo.On("ListByUser", ctx, int64(1)).Return(expectedCities, nil).Once()

		cities, err := svc.List(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, expectedCities, cities)
	})
}

func TestCityService_Delete(t *testing.T) {
	mockRepo := new(mockCityRepo)
	mockUserRepo := new(mocks.MockUserRepository)
	svc := NewCityService(mockRepo, mockUserRepo)
	ctx := context.Background()

	t.Run("Happy path", func(t *testing.T) {
		mockRepo.On("Delete", ctx, int64(1), int64(1)).Return(nil).Once()

		err := svc.Delete(ctx, 1, 1)
		require.NoError(t, err)
	})
}
