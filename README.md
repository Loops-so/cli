
# loops-cli

CLI client for the Loops API.

## Installation

- **Build locally**:
  - `npm install`
  - `npm run build`
  - `./loops --help`

## Authentication options

The CLI supports three ways to provide your Loops API key.

- **Local development**: use `loops login` (file-based config). It is convenient and reasonably secure.
- **CI/CD or scripts running in containers**: prefer `LOOPS_API_KEY` as an environment variable managed by your secrets store.
- **One-off debugging or quickly trying a different key**: use `--api-key` for the command.

If no key is found, commands that talk to the API will fail with:

> No API key found. Pass --api-key, set LOOPS_API_KEY, or run `loops login`.

### 1. File-based login (recommended for day-to-day use)

Run the interactive login command:

```bash
loops login
```

- You will be prompted: `Enter your Loops API key:`
- The key is stored in a config file at `~/.loops/config.json`.
- The config file is written with **chmod 600** and the directory with **chmod 700**.

**Pros**

- **Good security defaults**: stored in a dedicated file with restricted permissions.
- **Convenient**: no need to export env vars or pass flags on every command.
- **Supports multiple endpoints**: each endpoint URL can have its own stored key.

**Cons**

- **Machine-local**: not ideal for ephemeral CI environments where the filesystem is recreated.
- Requires an initial interactive step (`loops login`) on each machine.

### 2. Environment variable (`LOOPS_API_KEY`)

Set the API key in your shell environment:

```bash
export LOOPS_API_KEY=sk_live_123
loops events send --event-name signup ...
```

Or inline for a single command:

```bash
LOOPS_API_KEY=sk_live_123 loops events send --event-name signup ...
```

**Pros**

- **Great for CI/CD** systems using secrets managers.
- Easy to swap keys per shell session.
- Works without touching the filesystem.

**Cons**

- Keys can leak into shell history, process listings, or logs if not handled carefully.
- Slightly more setup each time you open a new shell (unless added to your shell profile).

### 3. `--api-key` flag

Pass the key explicitly on the command line:

```bash
loops --api-key sk_live_123 events send --event-name signup ...
```

This always overrides both `LOOPS_API_KEY` and any stored key from `loops login`.

**Pros**

- **Most explicit**: you always know exactly which key is being used.
- Handy for quick tests or one-off scripts.

**Cons**

- **Least safe**: the key is visible in shell history and process listings while the command runs.
- Not recommended for long-term or shared environments.


