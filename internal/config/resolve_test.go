package config

import "testing"

func envMap(m map[string]string) EnvLookup {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func TestResolveProfileSelection(t *testing.T) {
	c := &Config{
		ActiveProfile: "prod",
		Profiles: map[string]Profile{
			"prod":    {ID: "acme", Environment: EnvProduction},
			"sandbox": {ID: "acme", Environment: EnvSandbox},
		},
	}

	t.Run("flag wins over active", func(t *testing.T) {
		r, err := Resolve(c, Overrides{Profile: "sandbox"}, envMap(nil))
		if err != nil {
			t.Fatal(err)
		}
		if r.ProfileName != "sandbox" {
			t.Fatalf("got %q want sandbox", r.ProfileName)
		}
	})

	t.Run("env wins over active", func(t *testing.T) {
		r, err := Resolve(c, Overrides{}, envMap(map[string]string{EnvProfile: "sandbox"}))
		if err != nil {
			t.Fatal(err)
		}
		if r.ProfileName != "sandbox" {
			t.Fatalf("got %q want sandbox", r.ProfileName)
		}
	})

	t.Run("falls back to active", func(t *testing.T) {
		r, err := Resolve(c, Overrides{}, envMap(nil))
		if err != nil {
			t.Fatal(err)
		}
		if r.ProfileName != "prod" {
			t.Fatalf("got %q want prod", r.ProfileName)
		}
	})
}

func TestResolveFieldPrecedence(t *testing.T) {
	c := &Config{
		ActiveProfile: "p",
		Profiles: map[string]Profile{
			"p": {ID: "fromprofile", Environment: EnvSandbox, MerchantAccountID: "ma-profile"},
		},
	}
	env := envMap(map[string]string{
		EnvID:                "fromenv",
		EnvMerchantAccountID: "ma-env",
	})
	r, err := Resolve(c, Overrides{ID: "fromflag"}, env)
	if err != nil {
		t.Fatal(err)
	}
	if r.Profile.ID != "fromflag" {
		t.Errorf("id: flag should win, got %q", r.Profile.ID)
	}
	if r.Profile.MerchantAccountID != "ma-env" {
		t.Errorf("mid: env should win over profile, got %q", r.Profile.MerchantAccountID)
	}
}

func TestResolveDefaults(t *testing.T) {
	c := &Config{Profiles: map[string]Profile{}}
	r, err := Resolve(c, Overrides{ID: "acme"}, envMap(nil))
	if err != nil {
		t.Fatal(err)
	}
	if r.Profile.Environment != EnvSandbox {
		t.Errorf("environment default should be sandbox, got %q", r.Profile.Environment)
	}
	if r.Profile.AuthMethod != AuthKey {
		t.Errorf("auth_method default should be key, got %q", r.Profile.AuthMethod)
	}
}

func TestResolveInvalidEnvironment(t *testing.T) {
	c := &Config{Profiles: map[string]Profile{}}
	if _, err := Resolve(c, Overrides{Environment: "staging"}, envMap(nil)); err == nil {
		t.Fatal("expected error for invalid environment")
	}
}
