package useragent

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// recorder captures the User-Agent it is asked to send.
type recorder struct {
	seen  []string
	calls int
}

func (r *recorder) Do(req *http.Request) (*http.Response, error) {
	r.calls++
	r.seen = append(r.seen, req.Header.Get("User-Agent"))
	return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
}

func withVersion(t *testing.T, v string) {
	t.Helper()
	prev := version
	version = v
	t.Cleanup(func() { version = prev })
}

func TestStringUsesVersion(t *testing.T) {
	withVersion(t, "1.17.0")
	if got := String(); got != "gr4vy-cli/1.17.0" {
		t.Errorf("String()=%q", got)
	}
}

func TestSetVersionIgnoresEmpty(t *testing.T) {
	withVersion(t, "1.2.3")
	SetVersion("")
	if got := String(); got != "gr4vy-cli/1.2.3" {
		t.Errorf("empty SetVersion overwrote version: %q", got)
	}
}

func TestDoPrependsToSDKUserAgent(t *testing.T) {
	withVersion(t, "1.17.0")
	rec := &recorder{}
	req := httptest.NewRequest(http.MethodGet, "https://example.test/", nil)
	req.Header.Set("User-Agent", "speakeasy-sdk/go 1.13.0 2.927.0 1.0.0 github.com/gr4vy/gr4vy-go")

	if _, err := WrapClient(rec).Do(req); err != nil {
		t.Fatal(err)
	}

	want := "gr4vy-cli/1.17.0 speakeasy-sdk/go 1.13.0 2.927.0 1.0.0 github.com/gr4vy/gr4vy-go"
	if rec.seen[0] != want {
		t.Errorf("User-Agent=%q want %q", rec.seen[0], want)
	}
}

func TestDoSetsUserAgentWhenAbsent(t *testing.T) {
	withVersion(t, "1.17.0")
	rec := &recorder{}
	req := httptest.NewRequest(http.MethodGet, "https://example.test/", nil)
	req.Header.Del("User-Agent")

	if _, err := WrapClient(rec).Do(req); err != nil {
		t.Fatal(err)
	}

	if rec.seen[0] != "gr4vy-cli/1.17.0" {
		t.Errorf("User-Agent=%q", rec.seen[0])
	}
}

// The SDK's retry loop reuses one *http.Request across attempts, so Do must not
// compound the product token.
func TestDoIsIdempotentAcrossRetries(t *testing.T) {
	withVersion(t, "1.17.0")
	rec := &recorder{}
	c := WrapClient(rec)
	req := httptest.NewRequest(http.MethodGet, "https://example.test/", nil)
	req.Header.Set("User-Agent", "speakeasy-sdk/go 1.13.0 2.927.0 1.0.0 github.com/gr4vy/gr4vy-go")

	for i := 0; i < 3; i++ {
		if _, err := c.Do(req); err != nil {
			t.Fatal(err)
		}
	}

	if rec.calls != 3 {
		t.Fatalf("calls=%d want 3", rec.calls)
	}
	want := "gr4vy-cli/1.17.0 speakeasy-sdk/go 1.13.0 2.927.0 1.0.0 github.com/gr4vy/gr4vy-go"
	for i, got := range rec.seen {
		if got != want {
			t.Errorf("attempt %d User-Agent=%q want %q", i, got, want)
		}
	}
}
