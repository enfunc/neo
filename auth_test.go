package neo_test

import (
	"context"
	"testing"

	"github.com/enfunc/neo"
)

func TestSandboxAuth(t *testing.T) {
	if err := checkSandboxAuth(context.TODO(), sandboxClient()); err != nil {
		t.Fatal(err)
	}
}

func checkSandboxAuth(ctx context.Context, c *neo.Client) error {
	token, err := c.AccessToken(ctx)
	if err != nil {
		return err
	}
	_, err = c.RefreshToken(ctx, token)
	return err
}
