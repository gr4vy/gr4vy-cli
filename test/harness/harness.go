// Package harness drives the compiled gr4vy binary against the shared e2e
// sandbox (api.sandbox.e2e.gr4vy.app), mirroring how the gr4vy SDKs run their
// live test suites. It authenticates with the PRIVATE_KEY secret and skips when
// no credentials are available (e.g. fork PRs / local dev without a key).
package harness

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

const (
	// ServerID is the shared e2e instance id; with server=sandbox it resolves
	// to https://api.sandbox.e2e.gr4vy.app.
	ServerID = "e2e"
	Server   = "sandbox"
)

var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

// CLI runs the built binary with e2e configuration injected via the
// environment (no config file needed).
type CLI struct {
	bin string
	key string
	mid string
}

// New builds the binary (once) and loads credentials, skipping the test if none
// are available.
func New(t *testing.T) *CLI {
	t.Helper()
	key := loadKey(t)
	bin := buildBinary(t)
	mid := getenvDefault("GR4VY_E2E_MERCHANT_ACCOUNT_ID", "default")
	return &CLI{bin: bin, key: key, mid: mid}
}

// Result is the outcome of a CLI invocation.
type Result struct {
	Stdout string
	Stderr string
	Code   int
}

// Run executes the binary with the given args and returns its output.
func (c *CLI) Run(t *testing.T, args ...string) Result {
	t.Helper()
	cmd := exec.Command(c.bin, args...)
	cmd.Env = append(os.Environ(),
		"GR4VY_ID="+ServerID,
		"GR4VY_SERVER="+Server,
		"GR4VY_MERCHANT_ACCOUNT_ID="+c.mid,
		"GR4VY_PRIVATE_KEY="+c.key,
		"GR4VY_OUTPUT=json",
		"GR4VY_SECRET_BACKEND=file",
		"HOME="+t.TempDir(),
		// Isolate config/secret backends: both honor XDG_CONFIG_HOME ahead of
		// HOME, so pin it to a temp dir to keep the subprocess hermetic even
		// when the host sets XDG_CONFIG_HOME (common on Linux).
		"XDG_CONFIG_HOME="+t.TempDir(),
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run %v: %v", args, err)
	}
	return Result{Stdout: stdout.String(), Stderr: stderr.String(), Code: code}
}

// MustRun runs the command and fails the test if the exit code is non-zero.
func (c *CLI) MustRun(t *testing.T, args ...string) string {
	t.Helper()
	r := c.Run(t, args...)
	if r.Code != 0 {
		t.Fatalf("`gr4vy %s` failed (exit %d): %s", strings.Join(args, " "), r.Code, r.Stderr)
	}
	return r.Stdout
}

func loadKey(t *testing.T) string {
	if v := os.Getenv("PRIVATE_KEY"); v != "" {
		return v
	}
	if root := moduleRoot(); root != "" {
		if data, err := os.ReadFile(filepath.Join(root, "private_key.pem")); err == nil {
			return string(data)
		}
	}
	t.Skip("no PRIVATE_KEY env var or private_key.pem; skipping live e2e suite")
	return ""
}

func buildBinary(t *testing.T) string {
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "gr4vy-e2e")
		if err != nil {
			buildErr = err
			return
		}
		out := filepath.Join(dir, "gr4vy")
		cmd := exec.Command("go", "build", "-o", out, "github.com/gr4vy/gr4vy-cli")
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			buildErr = err
			return
		}
		binPath = out
	})
	if buildErr != nil {
		t.Fatalf("build binary: %v", buildErr)
	}
	return binPath
}

// moduleRoot walks up from this file to the directory containing go.mod.
func moduleRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	dir := filepath.Dir(file)
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
