# st-migrate-go — Design & Implementation Plan

## Context & Goals
- Build a reusable Go SDK (`sdk/`) and CLI (`cmd/st-migrate-go/`) for SuperTokens role/permission migrations.
- Migrations remain file-based for v1, but architecture must allow additional sources (S3, Git, embed) later.
- Align closely with `golang-migrate` so existing source/database drivers can be reused or slotted in with minimal shims.
- Maintain KISS/YAGNI: keep YAML schema v1 identical to the CoastalPASS implementation; only add abstraction needed for driver reuse and schema evolution.

## Prior Art Reviewed
- **CoastalPASS authschema** (`internal/authschema` + `migrations/auth`):
  - Discovers paired `NNNN_name.(up|down).yaml` files; validates that filename version matches spec.Version.
  - Stores applied versions in `auth_schema_migrations`; supports `ApplyUp(target)`, `ApplyDown()`, `Status()`.
  - Actions: `ensure: present|absent`, `add`/`remove` permissions, dedupe, simple role ensure/delete via SuperTokens `userroles`.
  - Tests cover discovery, spec validation, and sample platform roles migration.
- **golang-migrate**:
  - Ordering strictly by filename version; drivers for sources (file, s3, github, etc.) and databases (postgres, mysql, sqlite, etc.).
  - CLI uses `--source`, `--database` + commands `up`, `down`, `version/status`, `create`.
  - Drivers implement small interfaces and are intentionally thin; engine holds sequencing and locking logic.

## Key Decisions
- **Directory layout**
  - `cmd/st-migrate-go/` — CLI entrypoint.
  - `st-migrate/` — Public API surface for embedding in other apps.
  - `internal/` — Engine, schema parsers, runner, driver shims, test fixtures; no exported surface.
- **Schema versioning**
  - Filename number controls migration order (golang-migrate style).
  - `spec.version` now represents **schema version** (parser selection), starting at `1`.
  - Parser dispatch table: schema version → parser/validator. Unknown version → clear error.
- **Driver reuse**
  - Mirror `golang-migrate` `source.Driver` and `database.Driver` shapes so their drivers can be imported directly or via thin adapters.
  - Default shipping drivers: file source (built-in) and SQL state store (table `auth_schema_migrations`); adapters allow swapping in official migrate drivers when desired.
- **Runner behavior**
  - `Up(ctx, target *int)` — apply pending migrations by filename order up to target (nil = all).
  - `Down(ctx, steps int)` — roll back `steps` migrations (default 1).
  - `Status(ctx)` — current version + pending list.
  - `DryRun` flag logs planned actions without executing executor/store mutations.
- **Executor abstraction**
  - Interface wrapping SuperTokens role/permission operations to enable testing and future backends.

## YAML Schema v1 (unchanged behavior)
```yaml
version: 1          # schema version, not filename version
actions:
  - role: <string>  # required
    ensure: present|absent (default present)
    add:    [<permission>, ...]  # optional
    remove: [<permission>, ...]  # optional
```
- Filename: `NNNN_name.(up|down).yaml` (width configurable; default 4). Filename number is migration version.
- Validation: role non-empty; ensure ∈ {present, absent}; dedupe/normalize permissions (lowercase, trimmed).

## Components & File Map (implemented)
- `cmd/st-migrate-go/main.go` / `root.go` — Cobra CLI wiring to SDK (`up`, `down`, `status`, `create`), registers postgres/mysql/sqlite migrate drivers.
- `st-migrate/sdk.go`, `st-migrate/options.go` — Public facade building the runner; defaults to file source, SuperTokens executor, and in-memory state store; helper to wrap golang-migrate `database.Driver`.
- `internal/schema/{types.go,parser.go,v1.go}` — Schema version dispatch and v1 parser/validator (schema version defaults to 1).
- `internal/migration/{migration.go,loader.go,runner.go}` — Loads migrations via golang-migrate source drivers (ordered by filename version) and executes Up/Down/Status/Migrate with locking and dirty tracking.
- `internal/state/state.go` — Store interface mirroring migrate’s version/lock surface; `memory/` impl for tests; `file/` durable JSON-backed store (default for CLI); `migrate_adapter.go` to wrap a migrate `database.Driver`.
- `internal/executor/{executor.go,supertokens.go,mock.go}` — Backend abstraction plus SuperTokens default and test mock.
- `internal/create/scaffold.go` — Scaffolds paired up/down YAML files with next sequential version.
- `testdata/migrations/` — Sample schema v1 migrations used by tests.
- `docs/plan.md` — Living design document (this file).

## CLI Contract
- Flags: `--source` (default `file://backend/migrations/auth`), `--database` (DSN or driver URL), `--dry-run`, `--verbose`.
- Commands:
  - `up [target]`
  - `down [steps]` (default 1)
  - `status`
  - `create <name>` (generates `NNNN_name.up.yaml` + `.down.yaml` in source path)
- Exit codes: 0 success, non-zero on error; errors printed with context.

## Store (State) Contract
- Table: `auth_schema_migrations(version INT PRIMARY KEY, applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`.
- Required ops: `Current()`, `Applied()`, `Insert(version)`, `Delete(version)`.
- Compatibility: API mirrors golang-migrate `database.Driver` where feasible to enable reuse/adaptation.

## Implementation Steps (current status)
1) Scaffolded module layout/go.mod and added schema dispatch, executor interface, state store abstractions (done).  
2) Implemented migration loader using golang-migrate `source.Driver`, runner with filename-ordered Up/Down/Status, schema v1 parser, memory store + migrate adapter, SuperTokens executor (done).  
3) Built CLI (`cmd/st-migrate-go`) with `up`, `down`, `status`, `create`; wired flags for `--source`, `--database`, `--dry-run`, `--verbose`; create scaffolder uses schema version 1 by default (done).  
4) Added tests for schema parsing and runner using mock executor and sample migrations (done).  
5) Remaining polish: README refresh, richer store options (e.g., default durable store), additional source/store driver registrations as needed.  

## Open Items / Risks
- Confirm extent of direct compatibility with golang-migrate driver interfaces; may need thin adapter types if their interfaces change.  
- Decide on locking/transaction semantics for SQL store (follow migrate’s lightweight approach or add advisory locks later).  
- Multi-source support is deferred; interfaces should allow registration without redesign.  
