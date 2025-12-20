# AGENTS

## Global Rules and Required Reading
- Global AGENTS.md (authoritative; always read first).
- rules/Go_General.md (repo contains Go code; read and apply before these local rules).

## Repository Overview
- Go module that provides a SuperTokens role/permission migration SDK (`st-migrate/`) and a CLI (`cmd/st-migrate-go`).
- Core engine lives under `internal/` and is not part of the public API.

## Project Structure and Boundaries
- `cmd/st-migrate-go/`: Cobra-based CLI entrypoint and commands.
- `st-migrate/`: Public SDK (`package stmigrate`) for embedding in other Go services.
- `internal/migration/`: Loads/executes migrations and coordinates locking/state.
- `internal/schema/`: YAML schema parsing and normalization (schema versioned).
- `internal/executor/`: Executor interface + SuperTokens implementation + mock.
- `internal/state/`: Store interface and adapters; implementations in `internal/state/file/` and `internal/state/memory/`.
- `internal/create/`: Migration scaffold generation.
- `testdata/migrations/`: Fixture migrations used by tests.

Rules:
- Keep `internal/` packages unexported; external consumers should use `st-migrate/` only.
- Do not add imports from `cmd/` into library code.
- Keep CLI wiring in `cmd/` thin; core behavior belongs in `internal/` or `st-migrate/`.

## Domain Conventions
- Migration filenames define order: `NNNN_name.(up|down).yaml` (width defaults to 4).
- YAML `version` is the schema version (currently v1), not the filename version.
- `ensure` values are `present` or `absent`; permissions are normalized/deduped by schema parsers.
- Default CLI source is `file://backend/migrations/auth` and default state file is `.st-migrate/state.json`.

## Dependency and Integration Constraints
- SuperTokens operations are abstracted via `internal/executor.Executor`; use `executor.NewMock()` in tests.
- State tracking must go through `internal/state.Store` or a wrapped `golang-migrate` `database.Driver`.
- Source loading uses `golang-migrate` `source.Driver`; keep compatibility with its URL schemes.
- `st-migrate` should stay free of CLI-specific flags/output concerns.

## Testing and Validation
- Run `go test ./...` for all changes.
- CI enforces coverage >= 80% (see `.github/workflows/ci.yml`); add tests for new behavior.
- Prefer table-driven tests where appropriate and use `stretchr/testify/require` like existing tests.
- Use `testdata/migrations/` or `t.TempDir()` fixtures; avoid network access in tests.

## CI/CD Conventions
- GitHub Actions `Go CI` uses Go 1.24.x and fails if coverage < 80%.
- Releases are cut from tags `v*` via GoReleaser; keep CLI paths and module layout compatible.

## Extending This Repo
- New schema versions: add parser in `internal/schema/`, register it, and add tests.
- New executors: implement `Executor`, wire via `stmigrate.Config.Executor`, add unit tests.
- New state stores: implement `state.Store` or wrap a migrate driver; add tests for lock/version/close.
- New CLI commands/flags: add to `cmd/st-migrate-go/root.go` and cover with CLI tests.
- New migration sources: rely on `golang-migrate` source drivers and update tests/docs as needed.
