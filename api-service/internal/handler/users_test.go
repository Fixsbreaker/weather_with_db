package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Fixsbreaker/weather_with_db/internal/dto"
	"github.com/Fixsbreaker/weather_with_db/internal/model"
	"github.com/Fixsbreaker/weather_with_db/internal/repository"
	"github.com/Fixsbreaker/weather_with_db/internal/service"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) Create(ctx context.Context, in dto.CreateUserInput) (*model.User, error) {
	args := m.Called(ctx, in)
	var user *model.User
	if args.Get(0) != nil {
		user = args.Get(0).(*model.User)
	}
	return user, args.Error(1)
}

func (m *mockUserService) List(ctx context.Context) ([]*model.User, error) {
	args := m.Called(ctx)
	var users []*model.User
	if args.Get(0) != nil {
		users = args.Get(0).([]*model.User)
	}
	return users, args.Error(1)
}

func (m *mockUserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	args := m.Called(ctx, id)
	var user *model.User
	if args.Get(0) != nil {
		user = args.Get(0).(*model.User)
	}
	return user, args.Error(1)
}

func (m *mockUserService) Update(ctx context.Context, id int64, in dto.CreateUserInput) (*model.User, error) {
	args := m.Called(ctx, id, in)
	var user *model.User
	if args.Get(0) != nil {
		user = args.Get(0).(*model.User)
	}
	return user, args.Error(1)
}

func (m *mockUserService) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUserHandler_GetByID(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		mockSvc := new(mockUserService)
		handler := NewUserHandler(mockSvc)

		expectedUser := &model.User{ID: 1, Name: "Test", Email: "test@test.com", Role: "user"}
		mockSvc.On("GetByID", mock.Anything, int64(1)).Return(expectedUser, nil)

		r := chi.NewRouter()
		r.Get("/users/{id}", handler.GetByID)

		req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		
		var resp dto.UserResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.ID)
		assert.Equal(t, "Test", resp.Name)
		
		mockSvc.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		mockSvc := new(mockUserService)
		handler := NewUserHandler(mockSvc)

		r := chi.NewRouter()
		r.Get("/users/{id}", handler.GetByID)

		req := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		
		var resp map[string]string
		json.NewDecoder(rec.Body).Decode(&resp)
		assert.Equal(t, "invalid id", resp["error"])
	})

	t.Run("Not found", func(t *testing.T) {
		mockSvc := new(mockUserService)
		handler := NewUserHandler(mockSvc)

		mockSvc.On("GetByID", mock.Anything, int64(99)).Return(nil, repository.ErrNotFound)

		r := chi.NewRouter()
		r.Get("/users/{id}", handler.GetByID)

		req := httptest.NewRequest(http.MethodGet, "/users/99", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestUserHandler_Create(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		mockSvc := new(mockUserService)
		handler := NewUserHandler(mockSvc)

		in := dto.CreateUserInput{Name: "Bob", Email: "bob@test.com"}
		expectedUser := &model.User{ID: 2, Name: "Bob", Email: "bob@test.com"}
		
		mockSvc.On("Create", mock.Anything, in).Return(expectedUser, nil)

		body, _ := json.Marshal(in)
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handler.Create(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		mockSvc := new(mockUserService)
		handler := NewUserHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("{invalid-json}")))
		rec := httptest.NewRecorder()

		handler.Create(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		
		var resp map[string]string
		json.NewDecoder(rec.Body).Decode(&resp)
		assert.Equal(t, "invalid request body", resp["error"])
	})

	t.Run("Validation Error", func(t *testing.T) {
		mockSvc := new(mockUserService)
		handler := NewUserHandler(mockSvc)

		in := dto.CreateUserInput{Name: "", Email: "bob@test.com"}
		mockSvc.On("Create", mock.Anything, in).Return(nil, service.ErrValidation)

		body, _ := json.Marshal(in)
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handler.Create(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		mockSvc.AssertExpectations(t)
	})
}
