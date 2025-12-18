package services

import "context"

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, toEmail, verifyURL string) error
}
