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

## Auth

The CLI requires a Loops API key. Get one from [Settings > API](https://app.loops.so/settings?page=api).

### Keyring storage

Store a key with `loops auth login --name <name>`. Run this again with a different name to store keys for multiple teams.

- `loops auth use <name>` — set a stored key as the default
- `loops auth list` — list stored keys and see which is the default

Use `--team <name>` on any command to pick a specific stored key.

### Environment variable

Set `LOOPS_API_KEY` to use a key directly — useful for CI or when keyring storage isn't available.

### Precedence

When multiple keys are available, the CLI resolves them in this order:

1. `LOOPS_API_KEY` env var
1. `--team` flag
1. The current default (set via `loops auth use`)
