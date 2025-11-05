# Action Package Tests

This directory contains comprehensive unit tests for the `action` package, which handles all the interactive commands and operations in gcqlsh.

## Overview

The tests are designed to verify the functionality of all public methods in the action package:
- `ListKeyspaces` - Lists all available keyspaces
- `ListTables` - Lists all tables in the active keyspace
- `ListColumns` - Lists columns for a specific table
- `ProcessCommand` - Processes CQL commands and special commands (USE, DESC, TRACING, etc.)
- `NewTracer` - Creates a tracer for query tracing

## Test Infrastructure

### Docker-based Testing

The tests support two modes of operation:

1. **Local Mode** (using dockertest): Tests spin up their own Cassandra container using `ory/dockertest`
2. **Docker Compose Mode** (recommended for macOS): Tests connect to a Cassandra service managed by docker-compose

Both modes ensure that:
1. Tests run against a real Cassandra instance
2. Tests are isolated and repeatable
3. No external Cassandra setup is required

### Test Suite Structure

- `main_test.go` - Test suite entry point that manages the Docker container lifecycle
  - Detects environment (local vs docker-compose)
  - Starts Cassandra container (local mode) or connects to existing service (docker-compose mode)
  - Creates test keyspace and tables
  - Initializes test session
  - Cleans up after all tests complete

- `list_test.go` - Tests for list functions (keyspaces, tables, columns)
- `cql_test.go` - Tests for CQL command processing
- `tracer_test.go` - Tests for query tracing functionality

## Requirements

### System Requirements
- Docker installed and running
- Go 1.24 or later
- Internet connection (for pulling Cassandra Docker image on first run)

### Go Dependencies
All dependencies are managed via go.mod:
- `github.com/ory/dockertest/v3` - Docker container management
- `github.com/gocql/gocql` - Cassandra driver
- Standard testing package

## Running the Tests

### Using Makefile (Recommended)

The easiest way to run tests is using the provided Makefile targets:

#### Run Tests (Default - uses docker-compose)
```bash
make test
```
This is the recommended approach for all platforms, especially macOS.

#### Run Tests with Docker Compose (Best for macOS)
```bash
make test-docker-compose
```
This starts a Cassandra container via docker-compose and runs tests against it. This avoids the port mapping issues that occur on macOS with dockertest.

#### Run Tests Locally
```bash
make test-local
```
Uses dockertest to spin up a container. May fail on macOS due to Docker networking issues.

#### Run Tests with Coverage
```bash
make test-coverage
```
Generates a coverage report in `coverage.html`.

#### Run Tests in Docker Container
```bash
make test-docker
```
Runs tests inside a Docker container with access to the Docker socket.

### Direct Go Command

You can also run tests directly with go:

```bash
go test ./internal/action/... -v -timeout 15m
```

The `-timeout 15m` flag is important as the first run needs to:
1. Pull the Cassandra Docker image (~500MB)
2. Start the container
3. Wait for Cassandra to be ready
4. Run all tests

### Run Specific Test
```bash
go test ./internal/action/... -v -run TestListKeyspaces
```

### Run Tests without Docker

If Docker is not available, the tests will skip gracefully:
```
2025/11/05 15:43:00 Could not connect to Docker: ... Tests will be skipped.
ok  	github.com/npenkov/gcqlsh/internal/action	0.016s
```

## Test Data

The test suite automatically creates:

### Test Keyspace
- Name: `test_keyspace`
- Replication: SimpleStrategy with RF=1

### Test Tables

