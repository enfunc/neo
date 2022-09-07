package neo_test

import (
	"context"
	"errors"

	"github.com/enfunc/neo"
)

func checkSandboxTxs(ctx context.Context, api *neo.API, s *neo.Session, acc []*neo.Account, opts ...neo.Optional) error {
	for _, a := range acc {
		txs, sca, err := api.Txs(ctx, s.ID, a.ID, opts...)
		if err != nil {
			return err
		}
		if sca != nil {
			waitForConsent(sca.SCA)
			if _, err := sca.Retry(ctx, &txs); err != nil {
				return err
			}
		}
		if len(txs) == 0 {
			return errors.New("checkSandboxTxs: no txs")
		}
	}
	return nil
}
