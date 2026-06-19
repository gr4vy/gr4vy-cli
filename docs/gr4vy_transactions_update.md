## gr4vy transactions update

Manually update a transaction

### Synopsis

Manually update a transaction

Manually updates a transaction.

```
gr4vy transactions update <transaction-id> [flags]
```

### Options

```
      --data string   request body as JSON: inline, @file, or - for stdin (TransactionUpdate)
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

* [gr4vy transactions](gr4vy_transactions.md)	 - Manage transactions

