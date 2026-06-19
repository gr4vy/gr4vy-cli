package cli

import "testing"

func TestParseTTLSeconds(t *testing.T) {
	cases := []struct {
		in   string
		want int
		err  bool
	}{
		{"3600", 3600, false},
		{"1h", 3600, false},
		{"30m", 1800, false},
		{"10d", 864000, false},
		{"1d12h", 129600, false},
		{"", 0, true},
		{"5x", 0, true},
		{"1h30", 0, true},
	}
	for _, c := range cases {
		got, err := parseTTLSeconds(c.in)
		if c.err {
			if err == nil {
				t.Errorf("parseTTLSeconds(%q): expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseTTLSeconds(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseTTLSeconds(%q)=%d want %d", c.in, got, c.want)
		}
	}
}

func TestDecodeJWT(t *testing.T) {
	// {"alg":"ES512"} . {"scopes":["buyers.read"]} . sig
	token := "eyJhbGciOiJFUzUxMiJ9.eyJzY29wZXMiOlsiYnV5ZXJzLnJlYWQiXX0.sig"
	dec, err := decodeJWT(token)
	if err != nil {
		t.Fatal(err)
	}
	if dec.Header["alg"] != "ES512" {
		t.Errorf("alg=%v", dec.Header["alg"])
	}
	if _, ok := dec.Claims["scopes"]; !ok {
		t.Errorf("missing scopes claim: %v", dec.Claims)
	}
	if _, err := decodeJWT("notajwt"); err == nil {
		t.Error("expected error for non-JWT")
	}
}
