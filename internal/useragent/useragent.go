// Package useragent identifies CLI traffic on the wire. Requests made through
// the embedded gr4vy-go SDK are otherwise indistinguishable from any other Go
// SDK integration, so the CLI prepends its own product token to the SDK's
// User-Agent.
//
// It is a leaf package (net/http only) so both internal/auth and internal/cli
// can depend on it.
package useragent

import (
	"net/http"
	"strings"
)

// version is the CLI version, set once from main's build metadata.
var version = "dev"

// SetVersion records the running CLI version. It is called by cli.Execute
// before any command runs.
func SetVersion(v string) {
	if v != "" {
		version = v
	}
}

// String is the CLI's product token, e.g. "gr4vy-cli/1.17.0".
func String() string {
	return "gr4vy-cli/" + version
}

// doer is the subset of *http.Client the SDK needs. It matches gr4vy-go's
// exported HTTPClient interface structurally, so a Client satisfies both.
type doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client wraps a doer and prepends the CLI's product token to every outgoing
// User-Agent.
type Client struct {
	inner doer
}

// WrapClient returns inner with CLI identification applied to each request.
func WrapClient(inner doer) Client {
	return Client{inner: inner}
}

// Do prepends the CLI product token to the User-Agent set by the SDK, keeping
// the SDK and generator versions that follow it. Ordering matches RFC 9110:
// most significant product first.
//
// Do is idempotent. The SDK's retry loop reuses a single *http.Request across
// attempts, so an unconditional prepend would compound the token on retries.
func (c Client) Do(req *http.Request) (*http.Response, error) {
	prefix := String()
	if ua := req.Header.Get("User-Agent"); !strings.HasPrefix(ua, prefix) {
		req.Header.Set("User-Agent", strings.TrimSpace(prefix+" "+ua))
	}
	return c.inner.Do(req)
}
