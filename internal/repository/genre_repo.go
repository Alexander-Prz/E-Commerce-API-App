package repository

import (
	"context"
	"errors"
	"time"

	"GameStoreAPI/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GenreRepository struct {
	DB *pgxpool.Pool
}

func NewGenreRepository(db *pgxpool.Pool) *GenreRepository {
	return &GenreRepository{DB: db}
}

func (r *GenreRepository) Create(ctx context.Context, name string) (int64, error) {
	var id int64
	query := `INSERT INTO genres (genrename, created_at) VALUES ($1, $2) RETURNING genreid`
	if err := r.DB.QueryRow(ctx, query, name, time.Now()).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *GenreRepository) GetByID(ctx context.Context, id int64) (*model.Genre, error) {
	var g model.Genre
	query := `SELECT genreid, genrename FROM genres WHERE genreid=$1`
	if err := r.DB.QueryRow(ctx, query, id).Scan(&g.GenreID, &g.GenreName); err != nil {
		return nil, errors.New("genre not found")
	}
	return &g, nil
}

func (r *GenreRepository) List(ctx context.Context) ([]model.Genre, error) {
	query := `SELECT genreid, genrename FROM genres WHERE deleted_at IS NULL ORDER BY genreid`
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Genre
	for rows.Next() {
		var g model.Genre
		if err := rows.Scan(&g.GenreID, &g.GenreName); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, nil
}

func (r *GenreRepository) Update(ctx context.Context, id int64, name string) error {
	query := `UPDATE genres SET genrename=$1 WHERE genreid=$2 AND deleted_at IS NULL`
	tag, err := r.DB.Exec(ctx, query, name, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("genre not found or already deleted")
	}
	return nil
}

func (r *GenreRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE genres SET deleted_at=$1 WHERE genreid=$2 AND deleted_at IS NULL`
	tag, err := r.DB.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("genre not found or already deleted")
	}
	return nil
}

func (r *GenreRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM genres WHERE genrename=$1 AND deleted_at IS NULL)`
	if err := r.DB.QueryRow(ctx, query, name).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
