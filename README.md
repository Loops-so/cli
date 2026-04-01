# Loops CLI

The official Loops CLI

> [!NOTE]
> This is pre-release, alpha software. While we're cooking up our first official release there may be breaking changes and other bugs.

## Installation

### Homebrew

```
brew install loops-so/tap/loops
```

### Script

```
curl -fsSL --proto '=https' --tlsv1.2 https://raw.githubusercontent.com/loops-so/cli/main/install.sh | bash
```

You can optionally specify the release version and installation path with positional arguments,

```
... | bash -s -- v0.0.0

```

```
... | bash -s -- latest ~/.local/bin
```

## Getting Started

### Auth

The Loops CLI requires a Loops API key to use.

1. Grab an API key from https://app.loops.so/settings?page=api
1. `loops auth login --name name-for-your-key`

Alternatively, Loops will use the value from the `LOOPS_API_KEY` environment variable if set.

That's it! During development, consider the CLI's `--help` output as the source of truth for features and flags.
