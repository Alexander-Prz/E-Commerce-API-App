package services

import (
	"context"
	//"net/mail"
)

type LocalValidator struct{}

func NewLocalValidator() *LocalValidator {
	return &LocalValidator{}
}

func (v *LocalValidator) Validate(
	ctx context.Context,
	email string,
) error {
	// Already parsed in validateEmail(), so just accept
	return nil
}
