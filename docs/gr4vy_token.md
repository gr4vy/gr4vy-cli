## gr4vy token

Generate a server-to-server API access token (JWT)

### Synopsis

Generate a signed bearer token for the Gr4vy API using the profile's private key. Scopes default to the profile's default_scopes (or *.read and *.write).

```
gr4vy token [flags]
```

### Options

```
  -e, --expires-in string   token lifetime, e.g. 1h, 30m, 10d, 3600 (default 1h)
  -h, --help                help for token
      --list-scopes         list all valid scopes and exit
  -s, --scope strings       scope to include (repeatable); see --list-scopes
```

### Options inherited from parent commands

```
      --compact                      compact single-line JSON output
      --config string                path to the config file (env: GR4VY_CONFIG)
      --debug                        print debug information to stderr
      --id string                    Gr4vy instance id used for the API host (env: GR4VY_ID)
      --merchant-account-id string   merchant account id (env: GR4VY_MERCHANT_ACCOUNT_ID)
  -o, --output string                output format: json|yaml|table (env: GR4VY_OUTPUT)
      --profile string               configuration profile to use (env: GR4VY_PROFILE)
      --server string                server environment: sandbox|production (env: GR4VY_SERVER)
      --timeout duration             per-request timeout, e.g. 30s
      --token string                 pre-generated bearer token; skips JWT signing (env: GR4VY_TOKEN)
```

### SEE ALSO

* [gr4vy](gr4vy.md)	 - The Gr4vy CLI

