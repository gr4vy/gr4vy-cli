## gr4vy merchant-accounts create

Create a merchant account

### Synopsis

Create a merchant account

Create a new merchant account in an instance.

```
gr4vy merchant-accounts create [flags]
```

### Options

```
      --data string   request body as JSON: inline, @file, or - for stdin (MerchantAccountCreate)
  -h, --help          help for create
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

* [gr4vy merchant-accounts](gr4vy_merchant-accounts.md)	 - Manage merchant-accounts

