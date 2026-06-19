## gr4vy login

Log in with email and password

### Synopsis

Authenticate against the Gr4vy session endpoint and store the resulting tokens for the active profile. The CLI refreshes the access token automatically.

```
gr4vy login [flags]
```

### Options

```
      --email string     login email
  -h, --help             help for login
      --password-stdin   read the password from stdin
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

