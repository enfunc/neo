package neo_test

import (
	"context"
	"fmt"

	"github.com/enfunc/neo"
)

func checkSandboxSEPA(ctx context.Context, api *neo.API, s *neo.Session, from, to string, opts ...neo.Optional) error {
	pcr, h, err := api.SEPAPayment(ctx, s.ID, &neo.PaymentRequest{
		DebtorName: "Debtor",
		DebtorAccount: &neo.AccountInfo{
			IBAN: from,
		},
		CreditorName: "Creditor",
		CreditorAccount: &neo.AccountInfo{
			IBAN: to,
		},
		RemittanceInfoUnstructured: "test",
		InstrumentedAmount:         "1.00",
		Currency:                   "EUR",
		EndToEndIdentification:     "test",
		PaymentMetadata: &neo.PaymentMetadata{
			Address: &neo.Address{
				StreetName:     "Potetveien",
				BuildingNumber: "15",
				PostalCode:     "0150",
				City:           "Oslo",
				Country:        "Norway",
			},
		},
	}, opts...)
	if err != nil {
		return err
	}
	for h != nil {
		waitForConsent(h.SCA)
		h, err = h.Retry(ctx, pcr)
		if err != nil {
			return err
		}
	}
	if pcr.Status != "ACCP" {
		return fmt.Errorf("checkSandboxSEPA: invalid status %s", pcr.Status)
	}
	return nil
}

func checkSandboxDomestic(ctx context.Context, api *neo.API, s *neo.Session, from, to string, opts ...neo.Optional) error {
	pcr, h, err := api.DomesticPayment(ctx, s.ID, &neo.PaymentRequest{
		DebtorName: "Debtor",
		DebtorAccount: &neo.AccountInfo{
			BBAN: from,
		},
		CreditorName: "Creditor",
		CreditorAccount: &neo.AccountInfo{
			BBAN: to,
		},
		RemittanceInfoUnstructured: "test",
		InstrumentedAmount:         "1.00",
		Currency:                   "NOK",
		EndToEndIdentification:     "test",
		PaymentMetadata: &neo.PaymentMetadata{
			Address: &neo.Address{
				StreetName:     "Potetveien",
				BuildingNumber: "15",
				PostalCode:     "0150",
				City:           "Oslo",
				Country:        "Norway",
			},
		},
	}, opts...)
	if err != nil {
		return err
	}
	for h != nil {
		waitForConsent(h.SCA)
		h, err = h.Retry(ctx, pcr)
		if err != nil {
			return err
		}
	}
	if pcr.Status != "RCVD" {
		return fmt.Errorf("checkSandboxDomestic: invalid status %s", pcr.Status)
	}
	return nil
}
