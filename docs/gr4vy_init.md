## gr4vy init

Create your first profile interactively

### Synopsis

Bootstrap a configuration profile. Prompts interactively on a terminal; use flags (and --no-input) for scripted setup.

```
gr4vy init [name] [flags]
```

### Options

```
      --auth-host string             override the auth host for login
      --auth-method string           key or login (default key)
      --default-scope strings        default token scope (repeatable)
      --email string                 login email (for auth-method=login)
      --environment string           sandbox or production
  -h, --help                         help for init
      --id string                    Gr4vy instance id
      --key-env string               name of an env var holding the PEM private key
      --key-file string              path to a PEM private key to import into the secret store
      --key-path string              reference a PEM file in place (not copied into the store)
      --key-stdin                    read the PEM private key from stdin
      --merchant-account-id string   default merchant account id
      --no-input                     never prompt; fail if a required field is missing
      --set-active                   make this the active profile
      --token-ttl string             default token lifetime, e.g. 1h
```

### Options inherited from parent commands

```
      --compact            compact single-line JSON output
      --config string      path to the config file (env: GR4VY_CONFIG)
      --debug              print debug information to stderr
  -o, --output string      output format: json|yaml|table (env: GR4VY_OUTPUT)
      --profile string     configuration profile to use (env: GR4VY_PROFILE)
      --server string      server environment: sandbox|production (env: GR4VY_SERVER)
      --timeout duration   per-request timeout, e.g. 30s
      --token string       pre-generated bearer token; skips JWT signing (env: GR4VY_TOKEN)
```

### SEE ALSO

* [gr4vy](gr4vy.md)	 - The Gr4vy CLI

