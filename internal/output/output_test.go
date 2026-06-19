package output

import (
	"bytes"
	"strings"
	"testing"
)

func render(t *testing.T, v any, f Format) string {
	t.Helper()
	var b bytes.Buffer
	if err := Render(&b, v, f, Options{}); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func TestParse(t *testing.T) {
	if f, _ := Parse("", true); f != Table {
		t.Errorf("empty+TTY should be table, got %q", f)
	}
	if f, _ := Parse("", false); f != JSON {
		t.Errorf("empty+non-TTY should be json, got %q", f)
	}
	if _, err := Parse("xml", true); err == nil {
		t.Error("xml should be rejected")
	}
}

func TestRenderJSON(t *testing.T) {
	out := render(t, map[string]any{"id": "buyer_1", "amount": 1299}, JSON)
	if !strings.Contains(out, `"id": "buyer_1"`) {
		t.Errorf("json missing id: %s", out)
	}
}

func TestRenderTableList(t *testing.T) {
	v := map[string]any{"items": []any{
		map[string]any{"id": "tx_1", "status": "capture_pending", "amount": float64(1299)},
		map[string]any{"id": "tx_2", "status": "captured", "amount": float64(50)},
	}}
	out := render(t, v, Table)
	if !strings.Contains(out, "ID") || !strings.Contains(out, "STATUS") {
		t.Errorf("missing header: %s", out)
	}
	if !strings.Contains(out, "tx_1") || !strings.Contains(out, "1299") {
		t.Errorf("missing rows: %s", out)
	}
	// id should appear before status (preferred ordering).
	if strings.Index(out, "ID") > strings.Index(out, "STATUS") {
		t.Errorf("expected id column before status: %s", out)
	}
}

func TestRenderTableEmpty(t *testing.T) {
	out := render(t, map[string]any{"items": []any{}}, Table)
	if !strings.Contains(out, "no results") {
		t.Errorf("expected no-results message: %s", out)
	}
}

func TestScalarIntFormatting(t *testing.T) {
	if got := scalar(float64(1299)); got != "1299" {
		t.Errorf("got %q want 1299", got)
	}
	if got := scalar(float64(12.5)); got != "12.5" {
		t.Errorf("got %q want 12.5", got)
	}
	if got := scalar([]any{1, 2, 3}); got != "[3]" {
		t.Errorf("got %q want [3]", got)
	}
}
