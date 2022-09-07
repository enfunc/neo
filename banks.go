package neo

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Bool bool

func (b *Bool) UnmarshalJSON(data []byte) error {
	s := string(data)
	if strings.ToLower(s) == `"true"` {
		*b = true
		return nil
	}
	p, err := strconv.ParseBool(s)
	if err != nil {
		return fmt.Errorf("unable to unmarshal %s: %w", data, err)
	}
	*b = Bool(p)
	return nil
}

type Bank struct {
	CountryCode            string   `json:"countryCode"`
	BankingGroupName       string   `json:"bankingGroupName"`
	IdentificationRequired Bool     `json:"personalIdentificationRequired"`
	ID                     string   `json:"id"`
	BankDisplayName        string   `json:"bankDisplayName"`
	SupportedServices      []string `json:"supportedServices"`
	Bic                    string   `json:"bic"`
	BankOfficialName       string   `json:"bankOfficialName"`
	Status                 string   `json:"status"`
}

// Banks returns all available banks.
func (a *API) Banks(ctx context.Context) ([]*Bank, error) {
	req := a.request(ctx, http.MethodGet, "/ics/v3/banks", nil)
	return a.banks(req)
}

// BanksByCountry returns all available banks for the given location.
func (a *API) BanksByCountry(ctx context.Context, countryCode string) ([]*Bank, error) {
	req := a.request(ctx, http.MethodGet, "/ics/v3/banks?countryCode="+url.QueryEscape(countryCode), nil)
	return a.banks(req)
}

// BanksByName returns all available banks with the provided name.
func (a *API) BanksByName(ctx context.Context, name string) ([]*Bank, error) {
	req := a.request(ctx, http.MethodGet, "/ics/v3/banks?name="+url.QueryEscape(name), nil)
	return a.banks(req)
}

func (a *API) banks(req *http.Request) ([]*Bank, error) {
	banks := make([]*Bank, 0, 32)
	_, err := a.do(req, http.StatusOK, &banks)
	return availableOnly(banks), err
}

func availableOnly(banks []*Bank) []*Bank {
	c := make([]*Bank, 0, len(banks))
	for _, b := range banks {
		if b.Status == "AVAILABLE" {
			c = append(c, b)
		}
	}
	return c
}

// BankByID returns a bank with the given ID.
func (a *API) BankByID(ctx context.Context, bankID string) (*Bank, error) {
	if bankID == "" {
		return nil, ErrInvalidBankID
	}
	req := a.request(ctx, http.MethodGet, "/ics/v3/banks/"+bankID, nil)
	bnk := &Bank{}
	_, err := a.do(req, http.StatusOK, bnk)
	return bnk, err
}
