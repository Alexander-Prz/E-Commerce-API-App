package repository

import (
	"context"
	"errors"
	"time"

	"GameStoreAPI/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DeveloperRepository struct {
	DB *pgxpool.Pool
}

func NewDeveloperRepository(db *pgxpool.Pool) *DeveloperRepository {
	return &DeveloperRepository{DB: db}
}

// CreateDeveloper inserts a new developer and returns the created id.
func (r *DeveloperRepository) CreateDeveloper(ctx context.Context, name string, authID *int64) (int64, error) {
	var id int64
	query := `INSERT INTO developers (developername, authid, created_at) VALUES ($1, $2, $3) RETURNING developerid`
	if err := r.DB.QueryRow(ctx, query, name, authID, time.Now()).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *DeveloperRepository) GetByID(ctx context.Context, id int64) (*model.Developer, error) {
	var d model.Developer
	query := `SELECT developerid, developername, created_at, deleted_at FROM developers WHERE developerid=$1`
	if err := r.DB.QueryRow(ctx, query, id).Scan(&d.DeveloperID, &d.DeveloperName, &d.CreatedAt, &d.DeletedAt); err != nil {
		return nil, errors.New("developer not found")
	}
	return &d, nil
}

func (r *DeveloperRepository) GetByIDAdmin(ctx context.Context, id int64) (*model.Developer, error) {
	var d model.Developer
	query := `SELECT developerid, developername, authid, created_at, deleted_at FROM developers WHERE developerid=$1`
	if err := r.DB.QueryRow(ctx, query, id).Scan(&d.DeveloperID, &d.DeveloperName, &d.AuthID, &d.CreatedAt, &d.DeletedAt); err != nil {
		return nil, errors.New("developer not found")
	}
	return &d, nil
}

// GetByAuthID returns the developer row associated with the given authid
func (r *DeveloperRepository) GetByAuthID(ctx context.Context, authID int64) (*model.Developer, error) {
	var d model.Developer
	query := `SELECT developerid, developername, authid, created_at, deleted_at FROM developers WHERE authid=$1 AND deleted_at IS NULL`
	if err := r.DB.QueryRow(ctx, query, authID).Scan(&d.DeveloperID, &d.DeveloperName, &d.AuthID, &d.CreatedAt, &d.DeletedAt); err != nil {
		return nil, errors.New("developer not found")
	}
	return &d, nil
}

func (r *DeveloperRepository) GetAll(ctx context.Context) ([]model.Developer, error) {
	query := `SELECT developerid, developername, created_at, deleted_at FROM developers WHERE deleted_at IS NULL ORDER BY developerid`
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Developer
	for rows.Next() {
		var d model.Developer
		if err := rows.Scan(&d.DeveloperID, &d.DeveloperName, &d.CreatedAt, &d.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, nil
}

func (r *DeveloperRepository) GetAllAdmin(ctx context.Context) ([]model.Developer, error) {
	query := `SELECT developerid, developername, authid, created_at, deleted_at FROM developers WHERE deleted_at IS NULL ORDER BY developerid`
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Developer
	for rows.Next() {
		var d model.Developer
		if err := rows.Scan(&d.DeveloperID, &d.DeveloperName, &d.AuthID, &d.CreatedAt, &d.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, nil
}

func (r *DeveloperRepository) UpdateDeveloper(ctx context.Context, id int64, name string, authID *int64) error {
	query := `UPDATE developers SET developername=$1, authid=$2 WHERE developerid=$3 AND deleted_at IS NULL`
	tag, err := r.DB.Exec(ctx, query, name, authID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("developer not found or already deleted")
	}
	return nil
}

func (r *DeveloperRepository) DeleteDeveloper(ctx context.Context, id int64) error {
	query := `UPDATE developers SET deleted_at=$1 WHERE developerid=$2 AND deleted_at IS NULL`
	tag, err := r.DB.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("developer not found or already deleted")
	}
	return nil
}

func (r *DeveloperRepository) NameExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM developers WHERE developername=$1 AND deleted_at IS NULL)`
	if err := r.DB.QueryRow(ctx, query, name).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
