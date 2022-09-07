package neo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type (
	PaymentType string
	PaymentCode string
)

var (
	PaymentTypeDomestic          PaymentType = "domestic-transfer"
	PaymentTypeDomesticScheduled PaymentType = "domestic-scheduled-transfer"
	PaymentTypeSEPA              PaymentType = "sepa-credit"
	PaymentTypeSEPAScheduled     PaymentType = "sepa-scheduled-credit"
)

func (p PaymentType) OK() error {
	for _, t := range []PaymentType{
		PaymentTypeDomestic,
		PaymentTypeDomesticScheduled,
		PaymentTypeSEPAScheduled,
		PaymentTypeSEPA,
	} {
		if p == t {
			return nil
		}
	}
	return ErrInvalidPaymentType
}

var (
	PaymentCodeCommercial PaymentCode = "GDDS"
	PaymentCodeInvoice    PaymentCode = "IVPT"
	PaymentCodeP2P        PaymentCode = "MP2P"
	PaymentCodeOther      PaymentCode = "OTHR"
	PaymentCodeElectronic PaymentCode = "SCVE"
)

var (
	ErrInvalidPaymentRequest        = errors.New("invalid payment request")
	ErrInvalidAccountInfo           = errors.New("invalid account info")
	ErrInvalidRemittanceInformation = errors.New("invalid remittance info")
	ErrInvalidPaymentType           = errors.New("invalid payment type")
	ErrInvalidPaymentID             = errors.New("invalid payment ID")
)

type PaymentRequest struct {
	DebtorAccount              *AccountInfo              `json:"debtorAccount"`
	DebtorName                 string                    `json:"debtorName"`
	CreditorAccount            *AccountInfo              `json:"creditorAccount"`
	CreditorName               string                    `json:"creditorName"`
	RemittanceInfoUnstructured string                    `json:"remittanceInformationUnstructured,omitempty"`
	RemittanceInfoStructured   *RemittanceInfoStructured `json:"remittanceInformationStructured,omitempty"`
	InstrumentedAmount         string                    `json:"instrumentedAmount"`
	Currency                   string                    `json:"currency"`
	EndToEndIdentification     string                    `json:"endToEndIdentification"`
	PaymentMetadata            *PaymentMetadata          `json:"paymentMetadata,omitempty"`
	RequestedExecutionDate     *time.Time                `json:"requestedExecutionDate,omitempty"`
}

func (r *PaymentRequest) OK() error {
	if r == nil {
		return ErrInvalidPaymentRequest
	}
	for _, s := range []string{
		r.DebtorName,
		r.CreditorName,
		r.InstrumentedAmount,
		r.Currency,
		r.EndToEndIdentification,
	} {
		if s == "" {
			return ErrInvalidPaymentRequest
		}
	}
	if r.RemittanceInfoUnstructured == "" {
		if err := r.RemittanceInfoStructured.OK(); err != nil {
			return err
		}
	} else {
		if r.RemittanceInfoStructured != nil {
			return ErrInvalidPaymentRequest
		}
	}
	if err := r.DebtorAccount.OK(); err != nil {
		return err
	}
	if err := r.CreditorAccount.OK(); err != nil {
		return err
	}
	return nil
}

type AccountInfo struct {
	BBAN                  string `json:"bban,omitempty"`
	IBAN                  string `json:"iban,omitempty"`
	SortCodeAccountNumber string `json:"sortCodeAccountNumber,omitempty"`
}

func (info *AccountInfo) OK() error {
	if info == nil || (info.BBAN == "" && info.IBAN == "" && info.SortCodeAccountNumber == "") {
		return ErrInvalidAccountInfo
	}
	return nil
}

type RemittanceInfoStructured struct {
	Reference string `json:"reference,omitempty"`
	Issuer    string `json:"referenceIssuer,omitempty"`
	Type      string `json:"referenceType,omitempty"`
}

func (info *RemittanceInfoStructured) OK() error {
	if info == nil || (info.Reference == "" && info.Issuer == "" && info.Type == "") {
		return ErrInvalidRemittanceInformation
	}
	return nil
}

type Address struct {
	StreetName     string `json:"streetName"`
	BuildingNumber string `json:"buildingNumber"`
	PostalCode     string `json:"postalCode"`
	City           string `json:"city"`
	Country        string `json:"country"`
}

type Agent struct {
	ID     string `json:"identification"`
	IDType string `json:"identificationType"`
}

type PaymentMetadata struct {
	Address              *Address    `json:"creditorAddress,omitempty"`
	Agent                *Agent      `json:"creditorAgent,omitempty"`
	Code                 PaymentCode `json:"paymentContextCode,omitempty"`
	MerchantCategoryCode string      `json:"merchantCategoryCode,omitempty"`
	MerchantCustomerID   string      `json:"merchantCustomerIdentification,omitempty"`
}

type PaymentCreated struct {
	PaymentID string    `json:"paymentId"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"creationDateTime"`
}

