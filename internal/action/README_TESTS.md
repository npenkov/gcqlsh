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

The tests use `ory/dockertest` to run a Cassandra container for integration testing. This ensures that:
1. Tests run against a real Cassandra instance
2. Tests are isolated and repeatable
3. No external Cassandra setup is required

### Test Suite Structure

- `main_test.go` - Test suite entry point that manages the Docker container lifecycle
  - Starts Cassandra container before tests
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

### Run All Tests
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

### Run Tests with Coverage
```bash
go test ./internal/action/... -v -cover -timeout 15m
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

### Docker Connection Issues
If you get errors about Docker connection:
1. Ensure Docker daemon is running: `docker ps`
2. Check Docker permissions: your user should be in the `docker` group
3. Try running with sudo (not recommended): `sudo go test ...`

### Container Startup Timeouts
If Cassandra takes too long to start:
1. Increase timeout: `-timeout 20m`
2. Check Docker resources (CPU/Memory)
3. Check Docker logs: `docker logs <container-id>`

### Port Conflicts
If port 9042 is already in use:
- The tests use Docker's automatic port mapping
- Check for other Cassandra instances: `docker ps`

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
