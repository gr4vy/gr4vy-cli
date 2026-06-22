package auth

import (
	"testing"

	gr4vygo "github.com/gr4vy/gr4vy-go"

	"github.com/gr4vy/gr4vy-cli/internal/config"
)

func TestParseScopesAliasesAndValidation(t *testing.T) {
	got, err := ParseScopes([]string{"all.read", "buyers.write", "*.read"})
	if err != nil {
		t.Fatal(err)
	}
	want := []gr4vygo.JWTScope{gr4vygo.ReadAll, gr4vygo.BuyersWrite, gr4vygo.ReadAll}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("scope[%d]=%q want %q", i, got[i], want[i])
		}
	}

	if _, err := ParseScopes([]string{"buyers.reed"}); err == nil {
		t.Error("expected error for misspelled scope")
	}
}

func TestAPIBaseURL(t *testing.T) {
	cases := []struct {
		id, env, want string
	}{
		{"acme", config.EnvSandbox, "https://api.sandbox.acme.gr4vy.app"},
		{"acme", config.EnvProduction, "https://api.acme.gr4vy.app"},
	}
	for _, c := range cases {
		if got := APIBaseURL(c.id, c.env); got != c.want {
			t.Errorf("APIBaseURL(%q,%q)=%q want %q", c.id, c.env, got, c.want)
		}
	}
}
