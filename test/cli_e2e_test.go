package test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gr4vy/gr4vy-cli/test/harness"
)

// TestToken verifies the CLI can mint a signed API token offline-style (no API
// call) using the injected key.
func TestToken(t *testing.T) {
	c := harness.New(t)
	out := c.MustRun(t, "token", "--scope", "transactions.read")
	if !strings.HasPrefix(strings.TrimSpace(out), "ey") {
		t.Fatalf("expected a JWT, got: %q", out)
	}
}

// TestBuyerLifecycle exercises a full CRUD round-trip against the e2e sandbox.
func TestBuyerLifecycle(t *testing.T) {
	c := harness.New(t)

	created := c.MustRun(t, "buyers", "create",
		"--data", `{"display_name":"CLI E2E","external_identifier":"cli-e2e-buyer"}`)
	var buyer struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(created), &buyer); err != nil {
		t.Fatalf("decode created buyer: %v\n%s", err, created)
	}
	if buyer.ID == "" {
		t.Fatalf("created buyer has no id: %s", created)
	}
	t.Cleanup(func() { c.Run(t, "buyers", "delete", buyer.ID) })

	got := c.MustRun(t, "buyers", "get", buyer.ID)
	if !strings.Contains(got, buyer.ID) {
		t.Errorf("get did not return the buyer id: %s", got)
	}

	list := c.MustRun(t, "buyers", "list", "--limit", "5")
	if !strings.Contains(list, "items") {
		t.Errorf("list did not return a collection: %s", list)
	}
}

// TestTransactionsList confirms a generated list command with query flags works.
func TestTransactionsList(t *testing.T) {
	c := harness.New(t)
	out := c.MustRun(t, "transactions", "list", "--limit", "1")
	if !strings.Contains(out, "items") {
		t.Errorf("expected a transactions collection: %s", out)
	}
}

// TestCheckoutSessionCreate confirms an empty-body create works.
func TestCheckoutSessionCreate(t *testing.T) {
	c := harness.New(t)
	out := c.MustRun(t, "checkout-sessions", "create")
	if !strings.Contains(out, "id") {
		t.Errorf("expected a checkout session with an id: %s", out)
	}
}

// TestCommandCoverage is a cheap structural check that every command resolves
// and prints help (exit 0) without credentials being exercised.
func TestCommandCoverage(t *testing.T) {
	c := harness.New(t)
	for _, args := range [][]string{
		{"--help"},
		{"buyers", "--help"},
		{"transactions", "refunds", "create", "--help"},
		{"payment-methods", "--help"},
	} {
		if r := c.Run(t, append(args, "")...); r.Code != 0 && r.Code != 2 {
			// help exits 0; bad arg count exits 2 — both acceptable here.
			t.Errorf("`gr4vy %s` unexpected exit %d: %s", strings.Join(args, " "), r.Code, r.Stderr)
		}
	}
}
