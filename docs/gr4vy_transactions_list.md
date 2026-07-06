## gr4vy transactions list

List transactions

### Synopsis

List transactions

Returns a paginated list of transactions for the merchant account, sorted by most recently updated. You can filter, sort, and search transactions using query parameters.

```
gr4vy transactions list [flags]
```

### Options

```
      --amount-eq int                             amount-eq parameter
      --amount-gte int                            amount-gte parameter
      --amount-lte int                            amount-lte parameter
      --buyer-email-address string                buyer-email-address parameter
      --buyer-external-identifier string          buyer-external-identifier parameter
      --buyer-id string                           buyer-id parameter
      --buyer-search strings                      buyer-search parameter
      --checkout-session-id string                checkout-session-id parameter
      --country strings                           country parameter
      --currency strings                          currency parameter
      --cursor string                             pagination cursor
      --disputed                                  disputed parameter
      --error-code strings                        error-code parameter
      --external-identifier string                external-identifier parameter
      --gift-card-id string                       gift-card-id parameter
      --gift-card-last4 string                    gift-card-last4 parameter
      --has-gift-card-redemptions                 has-gift-card-redemptions parameter
      --has-refunds                               has-refunds parameter
      --has-settlements                           has-settlements parameter
  -h, --help                                      help for list
      --id string                                 id parameter
      --ip-address string                         ip-address parameter
      --is-subsequent-payment                     is-subsequent-payment parameter
      --limit int                                 maximum number of items to return
      --merchant-initiated                        merchant-initiated parameter
      --metadata strings                          metadata parameter
      --payment-link-id string                    payment-link-id parameter
      --payment-method-bin string                 payment-method-bin parameter
      --payment-method-country string             payment-method-country parameter
      --payment-method-fingerprint string         payment-method-fingerprint parameter
      --payment-method-id string                  payment-method-id parameter
      --payment-method-label string               payment-method-label parameter
      --payment-method-scheme strings             payment-method-scheme parameter
      --payment-service-id strings                payment-service-id parameter
      --payment-service-transaction-id string     payment-service-transaction-id parameter
      --pending-review                            pending-review parameter
      --reauthorized-from-transaction-id string   reauthorized-from-transaction-id parameter
      --reconciliation-id string                  reconciliation-id parameter
      --search string                             free-text search filter
      --used-3ds                                  used-3ds parameter
```

### Options inherited from parent commands

```
      --compact                      compact single-line JSON output
      --config string                path to the config file (env: GR4VY_CONFIG)
      --debug                        print debug information to stderr
      --merchant-account-id string   merchant account id (env: GR4VY_MERCHANT_ACCOUNT_ID)
  -o, --output string                output format: json|yaml|table (env: GR4VY_OUTPUT)
      --profile string               configuration profile to use (env: GR4VY_PROFILE)
      --server string                server environment: sandbox|production (env: GR4VY_SERVER)
      --timeout duration             per-request timeout, e.g. 30s
      --token string                 pre-generated bearer token; skips JWT signing (env: GR4VY_TOKEN)
```

### SEE ALSO

* [gr4vy transactions](gr4vy_transactions.md)	 - Manage transactions

