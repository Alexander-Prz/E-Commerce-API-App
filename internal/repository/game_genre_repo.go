package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GameGenreRepository struct {
	DB *pgxpool.Pool
}

func NewGameGenreRepository(db *pgxpool.Pool) *GameGenreRepository {
	return &GameGenreRepository{DB: db}
}

func (r *GameGenreRepository) AddGenreToGame(ctx context.Context, gameID, genreID int64) error {
	query := `INSERT INTO gamegenres (gameid, genreid) VALUES ($1, $2)
			  ON CONFLICT DO NOTHING`
	_, err := r.DB.Exec(ctx, query, gameID, genreID)
	return err
}

func (r *GameGenreRepository) RemoveGenreFromGame(ctx context.Context, gameID, genreID int64) error {
	query := `DELETE FROM gamegenres WHERE gameid=$1 AND genreid=$2`
	tag, err := r.DB.Exec(ctx, query, gameID, genreID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("genre was not assigned to game")
	}
	return nil
}

func (r *GameGenreRepository) GetGenresByGame(ctx context.Context, gameID int64) ([]int64, error) {
	query := `SELECT genreid FROM gamegenres WHERE gameid=$1`
	rows, err := r.DB.Query(ctx, query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
