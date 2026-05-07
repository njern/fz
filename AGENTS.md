# Repository Guidelines

## Project Structure & Module Organization

`fz` is a Go CLI for the Fizzy API. The entry point is `main.go`, with Cobra commands in `cmd/`. Reusable code lives under `internal/`: API client code in `internal/api`, configuration in `internal/config`, authentication in `internal/auth`, and terminal rendering in `internal/render`. Tests sit next to covered code as `*_test.go` files. Support scripts live in `script/`, GitHub Actions configuration is in `.github/workflows/`, and `fz.png` is the README image asset.

## Build, Test, and Development Commands

Run commands from the repository root.

- `make install-tools`: installs pinned tools (`golangci-lint`, `gofumpt`, `modernize`, `wsl`).
- `make build`: builds `./fz` with version metadata.
- `make run`: builds and runs the local CLI.
- `make test`: runs all Go tests with `go test ./...`.
- `make lint`: runs `golangci-lint run ./...`.
- `make smoke-docs`: executes README `smoke-docs` blocks in an isolated home.
- `make ci`: runs deterministic CI checks: lint, tests, docs smoke test, formatting check, and WSL check.
- `make integration-test`: builds first, then runs live Fizzy integration tests; valid credentials and account access are required.

## Coding Style & Naming Conventions

Use Go 1.26. Format code with `make fmt`, which runs `gofumpt -l -w .`; verify without rewriting via `make fmt-check`. Keep package names short and lowercase. Name Cobra command files after the command area, such as `cmd/card.go` and `cmd/card_test.go`. Prefer clear sentinel errors and typed helpers when behavior is shared across commands.

## Testing Guidelines

Use the standard `testing` package and table-driven tests where useful. Test names should describe behavior, for example `TestConfigSet_RepairsMalformedConfig`. For command behavior, use the existing helpers in `cmd/`. Add regression tests before fixing bugs. Use `make test` for unit coverage and `make ci` before opening a PR. Run `make integration-test` only when the change needs live Fizzy API coverage.

## Commit & Pull Request Guidelines

Recent commits use short, imperative summaries, usually lowercase, such as `refresh README development targets`. Keep commits focused on one logical change. Before committing, check staged scope with `git diff --cached --name-only`. Pull requests should explain the user-visible change, list validation commands run, call out integration-test requirements or skipped checks, and link related issues when available.

## Security & Configuration Tips

Do not commit tokens or generated config. `fz` stores local config at `os.UserConfigDir()/fz/config.json`. Prefer environment variables or local config for credentials, and avoid putting live account names or tokens in tests.
