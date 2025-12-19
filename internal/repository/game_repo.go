package repository

import (
	"context"
	"errors"
	"time"

	"GameStoreAPI/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GameRepository struct {
	DB *pgxpool.Pool
}

func NewGameRepository(db *pgxpool.Pool) *GameRepository {
	return &GameRepository{DB: db}
}

func (r *GameRepository) CreateGame(ctx context.Context, g *model.Game) (int64, error) {
	var id int64
	query := `INSERT INTO games (developerid, title, price, releasedate, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING gameid`
	if err := r.DB.QueryRow(ctx, query, g.DeveloperID, g.Title, g.Price, g.ReleaseDate, time.Now()).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

/*
func (r *GameRepository) GetByID(ctx context.Context, id int64) (*model.Game, error) {
	var g model.Game
	query := `SELECT gameid, developerid, title, price, releasedate, created_at, deleted_at FROM games WHERE gameid=$1`
	if err := r.DB.QueryRow(ctx, query, id).Scan(&g.GameID, &g.DeveloperID, &g.Title, &g.Price, &g.ReleaseDate, &g.CreatedAt, &g.DeletedAt); err != nil {
		return nil, errors.New("game not found")
	}
	return &g, nil
}
*/

func (r *GameRepository) GetByID(ctx context.Context, id int64) (*model.Game, error) {
	var g model.Game
	query := `
		SELECT gameid, developerid, title, price, releasedate, created_at, deleted_at
		FROM games
		WHERE gameid=$1 AND deleted_at IS NULL
	`
	if err := r.DB.
		QueryRow(ctx, query, id).
		Scan(&g.GameID, &g.DeveloperID, &g.Title, &g.Price, &g.ReleaseDate, &g.CreatedAt, &g.DeletedAt); err != nil {
		return nil, errors.New("game not found")
	}
	return &g, nil
}

func (r *GameRepository) List(ctx context.Context, limit, offset int) ([]model.Game, error) {
	query := `SELECT gameid, developerid, title, price, releasedate, created_at, deleted_at FROM games WHERE deleted_at IS NULL ORDER BY gameid LIMIT $1 OFFSET $2`
	rows, err := r.DB.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Game
	for rows.Next() {
		var g model.Game
		if err := rows.Scan(&g.GameID, &g.DeveloperID, &g.Title, &g.Price, &g.ReleaseDate, &g.CreatedAt, &g.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, nil
}

func (r *GameRepository) ListByDeveloper(ctx context.Context, developerID int64, limit, offset int) ([]model.Game, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	query := `SELECT gameid, developerid, title, price, releasedate, created_at, deleted_at FROM games WHERE developerid=$1 AND deleted_at IS NULL ORDER BY gameid LIMIT $2 OFFSET $3`
	rows, err := r.DB.Query(ctx, query, developerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Game
	for rows.Next() {
		var g model.Game
		if err := rows.Scan(&g.GameID, &g.DeveloperID, &g.Title, &g.Price, &g.ReleaseDate, &g.CreatedAt, &g.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, nil
}

func (r *GameRepository) UpdateGame(ctx context.Context, g *model.Game) error {
	query := `UPDATE games SET developerid=$1, title=$2, price=$3, releasedate=$4 WHERE gameid=$5 AND deleted_at IS NULL`
	tag, err := r.DB.Exec(ctx, query, g.DeveloperID, g.Title, g.Price, g.ReleaseDate, g.GameID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("game not found or deleted")
	}
	return nil
}

func (r *GameRepository) DeleteGame(ctx context.Context, id int64) error {
	query := `UPDATE games SET deleted_at=$1 WHERE gameid=$2 AND deleted_at IS NULL`
	tag, err := r.DB.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("game not found or already deleted")
	}
	return nil
}

func (r *GameRepository) ExistsByDeveloperID(ctx context.Context, developerID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM developers WHERE developerid=$1 AND deleted_at IS NULL)`
	if err := r.DB.QueryRow(ctx, query, developerID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
