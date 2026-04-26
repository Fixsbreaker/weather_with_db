package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Fixsbreaker/weather_with_db/internal/model"
)

type CityRepository struct {
	db *pgxpool.Pool
}

func NewCityRepository(db *pgxpool.Pool) *CityRepository {
	return &CityRepository{db: db}
}

func (r *CityRepository) Add(ctx context.Context, userID int64, name string) (*model.City, error) {
	c := &model.City{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO user_cities (user_id, name) VALUES ($1, $2)
		 RETURNING id, user_id, name`,
		userID, name,
	).Scan(&c.ID, &c.UserID, &c.Name)
	return c, err
}

func (r *CityRepository) ListByUser(ctx context.Context, userID int64) ([]*model.City, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, name FROM user_cities WHERE user_id = $1 ORDER BY id`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[model.City])
}

func (r *CityRepository) Delete(ctx context.Context, userID, cityID int64) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM user_cities WHERE id = $1 AND user_id = $2`,
		cityID, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CityRepository) GetCityNames(ctx context.Context, userID int64) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT name FROM user_cities WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}
