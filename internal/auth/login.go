package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	gr4vygo "github.com/gr4vy/gr4vy-go"

	"github.com/gr4vy/gr4vy-cli/internal/config"
	"github.com/gr4vy/gr4vy-cli/internal/secret"
)

// sessionPath is the (internal, unpublished) endpoint for email/password
// sessions. It is served from the same per-merchant API host.
const sessionPath = "/auth/sessions"

// refreshSkew refreshes the access token slightly before it actually expires.
const refreshSkew = 60 * time.Second

// ReauthError indicates the stored session is missing or unrefreshable and the
// user must log in again. Err, when set, carries the underlying cause (e.g. a
// network or 5xx failure during refresh) so transient problems stay diagnosable.
type ReauthError struct {
	Profile string
	Err     error
}

func (e *ReauthError) Error() string {
	msg := fmt.Sprintf("not logged in (or session expired) for profile %q; run `gr4vy login`", e.Profile)
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}

func (e *ReauthError) Unwrap() error { return e.Err }

// storedSession is the persisted login bundle.
type storedSession struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// sessionResponse is the /auth/sessions response body.
type sessionResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresIn    float64 `json:"expires_in"` // seconds
	TokenType    string  `json:"token_type"`
}

// LoginTokenProvider serves a stored access token, refreshing it via the
// refresh token when it nears expiry. All /auth/sessions HTTP lives here so the
// rest of the CLI stays agnostic to this internal endpoint.
type LoginTokenProvider struct {
	ProfileName string
	Store       secret.Store
	AuthHost    string
	HTTPClient  *http.Client
	env         config.EnvLookup
	now         func() time.Time
}

// NewLoginTokenProvider constructs a login provider for the resolved profile.
func NewLoginTokenProvider(r config.Resolved, store secret.Store, env config.EnvLookup) (*LoginTokenProvider, error) {
	host, err := authHost(r)
	if err != nil {
		return nil, err
	}
	if env == nil {
		env = config.OSEnv
	}
	return &LoginTokenProvider{
		ProfileName: r.ProfileName,
		Store:       store,
		AuthHost:    host,
		HTTPClient:  http.DefaultClient,
		env:         env,
		now:         time.Now,
	}, nil
}

// Scopes returns nil: login session scopes are server-defined by the user's role.
func (p *LoginTokenProvider) Scopes() []gr4vygo.JWTScope { return nil }

func (p *LoginTokenProvider) Token(ctx context.Context) (string, error) {
	if v, ok := p.env(EnvAccessToken); ok && v != "" {
		return v, nil // CI short-circuit; no refresh
	}
	s, err := p.load()
	if err != nil {
		return "", err
	}
	if p.now().Add(refreshSkew).Before(s.ExpiresAt) {
		return s.AccessToken, nil
	}
	refreshed, err := p.refresh(ctx, s.RefreshToken)
	if err != nil {
		return "", &ReauthError{Profile: p.ProfileName, Err: err}
	}
	if err := p.save(refreshed); err != nil {
		return "", err
	}
	return refreshed.AccessToken, nil
}

func (p *LoginTokenProvider) load() (storedSession, error) {
	var s storedSession
	raw, err := p.Store.Get(p.ProfileName, secret.KindLogin)
	if errors.Is(err, secret.ErrNotFound) {
		return s, &ReauthError{Profile: p.ProfileName}
	}
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return s, fmt.Errorf("corrupt stored session: %w", err)
	}
	return s, nil
}

func (p *LoginTokenProvider) save(s storedSession) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return p.Store.Set(p.ProfileName, secret.KindLogin, string(data))
}

func (p *LoginTokenProvider) refresh(ctx context.Context, refreshToken string) (storedSession, error) {
	resp, err := doSession(ctx, p.HTTPClient, http.MethodPut, p.AuthHost, nil, refreshToken)
	if err != nil {
		return storedSession{}, err
	}
	return p.fromResponse(resp), nil
}

func (p *LoginTokenProvider) fromResponse(r sessionResponse) storedSession {
	return storedSession{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		ExpiresAt:    p.now().Add(time.Duration(r.ExpiresIn) * time.Second),
		TokenType:    r.TokenType,
	}
}

// Login performs an email/password login and stores the resulting session.
func Login(ctx context.Context, r config.Resolved, store secret.Store, email, password string, client *http.Client) error {
	host, err := authHost(r)
	if err != nil {
		return err
	}
	if client == nil {
		client = http.DefaultClient
	}
	body := map[string]string{"email_address": email, "password": password}
	resp, err := doSession(ctx, client, http.MethodPost, host, body, "")
	if err != nil {
		return err
	}
	p := &LoginTokenProvider{ProfileName: r.ProfileName, Store: store, now: time.Now}
	return p.save(p.fromResponse(resp))
}

// Logout clears the stored session and best-effort revokes it server-side.
func Logout(ctx context.Context, r config.Resolved, store secret.Store, client *http.Client) error {
	if client == nil {
		client = http.DefaultClient
	}
	if host, err := authHost(r); err == nil {
		if raw, gerr := store.Get(r.ProfileName, secret.KindLogin); gerr == nil {
			var s storedSession
			if json.Unmarshal([]byte(raw), &s) == nil && s.AccessToken != "" {
				_, _ = doSession(ctx, client, http.MethodDelete, host, nil, s.AccessToken)
			}
		}
	}
	err := store.Delete(r.ProfileName, secret.KindLogin)
	if errors.Is(err, secret.ErrNotFound) {
		return nil
	}
	return err
}

// doSession issues a request to {host}/auth/sessions and decodes the session
// response. bearer, when set, is sent as the Authorization header (the access
// token for DELETE, the refresh token for PUT). DELETE returns 204 with no body.
func doSession(ctx context.Context, client *http.Client, method, host string, body map[string]string, bearer string) (sessionResponse, error) {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return sessionResponse{}, err
		}
		buf = bytes.NewReader(b)
	}
	url := strings.TrimRight(host, "/") + sessionPath
	req, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		return sessionResponse{}, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := client.Do(req)
	if err != nil {
		return sessionResponse{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return sessionResponse{}, fmt.Errorf("auth/sessions %s failed: %s: %s", method, resp.Status, strings.TrimSpace(string(respBody)))
	}
	if resp.StatusCode == http.StatusNoContent || len(respBody) == 0 {
		return sessionResponse{}, nil
	}
	var sr sessionResponse
	if err := json.Unmarshal(respBody, &sr); err != nil {
		return sessionResponse{}, fmt.Errorf("decode session response: %w", err)
	}
	return sr, nil
}

// authHost returns the base host for /auth/sessions: an explicit auth_host, or
// the API host derived from the instance id + environment.
func authHost(r config.Resolved) (string, error) {
	if r.Profile.AuthHost != "" {
		return r.Profile.AuthHost, nil
	}
	if r.Profile.ID == "" {
		return "", fmt.Errorf("cannot determine auth host: set --id/GR4VY_ID or auth_host on the profile")
	}
	return APIBaseURL(r.Profile.ID, r.Profile.Environment), nil
}
