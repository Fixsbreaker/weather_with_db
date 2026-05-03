package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Fixsbreaker/weather_with_db/internal/model"
)

type userRepo interface {
	Create(ctx context.Context, name, email, passwordHash, role string) (*model.User, error)
	List(ctx context.Context) ([]*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, id int64, name, email string) (*model.User, error)
	SoftDelete(ctx context.Context, id int64) error
}

type UserService struct {
	repo      userRepo
	jwtSecret string
}

func NewUserService(repo userRepo, jwtSecret string) *UserService {
	return &UserService{repo: repo, jwtSecret: jwtSecret}
}

type CreateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

var ErrValidation = errors.New("validation error")
var ErrUnauthorized = errors.New("unauthorized")

func validateUser(name, email string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if !strings.Contains(email, "@") {
		return fmt.Errorf("%w: invalid email", ErrValidation)
	}
	return nil
}

// Create is kept for backward compatibility or admin creation without password.
func (s *UserService) Create(ctx context.Context, in CreateUserInput) (*model.User, error) {
	if err := validateUser(in.Name, in.Email); err != nil {
		return nil, err
	}
	// For normal create via admin, we could generate a random password or leave empty
	return s.repo.Create(ctx, strings.TrimSpace(in.Name), strings.TrimSpace(in.Email), "", "user")
}

func (s *UserService) Register(ctx context.Context, in model.RegisterRequest) (*model.User, error) {
	if err := validateUser(in.Name, in.Email); err != nil {
		return nil, err
	}
	if len(in.Password) < 6 {
		return nil, fmt.Errorf("%w: password must be at least 6 characters", ErrValidation)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return s.repo.Create(ctx, strings.TrimSpace(in.Name), strings.TrimSpace(in.Email), string(hash), "user")
}

func (s *UserService) Login(ctx context.Context, in model.LoginRequest) (string, error) {
	user, err := s.repo.GetByEmail(ctx, in.Email)
	if err != nil {
		return "", ErrUnauthorized // don't leak whether user exists or password wrong
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return "", ErrUnauthorized
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
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
