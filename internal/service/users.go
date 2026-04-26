package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Fixsbreaker/weather_with_db/internal/model"
)

type userRepo interface {
	Create(ctx context.Context, name, email string) (*model.User, error)
	List(ctx context.Context) ([]*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	Update(ctx context.Context, id int64, name, email string) (*model.User, error)
	SoftDelete(ctx context.Context, id int64) error
}

type UserService struct {
	repo userRepo
}

func NewUserService(repo userRepo) *UserService {
	return &UserService{repo: repo}
}

type CreateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (s *UserService) Create(ctx context.Context, in CreateUserInput) (*model.User, error) {
	if err := validateUser(in.Name, in.Email); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, strings.TrimSpace(in.Name), strings.TrimSpace(in.Email))
}

func (s *UserService) List(ctx context.Context) ([]*model.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id int64, in CreateUserInput) (*model.User, error) {
	if err := validateUser(in.Name, in.Email); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, strings.TrimSpace(in.Name), strings.TrimSpace(in.Email))
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *UserService) IsActive(u *model.User) bool {
	return u.DeletedAt == nil
}

var ErrValidation = errors.New("validation error")

func validateUser(name, email string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if !strings.Contains(email, "@") {
		return fmt.Errorf("%w: invalid email", ErrValidation)
	}
	return nil
}
