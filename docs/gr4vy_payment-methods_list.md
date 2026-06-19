## gr4vy payment-methods list

List all payment methods

### Synopsis

List all payment methods

List all stored payment method.

```
gr4vy payment-methods list [flags]
```

### Options

```
      --buyer-external-identifier string   buyer-external-identifier parameter
      --buyer-id string                    buyer-id parameter
      --cursor string                      pagination cursor
      --external-identifier string         external-identifier parameter
  -h, --help                               help for list
      --limit int                          maximum number of items to return
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

* [gr4vy payment-methods](gr4vy_payment-methods.md)	 - Manage payment-methods

