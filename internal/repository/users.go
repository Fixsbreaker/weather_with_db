package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Fixsbreaker/weather_with_db/internal/model"
)

var ErrNotFound = errors.New("not found")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, name, email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (name, email) VALUES ($1, $2)
		 RETURNING id, name, email, created_at, deleted_at`,
		name, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.DeletedAt)
	return u, err
}

func (r *UserRepository) List(ctx context.Context) ([]*model.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, email, created_at, deleted_at
		 FROM users WHERE deleted_at IS NULL ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[model.User])
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, created_at, deleted_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *UserRepository) Update(ctx context.Context, id int64, name, email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(ctx,
		`UPDATE users SET name = $1, email = $2
		 WHERE id = $3 AND deleted_at IS NULL
		 RETURNING id, name, email, created_at, deleted_at`,
		name, email, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *UserRepository) SoftDelete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
