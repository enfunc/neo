package neo

import (
	"context"
	"net/http"
)

type Account struct {
	ID                    string     `json:"id"`
	BBAN                  string     `json:"bban"`
	IBAN                  string     `json:"iban"`
	SortCodeAccountNumber string     `json:"sortCodeAccountNumber"`
	AccountName           string     `json:"accountName"`
	AccountType           string     `json:"accountType"`
	OwnerName             string     `json:"ownerName"`
	DisplayName           string     `json:"displayName"`
	Balances              []*Balance `json:"balances"`
}

type Balance struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Type     string `json:"type"`
}

// Accounts returns a list of all accounts available in the given session.
func (a *API) Accounts(
	ctx context.Context,
	sessionID string,
	opts ...Optional,
) ([]*Account, *SCAHandler, error) {
	if sessionID == "" {
		return nil, nil, ErrInvalidSessionID
	}
	opts = append(opts, SessionID(sessionID))
	req := a.request(ctx, http.MethodGet, "/ics/v3/accounts", nil, opts...)
	acc := make([]*Account, 0, 8)
	sca, err := a.do(req, http.StatusOK, &acc)
	return acc, sca, err
}

// AccountByID returns an account with the given ID.
func (a *API) AccountByID(
	ctx context.Context,
	sessionID string,
	accountID string,
	opts ...Optional,
) (*Account, *SCAHandler, error) {
	if sessionID == "" {
		return nil, nil, ErrInvalidSessionID
	}
	if accountID == "" {
		return nil, nil, ErrInvalidAccountID
	}
	opts = append(opts, SessionID(sessionID))
	req := a.request(ctx, http.MethodGet, "/ics/v3/accounts/"+accountID, nil, opts...)
	acc := &Account{}
	sca, err := a.do(req, http.StatusOK, acc)
	return acc, sca, err
}
