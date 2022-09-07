package neo

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidAuthRequest = errors.New("invalid auth request")
	ErrInvalidConsent     = errors.New("invalid consent")
	ErrInvalidBankID      = errors.New("invalid bank ID")
	ErrInvalidSessionID   = errors.New("invalid session ID")
	ErrInvalidAccountID   = errors.New("invalid account ID")
	ErrInvalidSCAData     = errors.New("invalid SCA data")
)

type Error struct {
	ID        string  `json:"id"`
	ErrorCode string  `json:"errorCode"`
	Message   string  `json:"message"`
	Source    string  `json:"source"`
	Type      string  `json:"type"`
	Timestamp int64   `json:"timestamp"`
	Links     []*Link `json:"links"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s error %s: %s", e.Type, e.ErrorCode, e.Message)
}

func (e *Error) IsConsentError() bool {
	return e != nil && e.Type == "CONSENT" && e.ErrorCode == "1426"
}

func (e *Error) IsPaymentAuthError() bool {
	return e != nil && e.Type == "CONSENT" && e.ErrorCode == "1428"
}
