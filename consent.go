package neo

import (
	"context"
	"net/http"
)

type Consent struct {
	Message   string  `json:"message"`
	PaymentID string  `json:"paymentId"` // Only present if this is a payment consent.
	Links     []*Link `json:"links"`
}

// Consent returns the information required for the end-user to consent to an action being carried out.
func (a *API) Consent(ctx context.Context, sessionID string, opts ...Optional) (*Consent, error) {
	if sessionID == "" {
		return nil, ErrInvalidSessionID
	}
	req := a.request(ctx, http.MethodGet, "/ics/v3/consent/"+sessionID, nil, opts...)
	c := &Consent{}
	_, err := a.do(req, http.StatusOK, c)
	return c, err
}
