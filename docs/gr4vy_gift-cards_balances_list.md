## gr4vy gift-cards balances list

List gift card balances

### Synopsis

List gift card balances

Fetch the balances for one or more gift cards.

```
gr4vy gift-cards balances list [flags]
```

### Options

```
      --data string   request body as JSON: inline, @file, or - for stdin (GiftCardBalanceRequest)
  -h, --help          help for list
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

* [gr4vy gift-cards balances](gr4vy_gift-cards_balances.md)	 - Manage gift-cards balances

