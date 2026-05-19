package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/Fixsbreaker/weather_with_db/internal/model"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, name, email, passwordHash, role string) (*model.User, error) {
	args := m.Called(ctx, name, email, passwordHash, role)
	
	var user *model.User
	if args.Get(0) != nil {
		user = args.Get(0).(*model.User)
	}
	return user, args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context) ([]*model.User, error) {
	args := m.Called(ctx)
	
	var users []*model.User
	if args.Get(0) != nil {
		users = args.Get(0).([]*model.User)
	}
	return users, args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	args := m.Called(ctx, id)
	
	var user *model.User
	if args.Get(0) != nil {
		user = args.Get(0).(*model.User)
	}
	return user, args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	
	var user *model.User
	if args.Get(0) != nil {
		user = args.Get(0).(*model.User)
	}
	return user, args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, id int64, name, email string) (*model.User, error) {
	args := m.Called(ctx, id, name, email)
	
	var user *model.User
	if args.Get(0) != nil {
		user = args.Get(0).(*model.User)
	}
	return user, args.Error(1)
}

func (m *MockUserRepository) SoftDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
