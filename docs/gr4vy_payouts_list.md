## gr4vy payouts list

List payouts created

### Synopsis

List payouts created

Returns a list of payouts made.

```
gr4vy payouts list [flags]
```

### Options

```
      --cursor string                      pagination cursor
      --external-identifier string         external-identifier parameter
  -h, --help[=true|false]                  help for list
      --limit int                          maximum number of items to return
      --payment-service-payout-id string   payment-service-payout-id parameter
```

### Options inherited from parent commands

```
      --compact[=true|false]         compact single-line JSON output
      --config string                path to the config file (env: GR4VY_CONFIG)
      --debug[=true|false]           print debug information to stderr
      --id string                    Gr4vy instance id used for the API host (env: GR4VY_ID)
      --merchant-account-id string   merchant account id (env: GR4VY_MERCHANT_ACCOUNT_ID)
  -o, --output string                output format: json|yaml|table (env: GR4VY_OUTPUT)
      --profile string               configuration profile to use (env: GR4VY_PROFILE)
      --server string                server environment: sandbox|production (env: GR4VY_SERVER)
      --timeout duration             per-request timeout, e.g. 30s
      --token string                 pre-generated bearer token; skips JWT signing (env: GR4VY_TOKEN)
```

### SEE ALSO

* [gr4vy payouts](gr4vy_payouts.md)	 - Manage payouts

