package neo

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Tx struct {
	ID                   string     `json:"id"`
	TransactionReference string     `json:"transactionReference"`
	TransactionAmount    *Money     `json:"transactionAmount"`
	CreditDebitIndicator string     `json:"creditDebitIndicator"`
	BookingDate          *time.Time `json:"bookingDate"`
	ValueDate            *time.Time `json:"valueDate"`
	CounterpartyAccount  string     `json:"counterpartyAccount"`
	CounterpartyName     string     `json:"counterpartyName"`
	CounterpartyAgent    string     `json:"counterpartyAgent"`
}

type Money struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

// Txs returns a list of transactions for the given account.
func (a *API) Txs(ctx context.Context, sessionID, accountID string, opts ...Optional) ([]*Tx, *SCAHandler, error) {
	if sessionID == "" {
		return nil, nil, ErrInvalidSessionID
	}
	if accountID == "" {
		return nil, nil, ErrInvalidAccountID
	}
	opts = append(opts, SessionID(sessionID))
	uri := fmt.Sprintf("/ics/v3/accounts/%s/transactions", accountID)
	req := a.request(ctx, http.MethodGet, uri, nil, opts...)
	txs := make([]*Tx, 0, 32)
	sca, err := a.do(req, http.StatusOK, &txs)
	return txs, sca, err
}
