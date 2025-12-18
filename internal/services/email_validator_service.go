package services

import "context"

type EmailValidator interface {
	Validate(ctx context.Context, email string) error
}
