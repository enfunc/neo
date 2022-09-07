package neo_test

import (
	"context"
	"fmt"

	"github.com/enfunc/neo"
)

func checkSandboxSession(ctx context.Context, api *neo.API, bankID string) (*neo.Session, error) {
	s, err := api.NewSession(ctx, bankID)
	if err != nil {
		return nil, err
	}
	stat, err := api.Status(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	if bankID != stat.BankID {
		return nil, fmt.Errorf("checkSandboxSession: %s != %s", bankID, stat.BankID)
	}
	err = api.DeleteSession(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	return api.NewSession(ctx, bankID)
}
