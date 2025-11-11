# gcqlsh

Cassandra command line shell written in Golang

![](screenshots/gcqlsh_rec.gif?raw=true)

## Motivation

Having a cassandra command line shell utility in one binary distributable.

## Where it comes in hand?

- Building docker images for cassandra from Alpine with no Python.
- Running cql shell on all platforms.
- Automating cassandra schema creation without need to install python dependencies.

## Installation

### Using Homebrew (macOS/Linux)

```bash
brew tap npenkov/gcqlsh
brew install gcqlsh
```

Or install directly from the repository:

```bash
brew install npenkov/gcqlsh/gcqlsh
```

### Using Nix Flakes

#### Direct run without installation

```bash
nix run github:npenkov/gcqlsh
```

#### Install to user profile

```bash
nix profile install github:npenkov/gcqlsh
```

#### Add to NixOS configuration or Home Manager

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    gcqlsh.url = "github:npenkov/gcqlsh";
  };

  outputs = { self, nixpkgs, gcqlsh }: {
    # ... your configuration
    environment.systemPackages = [
      gcqlsh.packages.${system}.default
    ];
  };
}
```

#### Development shell

```bash
nix develop github:npenkov/gcqlsh
```

### Using Go

```bash
go install github.com/npenkov/gcqlsh/cmd/gcqlsh@latest
```

### Download Binary

Download the latest release binary for your platform from the [releases page](https://github.com/npenkov/gcqlsh/releases).

## Building

```
go build -o gocqlsh src/github.com/npenkov/gcqlsh/cmd/gcqlsh.go
```

## Fatures

- Running DDL script files from command line
- Support for Cassandra 2.1+/ScyllaDB
- CQL Support
- Statement tracing
- `desc` command with
  - `keyspaces` - simple list
  - `tables` - simple list
  - `table` - simple list of columns and types
- Auto completition for commands:
  - `use` - keyspaces
  - `desc` - tables
  - `select` - tables
  - `update` - tables and columns
  - `delete` - tables
  - `insert` - tables

## Still missing

- Paging in interactive results
- DDL Statements when describing Keyspaces and tables
- Expanded rows
- Code assistance for different keyspaces
- Node token awareness

## Command line help

```
gcqlsh -h
Usage of gcqlsh:
  -f string
        Execute file containing cql statements instead of having interacive session
  -fail-on-error
        Stop execution if statement from file fails.
  -host string
        Cassandra host to connect to (default "127.0.0.1")
  -k string
        Default keyspace to connect to (default "system")
  -no-color
        Console without colors
  -password string
        Password used for the connection
  -port int
        Cassandra RPC port (default 9042)
  -print-confirmation
        Print 'ok' on successfuly executed cql statement from the file
  -print-cql
        Print Statements that are executed from a file
  -username string
        Username used for the connection
  -v    Version information
```

## Planned features

- `desc` for table
- Column code assistance for
  - `select`
  - `update`
  - `delete`
  - `insert`

## Package dependencies

- [Readline](https://github.com/chzyer/readline)
- [Color](https://github.com/fatih/color)
- [Gocql](https://github.com/gocql/gocql)

## Releases

This project uses [GoReleaser](https://goreleaser.com/) for automated releases.

### Creating a Release

1. Update the `VERSION` file with the new version number:

   ```bash
   echo "0.0.4" > VERSION
   ```

2. Commit the version change:

   ```bash
   git add VERSION
   git commit -m "Bump version to 0.0.4"
   ```

3. Create and push a tag:

   ```bash
   git tag -a v0.0.4 -m "Release v0.0.4"
   git push origin v0.0.4
   ```

4. GitHub Actions will automatically:
   - Run tests
   - Build binaries for multiple platforms (Linux, macOS, Windows)
   - Create checksums
   - Generate release notes
   - Publish the release on GitHub

### Manual Release (Local Testing)

To test the release process locally without publishing:

```bash
# Install goreleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Test the release (snapshot mode)
goreleaser release --snapshot --clean

# Check the dist/ directory for generated artifacts
ls -la dist/
```

## Development

### Running Tests

The project includes comprehensive unit tests that run against a real Cassandra instance using Docker.

**Quick Start (Recommended for all platforms, especially macOS):**

```bash
make test
```

**Alternative test commands:**

```bash
make test-local          # Run tests locally (may fail on macOS)
make test-docker-compose # Run tests using docker-compose (best for macOS)
make test-coverage       # Run tests with coverage report
```

**macOS Users:** If you encounter connection errors like:

```
gocql: unable to dial control conn 127.0.0.1:55006: connect: connection refused
```

Use `make test` (which runs docker-compose mode) instead of `make test-local`.

**Why?** Docker Desktop for Mac uses a VM, and direct port mapping from dockertest doesn't always work. The docker-compose approach uses Docker's internal networking for reliable connectivity.

For more details, see [internal/action/README_TESTS.md](internal/action/README_TESTS.md).

### Building

```bash
make build              # Build all platforms
make linux              # Build for Linux
make darwin             # Build for macOS
make windows            # Build for Windows
```

---

Written with [vim-go](https://github.com/fatih/vim-go)

## License

> Copyright (c) 2016-2025 Nick Penkov. All rights reserved.
> Use of this source code is governed by a MIT-style
> license that can be found in the LICENSE file.
