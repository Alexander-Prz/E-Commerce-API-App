package model

import "time"

type EmailVerification struct {
	ID        int64
	AuthID    int64
	Token     string
	ExpiresAt time.Time
}
