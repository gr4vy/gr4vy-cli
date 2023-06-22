Gr4vy CLI
=================

The Gr4vy CLI is a useful tool for developers to quickly generate tokens, 
query data, and perform basic API manipulation. 

[![oclif](https://img.shields.io/badge/cli-oclif-brightgreen.svg)](https://oclif.io)
[![GitHub license](https://img.shields.io/github/license/gr4vy/gr4vy-cli)](https://github.com/gr4vy/gr4vy-cli/blob/main/LICENSE)

<!-- toc -->
* [Usage](#usage)
* [Getting started](#getting-started)
* [Commands](#commands)
<!-- tocstop -->
# Usage

The Gr4vy CLI is a Node library that can be installed as follows.

<!-- usage -->
```sh-session
$ npm install -g @gr4vy/cli
$ gr4vy COMMAND
running command...
$ gr4vy (--version)
@gr4vy/cli/0.1.3 darwin-arm64 node-v16.20.0
$ gr4vy --help [COMMAND]
USAGE
  $ gr4vy COMMAND
...
```
<!-- usagestop -->

# Getting started

The main use for this CLI is to create an API key. Download your API key and then initialize the SDK with the name of your instance, the target environment, and the filename your key is in.

```sh
gr4vy init acme sandbox private_key.pem
```

This will create a `~/.gr4vyrc.json` file with your config ready for use.

Next, you can create a JWT for use in Embed as well as for server-to-server use.

```sh
gr4vy token expiresIn=10d --scope=buyers.read --scope=buyers.write
gr4vy embed 1299 USD buyer_external_identifier=user-123
```

Additionally, you can inspect each token by passing the `--debug` flag.

```sh
gr4vy token expiresIn=10d --scope=buyers.read --scope=buyers.write --debug
gr4vy embed 1299 USD buyer_external_identifier=user-123 --debug
```

More details on each command is available below.

# Commands
<!-- commands -->
* [`gr4vy autocomplete [SHELL]`](#gr4vy-autocomplete-shell)
* [`gr4vy embed 1299 USD buyer_external_identifier=user-123`](#gr4vy-embed-1299-usd-buyer_external_identifieruser-123)
* [`gr4vy help [COMMANDS]`](#gr4vy-help-commands)
* [`gr4vy init acme sandbox private_key.pem`](#gr4vy-init-acme-sandbox-private_keypem)
* [`gr4vy token expiresIn=10d --scope=buyers.read --scope=buyers.write`](#gr4vy-token-expiresin10d---scopebuyersread---scopebuyerswrite)

## `gr4vy autocomplete [SHELL]`

display autocomplete installation instructions

```
USAGE
  $ gr4vy autocomplete [SHELL] [-r]

ARGUMENTS
  SHELL  (zsh|bash|powershell) Shell type

FLAGS
  -r, --refresh-cache  Refresh cache (ignores displaying instructions)

DESCRIPTION
  display autocomplete installation instructions

EXAMPLES
  $ gr4vy autocomplete

  $ gr4vy autocomplete bash

  $ gr4vy autocomplete zsh

  $ gr4vy autocomplete powershell

  $ gr4vy autocomplete --refresh-cache
```

_See code: [@oclif/plugin-autocomplete](https://github.com/oclif/plugin-autocomplete/blob/v2.3.0/src/commands/autocomplete/index.ts)_

## `gr4vy embed 1299 USD buyer_external_identifier=user-123`

Generate a token for use with Gr4vy Embed.

```
USAGE
  $ gr4vy embed 1299 USD buyer_external_identifier=user-123

ARGUMENTS
  AMOUNT    The amount to generate a token for. This amount needs to be in the smallest denomination for the currency,
            e.g. 1299 for $12.99
  CURRENCY  The 3 digit currency code to generate a token for.

FLAGS
  --debug  Returns the raw header and claim for the token

DESCRIPTION
  Generate a token for use with Gr4vy Embed.

  This token can be used with Embed as it is
  restricted to frontend scopes only.

  It accepts any number of key=value pairs as additional data to be
  pinned in the token.


FLAG DESCRIPTIONS
  --debug  Returns the raw header and claim for the token

    Returns the decoded header and claim from the JWT token without the signature
```

_See code: [dist/commands/embed.ts](https://github.com/gr4vy/gr4vy-cli/blob/v0.1.3/dist/commands/embed.ts)_

## `gr4vy help [COMMANDS]`

Display help for gr4vy.

```
USAGE
  $ gr4vy help [COMMANDS] [-n]

ARGUMENTS
  COMMANDS  Command to show help for.

FLAGS
  -n, --nested-commands  Include all nested commands in the output.

DESCRIPTION
  Display help for gr4vy.
```

_See code: [@oclif/plugin-help](https://github.com/oclif/plugin-help/blob/v5.2.9/src/commands/help.ts)_

## `gr4vy init acme sandbox private_key.pem`

Generate sample .gr4vyrc.json file

```
USAGE
  $ gr4vy init acme sandbox private_key.pem

ARGUMENTS
  GR4VYID      The ID of your instance.
  ENVIRONMENT  (production|sandbox) The environment of your instance.
  PRIVATEKEY   The filename of the private key to add to the config.

DESCRIPTION
  Generate sample .gr4vyrc.json file

  Generates a config file that can be used to generate the token.
```

_See code: [dist/commands/init.ts](https://github.com/gr4vy/gr4vy-cli/blob/v0.1.3/dist/commands/init.ts)_

## `gr4vy token expiresIn=10d --scope=buyers.read --scope=buyers.write`

Generate a bearer token for server-to-server API calls.

```
USAGE
  $ gr4vy token expiresIn=10d --scope=buyers.read --scope=buyers.write

FLAGS
  -e, --expiresIn=<value>
      [default: 1h] The expiry of the token

  -s, --scope=<option>...
      [default: *.read,*.write] A scope to add to this flag
      <options: all.read|all.write|*.read|*.write|anti-fraud-service-definitions.read|anti-fraud-service-definitions.write
      |anti-fraud-services.read|anti-fraud-services.write|buyers.read|buyers.write|buyers.billing-details.read|buyers.bill
      ing-details.write|connections.read|connections.write|digital-wallets.read|digital-wallets.write|flows.read|flows.wri
      te|payment-methods.read|payment-methods.write|payment-options.read|payment-options.write|payment-service-definitions
      .read|payment-service-definitions.write|payment-services.read|payment-services.write|reports.read|reports.write|role
      s.read|roles.write|transactions.read|transactions.write|audit-logs.read|audit-logs.write|checkout-sessions.read|chec
      kout-sessions.write|card-scheme-definitions.read|card-scheme-definitions.write|payment-method-definitions.read|payme
      nt-method-definitions.write|reset.read|reset.write|merchant-accounts.read|merchant-accounts.write>

  --debug
      Returns the raw header and claim for the token

DESCRIPTION
  Generate a bearer token for server-to-server API calls.

  This token should be used with care as it is not
  restricted to any specific frontend scopes only.


FLAG DESCRIPTIONS
  -e, --expiresIn=<value>  The expiry of the token

    The expiration expressed in seconds or a string describing a time span vercel/ms.

  -s, --scope=all.read|all.write|*.read|*.write|anti-fraud-service-definitions.read|anti-fraud-service-definitions.write|anti-fraud-services.read|anti-fraud-services.write|buyers.read|buyers.write|buyers.billing-details.read|buyers.billing-details.write|connections.read|connections.write|digital-wallets.read|digital-wallets.write|flows.read|flows.write|payment-methods.read|payment-methods.write|payment-options.read|payment-options.write|payment-service-definitions.read|payment-service-definitions.write|payment-services.read|payment-services.write|reports.read|reports.write|roles.read|roles.write|transactions.read|transactions.write|audit-logs.read|audit-logs.write|checkout-sessions.read|checkout-sessions.write|card-scheme-definitions.read|card-scheme-definitions.write|payment-method-definitions.read|payment-method-definitions.write|reset.read|reset.write|merchant-accounts.read|merchant-accounts.write...

    A scope to add to this flag

    A single scope to add to this JWT

  --debug  Returns the raw header and claim for the token

    Returns the decoded header and claim from the JWT token without the signature
```

_See code: [dist/commands/token.ts](https://github.com/gr4vy/gr4vy-cli/blob/v0.1.3/dist/commands/token.ts)_
<!-- commandsstop -->
