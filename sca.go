package neo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type (
	SCA struct {
		URL   string // The URL the end-user has to visit and consent to.
		ID    string // Metadata associated with the payload, either SessionID or PaymentID.
		Error *Error // Error returned by the Neonomics platform.
	}

	// SCAMapper maps the consent error into an SCA struct.
	// See DefaultSCAMapper.
	SCAMapper func(d Doer, r *http.Request, e *Error) (*SCA, error)

	SCAHandler struct {
		*SCA
		Retry func(ctx context.Context, v interface{}) (*SCAHandler, error)
	}
)

// DefaultSCAMapper maps a consent error into an SCA struct.
func DefaultSCAMapper(d Doer, r *http.Request, e *Error) (*SCA, error) { //nolint:cyclop
	if d == nil || r == nil || e == nil || len(e.Links) == 0 {
		return nil, ErrInvalidSCAData
	}
	fst := e.Links[0]
	if !strings.Contains(fst.Href, "neonomics") {
		return &SCA{
			URL:   fst.Href,
			ID:    fst.Meta.ID,
			Error: e,
		}, nil
	}
	req, err := http.NewRequestWithContext(r.Context(), fst.Type, fst.Href, nil)
	if err != nil {
		return nil, fmt.Errorf("neo: failed to create a new http.Request: %w", err)
	}
	req.Header = r.Header  // Copy request headers.
	resp, err := d.Do(req) //nolint:bodyclose
	if err != nil {
		return nil, fmt.Errorf("neo: failed to retrieve a SCA response: %w", err)
	}
	defer closeBody(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("neo: invalid SCA response: %s", resp.Status)
	}
	c := &Consent{}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return nil, fmt.Errorf("neo: failed to decode consent: %w", err)
	}
	if len(c.Links) == 0 {
		return nil, ErrInvalidConsent
	}
	var id string
	if c.PaymentID != "" {
		id = c.PaymentID
	} else {
		id = c.Links[0].Meta.ID
	}
	return &SCA{
		URL:   c.Links[0].Href,
		ID:    id,
		Error: e,
	}, nil
}