#### users table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    name TEXT,
    email TEXT,
    age INT,
    created_at TIMESTAMP
)
```

#### products table
```sql
CREATE TABLE products (
    id UUID PRIMARY KEY,
    name TEXT,
    price DECIMAL,
    stock INT
)
```

Test data is automatically inserted into these tables.

## Test Coverage

### list_test.go
- ✅ `TestListKeyspaces` - Verifies keyspace listing
- ✅ `TestListTables` - Verifies table listing
- ✅ `TestListColumns` - Verifies column listing for multiple tables
- ✅ `TestListColumnsInvalidTable` - Tests error handling

### cql_test.go
- ✅ `TestProcessCommand_Exit` - Tests exit command
- ✅ `TestProcessCommand_EmptyString` - Tests empty input handling
- ✅ `TestProcessCommand_Comment` - Tests comment handling
- ✅ `TestProcessCommand_UseKeyspace` - Tests USE keyspace command
- ✅ `TestProcessCommand_UseKeyspaceUppercase` - Tests case insensitivity
- ✅ `TestProcessCommand_DescribeKeyspaces` - Tests DESC KEYSPACES
- ✅ `TestProcessCommand_DescribeTables` - Tests DESC TABLES
- ✅ `TestProcessCommand_DescribeTable` - Tests DESC TABLE
- ✅ `TestProcessCommand_SelectQuery` - Tests SELECT queries
- ✅ `TestProcessCommand_InsertQuery` - Tests INSERT queries
- ✅ `TestProcessCommand_UpdateQuery` - Tests UPDATE queries
- ✅ `TestProcessCommand_DeleteQuery` - Tests DELETE queries
- ✅ `TestProcessCommand_InvalidQuery` - Tests error handling
- ✅ `TestProcessCommand_TracingOn` - Tests tracing enable
- ✅ `TestProcessCommand_TracingOff` - Tests tracing disable

### tracer_test.go
- ✅ `TestNewTracer` - Tests tracer creation with/without tracing
- ✅ `TestTracerQuery` - Tests query execution with tracer
- ✅ `TestTracerClose` - Tests tracer cleanup
- ✅ `TestTracerQueryWithValues` - Tests parameterized queries
- ✅ `TestNewTraceWriter` - Tests trace writer initialization

## Troubleshooting

### macOS Connection Errors (MOST COMMON)

If you see errors like:
```
gocql: unable to dial control conn ::1:55006: dial tcp [::1]:55006: connect: connection refused
gocql: unable to dial control conn 127.0.0.1:55006: dial tcp 127.0.0.1:55006: connect: connection refused
```

**Solution:** Use `make test` or `make test-docker-compose` instead of `make test-local`.

**Why?** Docker Desktop for Mac uses a VM, and dockertest's port mapping doesn't always work correctly. The docker-compose approach uses Docker's internal networking, which is more reliable.

### Docker Connection Issues
If you get errors about Docker connection:
1. Ensure Docker daemon is running: `docker ps`
2. Check Docker permissions: your user should be in the `docker` group
3. Try running with sudo (not recommended): `sudo go test ...`

### Container Startup Timeouts
If Cassandra takes too long to start:
1. Increase timeout: `-timeout 20m`
2. Check Docker resources (CPU/Memory) in Docker Desktop settings
3. Check Docker logs: `docker logs <container-id>`
4. Wait for healthcheck: The docker-compose setup includes a healthcheck

### Port Conflicts
If port 9042 is already in use:
- The docker-compose setup exposes 9042, so stop any running Cassandra instances
- Check for other Cassandra instances: `docker ps`
- Stop conflicting containers: `docker stop <container-id>`

### Cassandra Configuration Errors
If you see errors like:
```
listen_address cannot be a wildcard address (0.0.0.0)!
```

This has been fixed in the configuration. The docker-compose and test setup now use:
- `CASSANDRA_RPC_ADDRESS=0.0.0.0` (for client connections - correct)
- `CASSANDRA_BROADCAST_ADDRESS` (for node identification)
- No `CASSANDRA_LISTEN_ADDRESS` (avoids the wildcard error)

If you're modifying the configuration, remember:
- `LISTEN_ADDRESS` is for inter-node communication (cannot be 0.0.0.0 in Cassandra 4.x)
- `RPC_ADDRESS` is for CQL client connections (can be 0.0.0.0)
- `BROADCAST_ADDRESS` is the address nodes advertise to each other

## Continuous Integration

For CI/CD environments without Docker:
- Tests will skip gracefully
- Consider using Docker-in-Docker or similar solutions
- Alternative: Use testcontainers-go with appropriate configuration

## Future Improvements

Potential enhancements:
- [ ] Add benchmarks for performance testing
- [ ] Add tests for concurrent operations
- [ ] Add tests for authentication scenarios
- [ ] Add tests for SSL/TLS connections
- [ ] Add more edge cases and error scenarios
- [ ] Add integration tests with different Cassandra versions
