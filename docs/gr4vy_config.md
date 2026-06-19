## gr4vy config

Manage configuration profiles

### Options

```
  -h, --help   help for config
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

* [gr4vy](gr4vy.md)	 - The Gr4vy CLI
* [gr4vy config add](gr4vy_config_add.md)	 - Add or update a profile
* [gr4vy config import](gr4vy_config_import.md)	 - Import the legacy ~/.gr4vyrc.json into a profile
* [gr4vy config key](gr4vy_config_key.md)	 - Print the active profile's private key (PEM)
* [gr4vy config list](gr4vy_config_list.md)	 - List configured profiles
* [gr4vy config path](gr4vy_config_path.md)	 - Print the resolved config path and secret backend
* [gr4vy config remove](gr4vy_config_remove.md)	 - Remove a profile and its secrets
* [gr4vy config show](gr4vy_config_show.md)	 - Show a profile (secrets redacted)
* [gr4vy config use](gr4vy_config_use.md)	 - Set the active profile

