package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Fixsbreaker/weather_with_db/internal/model"
)

type WeatherRepository struct {
	db *pgxpool.Pool
}

func NewWeatherRepository(db *pgxpool.Pool) *WeatherRepository {
	return &WeatherRepository{db: db}
}

func (r *WeatherRepository) Save(ctx context.Context, userID int64, city string, temp float64, desc string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO weather_history (user_id, city, temperature, description)
		 VALUES ($1, $2, $3, $4)`,
		userID, city, temp, desc,
	)
	return err
}

type HistoryFilter struct {
	City   string
	Limit  int // 0 = no limit
	Offset int
}

func (r *WeatherRepository) GetHistory(ctx context.Context, userID int64, f HistoryFilter) ([]model.WeatherHistory, error) {
	// Build query dynamically but safely — city always goes through $N placeholder
	query := `SELECT id, user_id, city, temperature, description, requested_at
	          FROM weather_history
	          WHERE user_id = $1 AND city = $2
	          ORDER BY requested_at DESC`

	args := []any{userID, f.City}

	if f.Limit > 0 {
		args = append(args, f.Limit)
		query += ` LIMIT $` + itoa(len(args))
	}
	if f.Offset > 0 {
		args = append(args, f.Offset)
		query += ` OFFSET $` + itoa(len(args))
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.WeatherHistory
	for rows.Next() {
		var h model.WeatherHistory
		if err := rows.Scan(&h.ID, &h.UserID, &h.City, &h.Temperature, &h.Description, &h.RequestedAt); err != nil {
			return nil, err
		}
		result = append(result, h)
	}
	return result, rows.Err()
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
