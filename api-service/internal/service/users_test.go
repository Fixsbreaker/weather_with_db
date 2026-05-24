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
)

func TestUserService_GetByID(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	svc := NewUserService(mockRepo, "secret")
	ctx := context.Background()

	t.Run("Happy path", func(t *testing.T) {
		expectedUser := &model.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}

		mockRepo.On("GetByID", ctx, int64(1)).Return(expectedUser, nil).Once()

		user, err := svc.GetByID(ctx, 1)

		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not found", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, int64(99)).Return(nil, repository.ErrNotFound).Once()

		user, err := svc.GetByID(ctx, 99)

		require.ErrorIs(t, err, repository.ErrNotFound)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_Register(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	svc := NewUserService(mockRepo, "secret")
	ctx := context.Background()

	t.Run("Happy path", func(t *testing.T) {
		req := dto.RegisterRequest{
			Name:     "Alice",
			Email:    "alice@test.com",
			Password: "password123",
		}
		
		expectedUser := &model.User{
			ID:        1,
			Name:      "Alice",
			Email:     "alice@test.com",
			Role:      "user",
			CreatedAt: time.Now(),
		}

		mockRepo.On("Create", ctx, "Alice", "alice@test.com", mock.Anything, "user").Return(expectedUser, nil).Once()

		user, err := svc.Register(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Validation error - empty name", func(t *testing.T) {
		req := dto.RegisterRequest{
			Name:     "",
			Email:    "alice@test.com",
			Password: "password123",
		}

		user, err := svc.Register(ctx, req)

		require.ErrorIs(t, err, ErrValidation)
		assert.Contains(t, err.Error(), "name is required")
		assert.Nil(t, user)
	})

	t.Run("Validation error - invalid email", func(t *testing.T) {
		req := dto.RegisterRequest{
			Name:     "Alice",
			Email:    "alicetest.com",
			Password: "password123",
		}

		user, err := svc.Register(ctx, req)

		require.ErrorIs(t, err, ErrValidation)
		assert.Contains(t, err.Error(), "invalid email")
		assert.Nil(t, user)
	})

	t.Run("Validation error - password too short", func(t *testing.T) {
		req := dto.RegisterRequest{
			Name:     "Alice",
			Email:    "alice@test.com",
			Password: "123",
		}

		user, err := svc.Register(ctx, req)

		require.ErrorIs(t, err, ErrValidation)
		assert.Contains(t, err.Error(), "password must be at least 6 characters")
		assert.Nil(t, user)
	})
}

func TestUserService_List(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	svc := NewUserService(mockRepo, "secret")
	ctx := context.Background()

	t.Run("Happy path", func(t *testing.T) {
		expectedUsers := []*model.User{
			{ID: 1, Name: "User 1"},
			{ID: 2, Name: "User 2"},
		}
		mockRepo.On("List", ctx).Return(expectedUsers, nil).Once()

		users, err := svc.List(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedUsers, users)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_Delete(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	svc := NewUserService(mockRepo, "secret")
	ctx := context.Background()

	t.Run("Happy path", func(t *testing.T) {
		mockRepo.On("SoftDelete", ctx, int64(1)).Return(nil).Once()

		err := svc.Delete(ctx, 1)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_Create(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	svc := NewUserService(mockRepo, "secret")
	ctx := context.Background()

	t.Run("Happy path", func(t *testing.T) {
		req := dto.CreateUserInput{
			Name:  "Admin",
			Email: "admin@test.com",
		}
		expectedUser := &model.User{ID: 1, Name: "Admin", Email: "admin@test.com", Role: "user"}

		mockRepo.On("Create", ctx, "Admin", "admin@test.com", "", "user").Return(expectedUser, nil).Once()

		user, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})
}
