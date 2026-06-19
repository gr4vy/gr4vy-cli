## gr4vy transactions events list

List transaction events

### Synopsis

List transaction events

Retrieve a paginated list of events related to processing a transaction, including status changes, API requests, and webhook delivery attempts. Events are listed in chronological order, with the most recent events first.

```
gr4vy transactions events list <transaction-id> [flags]
```

### Options

```
      --cursor string   pagination cursor
  -h, --help            help for list
      --limit int       maximum number of items to return
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

* [gr4vy transactions events](gr4vy_transactions_events.md)	 - Manage transactions events

