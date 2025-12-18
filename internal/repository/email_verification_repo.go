package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailVerificationRepository struct {
	db *pgxpool.Pool
}

func NewEmailVerificationRepository(db *pgxpool.Pool) *EmailVerificationRepository {
	return &EmailVerificationRepository{db: db}
}

func (r *EmailVerificationRepository) Create(ctx context.Context, authID int64, token string, exp time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO email_verifications (authid, token, expires_at)
		VALUES ($1, $2, $3)
	`, authID, token, exp)
	return err
}

func (r *EmailVerificationRepository) GetAuthID(ctx context.Context, token string) (int64, error) {
	var authID int64
	err := r.db.QueryRow(ctx, `
		SELECT authid FROM email_verifications
		WHERE token = $1 AND expires_at > now()
	`, token).Scan(&authID)
	return authID, err
}

func (r *EmailVerificationRepository) Delete(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM email_verifications WHERE token = $1`, token)
	return err
}
