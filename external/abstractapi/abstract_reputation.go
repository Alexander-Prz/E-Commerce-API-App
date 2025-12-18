package abstractapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type AbstractReputationValidator struct {
	apiKey string
	client *http.Client
}

func NewAbstractReputationValidator() (*AbstractReputationValidator, error) {
	key := os.Getenv("ABSTRACT_EMAIL_API_KEY")
	if key == "" {
		return nil, errors.New("ABSTRACT_EMAIL_API_KEY not set")
	}

	return &AbstractReputationValidator{
		apiKey: key,
		client: &http.Client{Timeout: 5 * time.Second},
	}, nil
}

type reputationResponse struct {
	EmailReputation string `json:"email_reputation"` // LOW, MEDIUM, HIGH
	IsDisposable    bool   `json:"is_disposable_email"`
	IsRoleEmail     bool   `json:"is_role_email"`
}

func (v *AbstractReputationValidator) Validate(
	ctx context.Context,
	email string,
) error {
	u, _ := url.Parse("https://emailreputation.abstractapi.com/v1/")
	q := u.Query()
	q.Set("api_key", v.apiKey)
	q.Set("email", email)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("email reputation service error: %s", resp.Status)
	}

	var out reputationResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}

	// ---- Rules ----
	if out.IsDisposable {
		return errors.New("disposable email is not allowed")
	}

	if out.IsRoleEmail {
		return errors.New("role-based email is not allowed")
	}

	if out.EmailReputation == "LOW" {
		return errors.New("email reputation is too low")
	}

	return nil
}
