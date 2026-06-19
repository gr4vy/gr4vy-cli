package clierr

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gr4vy/gr4vy-go/models/apierrors"
)

func TestExitCodeFor(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, ExitOK},
		{"generic", errors.New("boom"), ExitGeneric},
		{"config", Config(errors.New("no key")), ExitConfig},
		{"usage", Usage(errors.New("bad flag")), ExitUsage},
		{"wrapped config", fmt.Errorf("ctx: %w", Config(errors.New("x"))), ExitConfig},
		{"api 404", &apierrors.Error404{}, ExitAPIClient},
		{"api 429", &apierrors.Error429{}, ExitRateLimit},
		{"api 500", &apierrors.Error500{}, ExitAPIServer},
		{"transport 4xx", apierrors.NewAPIError("nope", 403, "", nil), ExitAPIClient},
		{"transport 5xx", apierrors.NewAPIError("nope", 503, "", nil), ExitAPIServer},
	}
	for _, c := range cases {
		if got := ExitCodeFor(c.err); got != c.want {
			t.Errorf("%s: ExitCodeFor=%d want %d", c.name, got, c.want)
		}
	}
}

func TestFormatError(t *testing.T) {
	code := "not_found"
	msg := "Buyer not found"
	e := &apierrors.Error404{Code: &code, Message: &msg}
	if got := FormatError(e); got != "404 not_found — Buyer not found" {
		t.Errorf("FormatError=%q", got)
	}
}
