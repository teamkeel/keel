# Repository Guidelines

## Product Context
Keel is a schema-first backend platform: define models, actions, and policies in `keel.schema` and the CLI compiles them into a Go runtime that exposes APIs, migrations, cron, and pub/sub. Docs describe parity between the local runtime and the managed service, which layers hosted Postgres, secrets, deploy workflows, and observability. This repository ships the runtime, CLI, and testing harness for that experience.

## Project Structure & Module Organization
The CLI entrypoint is in `cmd/keel`; runtime and compiler logic live in `runtime`, `schema`, `permissions`, and `util`. Proto assets, fixtures, and migrations reside in `proto/`, `schema/testdata/`, `tools/`, and `migrations/`. TypeScript runtimes and SDKs sit under `packages/*`. Scenario suites live in `integration/` and `testing/`, automation helpers in `scripts/` and `config/`, and built binaries in `bin/`.

## Build, Test, and Development Commands
- `nix-shell` — enter the pinned toolchain before running other workflows.
- `make install` — install Go modules plus `pnpm` dependencies for JS packages.
- `docker compose up -d` — start the Postgres instance used by integration tests.
- `make build` — compile the CLI to `bin/keel`.
- `make test [PACKAGES=./schema/... RUN=TestName]` — execute Go suites with optional filters.
- `make test-js` — trigger every package’s `pnpm run test`.
- `go run cmd/keel/main.go run` — launch the local runtime that mirrors the hosted service.

## Coding Style & Naming Conventions
Keep Go code `gofmt`-clean with idiomatic casing (PascalCase exports, camelCase locals). Run `make lint` for the shared `golangci-lint` policy. TypeScript should match the repo Prettier profile (`make prettier`) and strict `tsconfig.json` settings. Prefer descriptive lowercase filenames with underscores, and only commit regenerated assets (WASM, proto output) when intentional and called out in the PR.

## Testing Guidelines
Write deterministic Go tests (`*_test.go`) and keep fixtures in sibling `testdata` directories. For schema regressions, use `scripts/new-test-case.sh <name>` to scaffold `schema.keel` and `errors.json`, then refresh with `make testdata` when copy shifts. TypeScript tests live beside sources as `.test.ts`; keep snapshots reviewed and mocks minimal. Note flaky or environment-sensitive cases in the PR description.

## Commit & Pull Request Guidelines
Follow Conventional Commits (`feat:`, `fix:`, `chore:`, `refactor:`) with concise imperative subjects under 72 characters—for example `fix: guard schema loader nil pointer`. Rebase or squash so history stays linear. PRs should explain intent, cite test evidence, link issues or docs, and highlight migration, generated-code, or configuration impacts. Add screenshots or logs when behaviour changes, and loop in owners early for cross-package updates.
