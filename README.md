# Loops CLI

The official Loops CLI

> [!NOTE]
> This is pre-release, alpha software. While we're cooking up our first official release, the easiest way to get started is to build the project yourself.

## Install

Install script,
```
# latest release
curl -fsSL --proto '=https' --tlsv1.2 https://raw.githubusercontent.com/loops-so/cli/main/install.sh | bash

# specify the release version, "latest" or a release from GH, e.g., v0.0.0
curl -fsSL --proto '=https' --tlsv1.2 https://raw.githubusercontent.com/loops-so/cli/main/install.sh | bash -- v0.0.0

# install to a specific path (requires version)
curl -fsSL --proto '=https' --tlsv1.2 https://raw.githubusercontent.com/loops-so/cli/main/install.sh | bash -- latest ~/bin
```

## Getting Started

### Build
1. Clone this repo and cd into the directory.
1. Install Go 1.26.1 (or whatever is listed in `go.mod`)
1. Run `go build -o ./loops .` to build a `loops` binary in the current directory. You're free to move it (`./loops`) to a location in your `$PATH` if you'd like.

### Auth

The Loops CLI requires a Loops API key to use.

1. Grab an API key from https://app.loops.so/settings?page=api
1. `loops auth login --name name-for-your-key`

Alternatively, Loops will use the value from the `LOOPS_API_KEY` environment variable if set.

That's it! During development, consider the CLI's `--help` output as the source of truth for features and flags.
