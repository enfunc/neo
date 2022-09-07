package neo_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/enfunc/neo"
)

var (
	client     = ""
	secret     = ""
	encKeyPath = ""

	bankDNB     = "RG5iLm5vLnYxRE5CQU5PS0s="
	bankHizonti = "aGl6b250aWJhbmsudjFISVpPTk9LSw=="
	bankSbanken = "U2Jhbmtlbi52MVNCQUtOT0JC"
)

func TestMain(m *testing.M) {
	client = os.Getenv("NEO_CLIENT")
	if client == "" {
		os.Exit(1)
		return
	}
	secret = os.Getenv("NEO_SECRET")
	if secret == "" {
		os.Exit(1)
		return
	}
	encKeyPath = os.Getenv("NEO_KEY_PATH")
	if encKeyPath == "" {
		os.Exit(1)
		return
	}
	os.Exit(m.Run())
}

func sandboxClient() *neo.Client {
	return neo.NewSandboxClient(client, secret, http.DefaultClient)
}

func sandboxAPI(ctx context.Context) *neo.API {
	c := sandboxClient()
	a, err := c.API(ctx, "test-device-id")
	if err != nil {
		panic(err)
	}
	return a
}

func waitForConsent(sca *neo.SCA) {
	fmt.Println(sca.URL)
	time.Sleep(15 * time.Second)
}

func TestSandboxDNBAccounts(t *testing.T) {
	cli := sandboxClient()
	ctx := context.TODO()

	if err := checkSandboxAuth(ctx, cli); err != nil {
		t.Fatal(err)
		return
	}

	api, err := cli.API(ctx, "test-device-id")
	if err != nil {
		t.Fatal(err)
		return
	}

	if err := checkSandboxBanks(ctx, api); err != nil {
		t.Fatal(err)
		return
	}

	session, err := checkSandboxSession(ctx, api, bankDNB)
	if err != nil {
		t.Fatal(err)
		return
	}

	encrypt, err := neo.NewEncrypter(encKeyPath)
	if err != nil {
		t.Fatal(err)
		return
	}

	// See https://docs.neonomics.io/documentation/development/testing-in-sandbox.
	ssid, err := encrypt("31125461118")
	if err != nil {
		t.Fatal(err)
		return
	}

	accounts, err := checkSandboxAccounts(ctx, api, session, neo.PsuID(ssid))
	if err != nil {
		t.Fatal(err)
		return
	}

	if err := checkSandboxTxs(ctx, api, session, accounts, neo.PsuID(ssid)); err != nil {
		t.Fatal(err)
		return
	}
}

func TestSandboxHizontiSEPA(t *testing.T) {
	cli := sandboxClient()
	ctx := context.TODO()

	api, err := cli.API(ctx, "test-device-id")
	if err != nil {
		t.Fatal(err)
		return
	}

	session, err := api.NewSession(ctx, bankHizonti)
	if err != nil {
		t.Fatal(err)
		return
	}

	hizontiIBAN := "NO7013086520592"
	justoIBAN := "NO5004947573768"
	if err := checkSandboxSEPA(ctx, api, session, hizontiIBAN, justoIBAN); err != nil {
		t.Fatal(err)
		return
	}
}

func TestSandboxSbankenDomestic(t *testing.T) {
	cli := sandboxClient()
	ctx := context.TODO()

	api, err := cli.API(ctx, "test-device-id")
	if err != nil {
		t.Fatal(err)
		return
	}

	session, err := api.NewSession(ctx, bankSbanken)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Sample SSN: 01078900497
	if err := checkSandboxDomestic(
		ctx,
		api,
		session,
		"90412263056",
		"90522037388",
		neo.PsuIP("109.74.179.3"),
	); err != nil {
		t.Fatal(err)
		return
	}
}
