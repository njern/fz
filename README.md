# fz

`fz` is a command-line client for the [Fizzy](https://app.fizzy.do) project management API.

It lets you work with boards, cards, columns, comments, steps, notifications, tags, users, webhooks, and raw API endpoints without leaving the terminal.

## Features

- Authenticate with Fizzy using a magic link or a personal access token
- View and manage boards, columns, cards, comments, and steps
- Manage notifications, tags, pinned cards, users, and webhooks
- Make raw authenticated API requests with `fz api`
- Generate shell completion scripts for Bash, Zsh, Fish, and PowerShell

## Installation

### Build from source

`fz` currently builds with Go 1.26.

Build the binary in the repository root:

```sh smoke-docs
make build
./fz --help
```

If you want `fz` on your `PATH`, install it with Go:

```sh
go install .
```

If you prefer not to keep a local checkout, install the module directly:

```sh
go install github.com/njern/fz@latest
```

## Authentication and configuration

`fz` stores its config in `os.UserConfigDir()/fz/config.json`.

The main configuration values are:

- `host`: Fizzy instance URL, defaulting to `https://app.fizzy.do`
- `account`: Default account slug used when `--account` is not passed

The commands in this section and the quick-start section below talk to a live Fizzy account. They are covered by `make integration-test`, not the offline `make smoke-docs` check.

Start with the interactive login flow:

```sh
fz auth login
fz auth status
```

To log in with a personal access token instead:

```sh
printf '%s\n' "$FIZZY_TOKEN" | fz auth login --with-token
```

If you belong to more than one Fizzy account, `fz auth login` will save a default account for future commands. You can override it per command with `--account` or `-a`.

To inspect or update configuration:

```sh smoke-docs
fz config list
fz config set account demo-account
fz config get account
```

To point `fz` at a different Fizzy host:

```sh
fz config set host https://fizzy.example.com
```

To switch the default account without logging in again:

```sh
fz config set account my-account
```

To inspect the available authentication commands and flags:

```sh smoke-docs
fz auth login --help
fz auth status --help
fz auth create-token --help
```

## Quick start

### 1. Check your current work

```sh
fz status
fz notification list
fz pin list
```

### 2. Find a board and inspect its cards

```sh
fz board list
fz board view <board-id>
fz card list --board <board-id>
fz card view <card-number>
```

### 3. Create and update records

```sh
fz board create "Roadmap"
fz card create --board <board-id> --title "Ship v1.0"
fz comment create <card-number> --body "Looks good to me."
fz board edit <board-id> --description "Public roadmap"
```

### 4. Make a raw API call when you need something lower-level

```sh
fz api /my/identity
fz api cards
```

When scripting against multiple accounts, override the saved default account for a single command:

```sh
fz board list --account my-account
fz card list --board <board-id> --account another-account
```

## Scripting and automation

For destructive commands in scripts or other non-interactive environments, pass `--yes` to skip confirmation prompts:

```sh
fz board delete <board-id> --yes
fz comment delete <card-number> <comment-id> --yes
```

Use `fz auth status --check` for a machine-friendly authentication check that exits non-zero when credentials are missing or invalid:

```sh
fz auth status --check
```

To pass the saved token to another tool:

```sh
fz auth token
```

## Shell completion

Generate completion scripts with:

```sh
fz completion <shell>
```

### Zsh

Load completions in the current shell:

```zsh smoke-docs
source <(fz completion zsh)
```

Load completions automatically in every new shell by adding this to `~/.zshrc`:

```zsh smoke-docs
eval "$(fz completion zsh)"
```

If you prefer completion files instead of `eval`, write the script to a completion directory:

```sh smoke-docs
mkdir -p ~/.zsh/completions
fz completion zsh > ~/.zsh/completions/_fz
```

Then make sure your `~/.zshrc` contains:

```zsh smoke-docs
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit
compinit
```

### Other shells

Bash:

```bash smoke-docs
source <(fz completion bash)
```

Fish:

```sh smoke-docs
mkdir -p ~/.config/fish/completions
fz completion fish > ~/.config/fish/completions/fz.fish
```

PowerShell:

```powershell
fz completion powershell | Out-String | Invoke-Expression
```

## Command reference

Core command groups:

| Command | Purpose |
| --- | --- |
| `fz auth` | Log in, log out, inspect auth state, print or create tokens |
| `fz config` | Inspect and change `host` and default `account` |
| `fz status` | Show notifications and assigned cards |
| `fz board`, `fz column`, `fz card`, `fz comment`, `fz step` | Work with boards and their contents |
| `fz notification`, `fz pin`, `fz tag`, `fz user`, `fz webhook` | Work with account-level resources |
| `fz api` | Make raw authenticated API requests |
| `fz completion` | Generate shell completion scripts |

Use help output to discover flags and subcommands:

```sh smoke-docs
fz --help
fz api --help
fz config --help
fz board create --help
fz card create --help
```

## Development

Common targets:

```sh
make build
make smoke-docs
make test
make lint
make fmt
make integration-test
```

`make smoke-docs` executes the `smoke-docs` code blocks in this README in an isolated temporary home directory.

`make integration-test` expects a working Fizzy environment with valid credentials and account access already configured.

Use `fz --help` or `fz <command> --help` for the full command reference.
