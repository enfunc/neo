package neo_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/enfunc/neo"
)

func checkSandboxAccounts(ctx context.Context, api *neo.API, s *neo.Session, opts ...neo.Optional) ([]*neo.Account, error) {
	acc, sca, err := api.Accounts(ctx, s.ID, opts...)
	if err != nil {
		return nil, err
	}
	if sca != nil {
		waitForConsent(sca.SCA)
		if _, err := sca.Retry(ctx, &acc); err != nil {
			return nil, err
		}
	}
	if len(acc) == 0 {
		return nil, errors.New("checkSandboxAccounts: no accounts")
	}

	fst := acc[0]
	snd, sca, err := api.AccountByID(ctx, s.ID, fst.ID, opts...)
	if err != nil {
		return nil, err
	}
	if sca != nil {
		waitForConsent(sca.SCA)
		if _, err := sca.Retry(ctx, snd); err != nil {
			return nil, err
		}
	}
	if !reflect.DeepEqual(fst, snd) {
		return nil, fmt.Errorf("checkSandboxAccounts: %v != %v", fst, snd)
	}
	return acc, nil
}
