## gr4vy config import

Import the legacy ~/.gr4vyrc.json into a profile

```
gr4vy config import [flags]
```

### Options

```
      --delete-legacy[=true|false]   delete the legacy file after import
      --from string                  path to the legacy config (default ~/.gr4vyrc.json)
  -h, --help[=true|false]            help for import
      --name string                  profile name to create (default "default")
      --set-active[=true|false]      make the imported profile active
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

* [gr4vy config](gr4vy_config.md)	 - Manage configuration profiles

