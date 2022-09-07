package neo_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/enfunc/neo"
)

func TestSandboxBanks(t *testing.T) {
	ctx := context.TODO()
	if err := checkSandboxBanks(ctx, sandboxAPI(ctx)); err != nil {
		t.Fatal(err)
	}
}

func checkSandboxBanks(ctx context.Context, api *neo.API) error {
	banks, err := api.Banks(ctx)
	if err != nil {
		return err
	}
	if len(banks) == 0 {
		return errors.New("checkSandboxBanks: no banks")
	}

	fst := banks[0]
	snd, err := api.BankByID(ctx, fst.ID)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(fst, snd) {
		return fmt.Errorf("checkSandboxBanks: %v != %v", fst, snd)
	}

	trd, err := api.BanksByName(ctx, fst.BankDisplayName)
	if err != nil {
		return err
	}
	if len(trd) != 1 {
		return errors.New("checkSandboxBanks: bank by an existing name not found")
	}
	if !reflect.DeepEqual(fst, trd[0]) {
		return fmt.Errorf("checkSandboxBanks: %v != %v", fst, trd[0])
	}

	fth, err := api.BanksByCountry(ctx, "NO")
	if err != nil {
		return err
	}
	if len(fth) == 0 {
		return errors.New("checkSandboxBanks: no banks by country")
	}
	return nil
}