//nolint:cyclop
func (a *API) payment(
	ctx context.Context,
	sessionID string,
	paymentType PaymentType,
	r *PaymentRequest,
	opts ...Optional,
) (*PaymentCreated, *SCAHandler, error) {
	if sessionID == "" {
		return nil, nil, ErrInvalidSessionID
	}
	if err := paymentType.OK(); err != nil {
		return nil, nil, err
	}
	if err := r.OK(); err != nil {
		return nil, nil, err
	}
	prq, err := json.Marshal(r)
	if err != nil {
		return nil, nil, fmt.Errorf("neo: marshaling *PaymentRequest failed: %w", err)
	}
	opts = append(opts, SessionID(sessionID))
	req := a.request(ctx, http.MethodPost, "/ics/v3/payments/"+string(paymentType), bytes.NewReader(prq), opts...)
	pcr := &PaymentCreated{}
	sca, err := a.do(req, http.StatusCreated, pcr)
	if err != nil {
		return nil, nil, err
	}
	if sca != nil { //nolint:nestif
		r := sca.Retry
		sca.Retry = func(ctx context.Context, v interface{}) (*SCAHandler, error) {
			sca, err := r(ctx, v)
			if err != nil {
				return nil, err
			}
			if sca != nil && sca.Error.IsPaymentAuthError() {
				// Since this is a payment authorization request, the retry should be a call to CompletePayment.
				sca.Retry = func(ctx context.Context, u interface{}) (*SCAHandler, error) {
					v, sca, err := a.CompletePayment(ctx, sessionID, paymentType, sca.ID, opts...)
					if err != nil {
						return nil, err
					}
					if sca != nil {
						return sca, nil
					}
					switch t := u.(type) {
					case *PaymentCreated:
						t.PaymentID = v.PaymentID
						t.CreatedAt = v.CreatedAt
						t.Status = v.Status
						return nil, nil
					default:
						return nil, fmt.Errorf("expected *PaymentCreated, but got: %v", u)
					}
				}
			}
			return sca, nil
		}
	}
	return pcr, sca, nil
}

func (a *API) CompletePayment(
	ctx context.Context,
	sessionID string,
	paymentType PaymentType,
	paymentID string,
	opts ...Optional,
) (*PaymentCreated, *SCAHandler, error) {
	if sessionID == "" {
		return nil, nil, ErrInvalidSessionID
	}
	if err := paymentType.OK(); err != nil {
		return nil, nil, err
	}
	if paymentID == "" {
		return nil, nil, ErrInvalidPaymentID
	}
	opts = append(opts, SessionID(sessionID))
	uri := fmt.Sprintf("/ics/v3/payments/%s/%s/complete", paymentType, paymentID)
	req := a.request(ctx, http.MethodPost, uri, nil, opts...)
	pcr := &PaymentCreated{}
	sca, err := a.do(req, http.StatusCreated, pcr)
	return pcr, sca, err
}

func (a *API) SEPAPayment(
	ctx context.Context,
	sessionID string,
	r *PaymentRequest,
	opts ...Optional,
) (*PaymentCreated, *SCAHandler, error) {
	return a.payment(ctx, sessionID, PaymentTypeSEPA, r, opts...)
}

func (a *API) SEPAScheduledPayment(
	ctx context.Context,
	sessionID string,
	r *PaymentRequest,
	opts ...Optional,
) (*PaymentCreated, *SCAHandler, error) {
	if r == nil || r.RequestedExecutionDate == nil {
		return nil, nil, ErrInvalidPaymentRequest
	}
	return a.payment(ctx, sessionID, PaymentTypeSEPAScheduled, r, opts...)
}

func (a *API) DomesticPayment(
	ctx context.Context,
	sessionID string,
	r *PaymentRequest,
	opts ...Optional,
) (*PaymentCreated, *SCAHandler, error) {
	return a.payment(ctx, sessionID, PaymentTypeDomestic, r, opts...)
}

func (a *API) DomesticScheduledPayment(
	ctx context.Context,
	sessionID string,
	r *PaymentRequest,
	opts ...Optional,
) (*PaymentCreated, *SCAHandler, error) {
	if r == nil || r.RequestedExecutionDate == nil {
		return nil, nil, ErrInvalidPaymentRequest
	}
	return a.payment(ctx, sessionID, PaymentTypeDomesticScheduled, r, opts...)
}

func (a *API) AuthorizePayment(
	ctx context.Context,
	sessionID string,
	paymentType PaymentType,
	paymentID string,
	opts ...Optional,
) (*SCA, error) {
	if sessionID == "" {
		return nil, ErrInvalidSessionID
	}
	if err := paymentType.OK(); err != nil {
		return nil, err
	}
	if paymentID == "" {
		return nil, ErrInvalidPaymentID
	}
	opts = append(opts, SessionID(sessionID))
	uri := fmt.Sprintf("/ics/v3/payments/%s/%s/authorize", paymentType, paymentID)
	req := a.request(ctx, http.MethodGet, uri, nil, opts...)
	c := &Consent{}
	if _, err := a.do(req, http.StatusOK, c); err != nil {
		return nil, err
	}
	return &SCA{
		URL: c.Links[0].Href,
		ID:  c.PaymentID,
	}, nil
}
