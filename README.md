<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>

[![Go Reference][go-ref-shield]][go-ref-url]
[![Go Report Card][goreport-shield]][goreport-url]
[![License][license-shield]][license-url]

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/BeardedWonderDev/st-migrate-go">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a>

  <h3 align="center">st-migrate-go</h3>

  <p align="center">
    SuperTokens role/permission migrations with golang-migrate style sources, as an SDK and CLI.
    <br />
    <a href="https://github.com/BeardedWonderDev/st-migrate-go"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/BeardedWonderDev/st-migrate-go/issues">Report Bug</a>
    ·
    <a href="https://github.com/BeardedWonderDev/st-migrate-go/issues">Request Feature</a>
  </p>
</div>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li><a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li><a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->
## About The Project

st-migrate-go is a migration runner for SuperTokens roles and permissions. It mirrors the sequencing and source semantics of `golang-migrate`, so you can point at any migrate-compatible source (default: local files) while using a SuperTokens-aware executor. It ships as:
- **SDK (`sdk/`)** for embedding migration logic in Go services.
- **CLI (`cmd/st-migrate-go`)** for terminal-driven migration management.

Key behaviors:
- Migration order is defined solely by filename versions (`0001_name.up.yaml` / `.down.yaml`).
- YAML `version` key now denotes the **schema version** (v1 today) to allow future format evolution.
- Supports golang-migrate source drivers out of the box; state can be backed by migrate database drivers or the default in-memory store.
- Default executor applies changes via `supertokens-golang` role/permission APIs; pluggable for other backends.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

- [Go](https://go.dev/)
- [Cobra](https://github.com/spf13/cobra)
- [golang-migrate](https://github.com/golang-migrate/migrate) (sources/stores)
- [SuperTokens Go SDK](https://github.com/supertokens/supertokens-golang)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->
## Getting Started

Follow these steps to install and run the CLI, or embed the SDK in your Go application.

### Prerequisites
- Go 1.20+ (tested with 1.24)
- Access to your SuperTokens instance for applying role changes

### Installation
1. Clone the repo
   ```sh
   git clone https://github.com/BeardedWonderDev/st-migrate-go.git
   cd st-migrate-go
   ```
2. Install the CLI
   ```sh
   go install ./cmd/st-migrate-go
   ```
3. (Optional) Vendor dependencies
   ```sh
   go mod tidy
   ```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE EXAMPLES -->
## Usage

### CLI
```sh
# Show current and pending migrations
st-migrate-go --source file://backend/migrations/auth status

# Apply all pending migrations
st-migrate-go up

# Apply up to a target version
st-migrate-go up 5

# Roll back one migration (default) or N steps
st-migrate-go down
st-migrate-go down 2

# Generate paired up/down files with the next version number
st-migrate-go create add-reporting-roles
```
Flags:
- `--source` migrate-style source URL (default `file://backend/migrations/auth`)
- `--database` migrate database driver URL for state tracking (postgres, mysql, sqlite registered in CLI build)
- `--state-file` path to a JSON state store used when `--database` is empty (default `.st-migrate/state.json`)
- `--dry-run` log actions without executing
- `--verbose` enable debug logging

### SDK
```go
cfg := sdk.Config{
    SourceURL: "file://backend/migrations/auth",
    // Optional: Store, Executor, Logger, DryRun, Registry
}
r, err := sdk.New(cfg)
if err != nil {
    // handle
}
defer r.Close()
if err := r.Up(context.Background(), nil); err != nil {
    // handle
}
```

Using a golang-migrate database driver (example: Postgres):
```go
import (
    "github.com/BeardedWonderDev/st-migrate-go/sdk"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/lib/pq"
)

// create or reuse *sql.DB ...
db, _ := sql.Open("postgres", "<dsn>")
driver, _ := postgres.WithInstance(db, &postgres.Config{})
cfg := sdk.Config{
    SourceURL: "file://backend/migrations/auth",
    Store:     sdk.WrapMigrateDatabase(driver),
}
r, _ := sdk.New(cfg)
defer r.Close()
_ = r.Up(context.Background(), nil)
```

YAML schema v1:
```yaml
version: 1          # schema version
actions:
  - role: app:admin
    ensure: present  # present|absent (default present)
    add: [app:read, app:write]
    remove: []
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ROADMAP -->
## Roadmap

- Add additional sources (embed, git, s3)
- Schema v2 exploration (richer permission metadata)
- Advisory locking strategy for multi-runner safety

See the [open issues](https://github.com/BeardedWonderDev/st-migrate-go/issues) for a full list of proposed features and known issues.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTRIBUTING -->
## Contributing

Contributions are welcome! Please open an issue to discuss changes before submitting a PR. Keep changes small and covered by tests.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- LICENSE -->
## License

Distributed under the Unlicense. See `LICENSE.txt` for details.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTACT -->
## Contact

Project Link: [https://github.com/BeardedWonderDev/st-migrate-go](https://github.com/BeardedWonderDev/st-migrate-go)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

- [golang-migrate](https://github.com/golang-migrate/migrate)
- [SuperTokens](https://supertokens.com/)
- [Cobra](https://github.com/spf13/cobra)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
[go-ref-shield]: https://pkg.go.dev/badge/github.com/BeardedWonderDev/st-migrate-go.svg
[go-ref-url]: https://pkg.go.dev/github.com/BeardedWonderDev/st-migrate-go
[goreport-shield]: https://goreportcard.com/badge/github.com/BeardedWonderDev/st-migrate-go
[goreport-url]: https://goreportcard.com/report/github.com/BeardedWonderDev/st-migrate-go
[license-shield]: https://img.shields.io/github/license/BeardedWonderDev/st-migrate-go.svg
[license-url]: https://github.com/BeardedWonderDev/st-migrate-go/blob/main/LICENSE.txt
