package neo

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type Session struct {
	ID string `json:"sessionId"`
}

// NewSession creates a new session for the given bank ID.
func (a *API) NewSession(ctx context.Context, bankID string) (*Session, error) {
	if bankID == "" {
		return nil, ErrInvalidBankID
	}
	body := fmt.Sprintf(`{"bankId":"%s"}`, bankID)
	req := a.request(ctx, http.MethodPost, "/ics/v3/session", strings.NewReader(body))
	s := &Session{}
	_, err := a.do(req, http.StatusCreated, s)
	return s, err
}

type SessionStatus struct {
	BankID     string `json:"bankId"`
	BankName   string `json:"bankName"`
	CreatedAt  string `json:"createdAt"`
	ProviderID string `json:"providerId"`
}

// Status returns the details of the given session.
func (a *API) Status(ctx context.Context, sessionID string) (*SessionStatus, error) {
	if sessionID == "" {
		return nil, ErrInvalidSessionID
	}
	req := a.request(ctx, http.MethodGet, "/ics/v3/session/"+sessionID, nil)
	s := &SessionStatus{}
	_, err := a.do(req, http.StatusOK, s)
	return s, err
}

// DeleteSession deletes & invalidates a given session.
func (a *API) DeleteSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return ErrInvalidSessionID
	}
	req := a.request(ctx, http.MethodDelete, "/ics/v3/session/"+sessionID, nil)
	_, err := a.do(req, http.StatusNoContent, nil)
	return err
}
