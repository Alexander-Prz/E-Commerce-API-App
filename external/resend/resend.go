package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"
)

type ResendMailer struct {
	apiKey  string
	from    string
	client  *http.Client
	baseURL string
}

func NewResendMailer(from string) (*ResendMailer, error) {
	key := os.Getenv("RESEND_API_KEY")
	if key == "" {
		return nil, errors.New("RESEND_API_KEY not set")
	}

	return &ResendMailer{
		apiKey: key,
		from:   from,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		baseURL: "https://api.resend.com",
	}, nil
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (m *ResendMailer) SendVerificationEmail(
	ctx context.Context,
	toEmail string,
	verifyURL string,
) error {
	body := sendRequest{
		From:    m.from,
		To:      []string{toEmail},
		Subject: "Verify your email",
		HTML: `
			<p>Welcome!</p>
			<p>Please verify your email by clicking the link below:</p>
			<p><a href="` + verifyURL + `">Verify Email</a></p>
		`,
	}

	b, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		m.baseURL+"/emails",
		bytes.NewBuffer(b),
	)

	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return errors.New(
			"failed to send verification email: " + buf.String(),
		)
	}

	return nil
}
