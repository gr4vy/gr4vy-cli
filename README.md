# Gr4vy CLI

[![GitHub license](https://img.shields.io/github/license/gr4vy/gr4vy-cli)](https://github.com/gr4vy/gr4vy-cli/blob/main/LICENSE)

A fast, single-binary command-line interface for the [Gr4vy](https://gr4vy.com) payment
orchestration platform. It wraps the official [`gr4vy-go`](https://github.com/gr4vy/gr4vy-go)
SDK and exposes **every operation the SDK supports** — buyers, transactions, payment
methods, checkout sessions, reports, and more — plus helpers for tokens, Embed, and
profile management.

The command surface is **generated from the `gr4vy-go` SDK's types**, so it stays in sync
with the SDK automatically.

> The previous TypeScript/npm CLI (`@gr4vy/cli`) is deprecated. This Go rewrite replaces it
> and is distributed as a self-contained binary.

## Install

### Homebrew

```sh
brew install gr4vy/tap/gr4vy
```

### Scoop (Windows)

```powershell
scoop bucket add gr4vy https://github.com/gr4vy/scoop-bucket
scoop install gr4vy
```

### Go

```sh
go install github.com/gr4vy/gr4vy-cli@latest
```

### Binaries

Download a prebuilt archive for your platform from the
[releases page](https://github.com/gr4vy/gr4vy-cli/releases).

## Quick start

```sh
# Create a profile (interactive); stores your private key in the OS keychain.
gr4vy init

# …or non-interactively:
gr4vy config add acme \
  --id acme --environment sandbox \
  --merchant-account-id default \
  --key-file ./private_key.pem --set-active

# Call the API — output is a table on a terminal, JSON when piped.
gr4vy buyers list
gr4vy transactions list --limit 20 --status capture_pending -o table
gr4vy buyers create --data '{"display_name":"Jane Doe","external_identifier":"user-123"}'
gr4vy transactions get <transaction-id>
gr4vy transactions refunds create <transaction-id> --data '{"amount":500}'
```

## Profiles & multiple instances

A merchant can keep many instances/keys as named profiles in
`~/.config/gr4vy/config.toml` (override with `--config` or `GR4VY_CONFIG`). Only non-secret
data lives there; **private keys and login tokens are kept in the OS keychain** (with a
`0600` file fallback on headless systems).

```sh
gr4vy config list
gr4vy config use acme-prod
gr4vy config show acme
gr4vy --profile acme-sandbox transactions list
```

Resolution precedence for every setting: **flag > environment variable > active profile >
default**. Common env vars: `GR4VY_PROFILE`, `GR4VY_ID`, `GR4VY_SERVER`,
`GR4VY_MERCHANT_ACCOUNT_ID`, `GR4VY_PRIVATE_KEY` / `GR4VY_PRIVATE_KEY_FILE`,
`GR4VY_TOKEN`, `GR4VY_OUTPUT`.

Any private-key source — `GR4VY_PRIVATE_KEY`, a `--key-file`/`--key-path` file, `--key-stdin`,
or `--key-env` — accepts **either a raw PEM or a base64-encoded PEM**. Base64 is convenient
for CI, where multi-line PEMs are awkward in environment variables:

```sh
export GR4VY_PRIVATE_KEY="$(base64 < private_key.pem)"
gr4vy transactions list
```

### Migrating from the legacy CLI

```sh
gr4vy config import            # imports ~/.gr4vyrc.json into a profile
```

## Authentication

Two modes per profile:

- **Key (default).** A merchant private key signs ES512 JWTs locally for every request.
- **Login (email/password).** `gr4vy login` exchanges your dashboard credentials for a
  session and refreshes it automatically.

```sh
gr4vy login --email you@example.com     # prompts for the password
gr4vy logout
```

## Tokens

```sh
# Server-to-server API token (JWT).
gr4vy token --scope transactions.read --scope buyers.write --expires-in 1h
gr4vy token --list-scopes
gr4vy token --debug                     # also prints the decoded claims

# Embed token for the checkout form.
gr4vy embed 1299 USD buyer_external_identifier=user-123
gr4vy embed 1299 USD --checkout-session # also creates a checkout session
```

## Output

`-o, --output json|yaml|table` (defaults to `table` on a TTY, `json` otherwise).
Lists accept `--limit` and `--cursor`; the `next_cursor` is returned for paging.

## How it stays up to date

The CLI generates one typed command per SDK operation directly from the `gr4vy-go` types —
the resource tree, method signatures, and doc comments. When a new SDK is published (an
`sdk_updated` dispatch, or a daily cron), a workflow bumps `gr4vy-go`, regenerates, and opens
a draft PR — so the CLI self-maintains. There's no dependency on the OpenAPI spec: a typed
CLI can only expose what the SDK ships.

## Development

```sh
make build        # build ./gr4vy with version info
make gen          # regenerate commands from the gr4vy-go SDK types
make gen-refresh  # bump gr4vy-go to latest, then regenerate
make test         # unit + golden tests
make e2e          # live e2e suite (needs PRIVATE_KEY or ./private_key.pem)
```

Layout: hand-written commands and runtime live under `internal/`; generated API commands
are in `internal/commands/generated`; the generator is `internal/gen`.
