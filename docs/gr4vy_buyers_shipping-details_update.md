## gr4vy buyers shipping-details update

Update a buyer's shipping details

### Synopsis

Update a buyer's shipping details

Update the shipping details associated to a specific buyer.

```
gr4vy buyers shipping-details update <buyer-id> <shipping-details-id> [flags]
```

### Options

```
      --data string   request body as JSON: inline, @file, or - for stdin (ShippingDetailsUpdate)
  -h, --help          help for update
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

* [gr4vy buyers shipping-details](gr4vy_buyers_shipping-details.md)	 - Manage buyers shipping-details

