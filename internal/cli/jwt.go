package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// decodedJWT holds the (unverified) header and claims of a JWT.
type decodedJWT struct {
	Header map[string]any `json:"header"`
	Claims map[string]any `json:"claims"`
}

// decodeJWT base64url-decodes a JWT's header and claims without verifying the
// signature. It mirrors the legacy CLI's --debug behaviour.
func decodeJWT(token string) (decodedJWT, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return decodedJWT{}, fmt.Errorf("not a JWT (expected at least two segments)")
	}
	header, err := decodeSegment(parts[0])
	if err != nil {
		return decodedJWT{}, fmt.Errorf("decode header: %w", err)
	}
	claims, err := decodeSegment(parts[1])
	if err != nil {
		return decodedJWT{}, fmt.Errorf("decode claims: %w", err)
	}
	return decodedJWT{Header: header, Claims: claims}, nil
}

func decodeSegment(seg string) (map[string]any, error) {
	data, err := base64.RawURLEncoding.DecodeString(seg)
	if err != nil {
		// Tolerate padded base64url encodings.
		data, err = base64.URLEncoding.DecodeString(seg)
		if err != nil {
			return nil, err
		}
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// parseTTLSeconds parses a human duration into seconds. It accepts a bare
// integer (seconds), Go durations (e.g. "1h30m"), and a day unit ("10d",
// "1d12h").
func parseTTLSeconds(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	// Bare integer => seconds.
	if n, err := parseAllDigits(s); err == nil {
		return n, nil
	}
	total := 0
	num := strings.Builder{}
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			num.WriteRune(r)
		case r == 'd':
			n, err := parseAllDigits(num.String())
			if err != nil {
				return 0, fmt.Errorf("invalid duration %q", s)
			}
			total += n * 86400
			num.Reset()
		case r == 'h':
			n, err := parseAllDigits(num.String())
			if err != nil {
				return 0, fmt.Errorf("invalid duration %q", s)
			}
			total += n * 3600
			num.Reset()
		case r == 'm':
			n, err := parseAllDigits(num.String())
			if err != nil {
				return 0, fmt.Errorf("invalid duration %q", s)
			}
			total += n * 60
			num.Reset()
		case r == 's':
			n, err := parseAllDigits(num.String())
			if err != nil {
				return 0, fmt.Errorf("invalid duration %q", s)
			}
			total += n
			num.Reset()
		default:
			return 0, fmt.Errorf("invalid duration %q", s)
		}
	}
	if num.Len() > 0 {
		return 0, fmt.Errorf("invalid duration %q (trailing number without unit)", s)
	}
	if total == 0 {
		return 0, fmt.Errorf("duration must be greater than zero")
	}
	return total, nil
}

func parseAllDigits(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("not a number")
		}
		n = n*10 + int(r-'0')
	}
	return n, nil
}
