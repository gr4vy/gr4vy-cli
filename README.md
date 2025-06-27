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
@gr4vy/cli/0.1.4 darwin-arm64 node-v22.15.0
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
* [`gr4vy help [COMMAND]`](#gr4vy-help-command)

## `gr4vy autocomplete [SHELL]`

Display autocomplete installation instructions.

```
USAGE
  $ gr4vy autocomplete [SHELL] [-r]

ARGUMENTS
  SHELL  (zsh|bash|powershell) Shell type

FLAGS
  -r, --refresh-cache  Refresh cache (ignores displaying instructions)

DESCRIPTION
  Display autocomplete installation instructions.

EXAMPLES
  $ gr4vy autocomplete

  $ gr4vy autocomplete bash

  $ gr4vy autocomplete zsh

  $ gr4vy autocomplete powershell

  $ gr4vy autocomplete --refresh-cache
```

_See code: [@oclif/plugin-autocomplete](https://github.com/oclif/plugin-autocomplete/blob/v3.2.31/src/commands/autocomplete/index.ts)_

## `gr4vy help [COMMAND]`

Display help for gr4vy.

```
USAGE
  $ gr4vy help [COMMAND...] [-n]

ARGUMENTS
  COMMAND...  Command to show help for.

FLAGS
  -n, --nested-commands  Include all nested commands in the output.

DESCRIPTION
  Display help for gr4vy.
```

_See code: [@oclif/plugin-help](https://github.com/oclif/plugin-help/blob/v6.2.29/src/commands/help.ts)_
<!-- commandsstop -->
