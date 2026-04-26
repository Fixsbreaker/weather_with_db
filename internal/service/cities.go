package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Fixsbreaker/weather_with_db/internal/model"
)

type cityRepo interface {
	Add(ctx context.Context, userID int64, name string) (*model.City, error)
	ListByUser(ctx context.Context, userID int64) ([]*model.City, error)
	Delete(ctx context.Context, userID, cityID int64) error
	GetCityNames(ctx context.Context, userID int64) ([]string, error)
}

type CityService struct {
	repo     cityRepo
	userRepo userRepo
}

func NewCityService(repo cityRepo, userRepo userRepo) *CityService {
	return &CityService{repo: repo, userRepo: userRepo}
}

type AddCityInput struct {
	Name string `json:"name"`
}

func (s *CityService) Add(ctx context.Context, userID int64, in AddCityInput) (*model.City, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: city name is required", ErrValidation)
	}

	// ensure user exists and is active
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.DeletedAt != nil {
		return nil, fmt.Errorf("%w: user is deleted", ErrValidation)
	}

	return s.repo.Add(ctx, userID, name)
}

func (s *CityService) List(ctx context.Context, userID int64) ([]*model.City, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *CityService) Delete(ctx context.Context, userID, cityID int64) error {
	return s.repo.Delete(ctx, userID, cityID)
}
