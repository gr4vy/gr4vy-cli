## gr4vy merchant-accounts three-ds-configuration list

List 3DS configurations for merchant

### Synopsis

List 3DS configurations for merchant

List all 3DS configurations for a merchant account.

```
gr4vy merchant-accounts three-ds-configuration list <merchant-account-id> [flags]
```

### Options

```
      --currency string   currency parameter
  -h, --help              help for list
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

* [gr4vy merchant-accounts three-ds-configuration](gr4vy_merchant-accounts_three-ds-configuration.md)	 - Manage merchant-accounts three-ds-configuration

