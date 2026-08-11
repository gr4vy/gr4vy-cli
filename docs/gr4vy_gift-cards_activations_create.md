## gr4vy gift-cards activations create

Activate a gift card

### Synopsis

Activate a gift card

Activate a physical gift card through the primary gift card service. Set `store` to `true` to also store the activated gift card.

```
gr4vy gift-cards activations create [flags]
```

### Options

```
      --data string              request body as JSON: inline, @file, or - for stdin (GiftCardActivationCreate)
  -h, --help                     help for create
      --idempotency-key string   unique key to make the request idempotent
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

* [gr4vy gift-cards activations](gr4vy_gift-cards_activations.md)	 - Manage gift-cards activations

