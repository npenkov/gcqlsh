package action

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/npenkov/gcqlsh/internal/db"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	cassandraPool     *dockertest.Pool
	cassandraResource *dockertest.Resource
	testSession       *db.CQLKeyspaceSession
	cassandraHost     string
	cassandraPort     int
	dockerAvailable   bool
)

// TestMain acts as the test suite entry point
func TestMain(m *testing.M) {
	var err error

	// Create dockertest pool
	cassandraPool, err = dockertest.NewPool("")
	if err != nil {
		log.Printf("Could not construct pool: %s. Tests will be skipped.", err)
		dockerAvailable = false
		os.Exit(0)
	}

	err = cassandraPool.Client.Ping()
	if err != nil {
		log.Printf("Could not connect to Docker: %s. Tests will be skipped.", err)
		dockerAvailable = false
		os.Exit(0)
	}

	dockerAvailable = true

	// Pull and start Cassandra container
	cassandraResource, err = cassandraPool.RunWithOptions(&dockertest.RunOptions{
		Repository: "cassandra",
		Tag:        "4.1",
		Env: []string{
			"CASSANDRA_BROADCAST_ADDRESS=127.0.0.1",
			"CASSANDRA_LISTEN_ADDRESS=0.0.0.0",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start Cassandra resource: %s", err)
	}

	// Set expiration for the container to 10 minutes
	if err := cassandraResource.Expire(600); err != nil {
		log.Fatalf("Could not set expiration: %s", err)
	}

	cassandraHost = "localhost"
	cassandraPort = 9042

	// Get the actual port that Docker assigned
	hostPort := cassandraResource.GetPort("9042/tcp")
	fmt.Printf("Cassandra container started on port %s\n", hostPort)

	// Wait for Cassandra to be ready
	if err := cassandraPool.Retry(func() error {
		cluster := gocql.NewCluster(cassandraResource.GetHostPort("9042/tcp"))
		cluster.Consistency = gocql.One
		cluster.Timeout = 10 * time.Second
		cluster.ProtoVersion = 3
		cluster.IgnorePeerAddr = true
		cluster.DisableInitialHostLookup = true

		session, err := cluster.CreateSession()
		if err != nil {
			return err
		}
		defer session.Close()

		// Try a simple query to ensure Cassandra is ready
		if err := session.Query("SELECT now() FROM system.local").Exec(); err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Fatalf("Could not connect to Cassandra: %s", err)
	}

	// Setup test keyspace and tables
	if err := setupTestSchema(cassandraResource.GetHostPort("9042/tcp")); err != nil {
		log.Fatalf("Could not setup test schema: %s", err)
	}

	// Create test session
	session, closeFunc, err := db.NewSession(cassandraResource.GetHostPort("9042/tcp"), 0, "", "", "test_keyspace")
	if err != nil {
		log.Fatalf("Could not create test session: %s", err)
	}

	testSession = &db.CQLKeyspaceSession{
		Host:             cassandraHost,
		Port:             cassandraPort,
		Username:         "",
		Password:         "",
		Session:          session,
		ActiveKeyspace:   "test_keyspace",
		NewSchema:        true,
		IsInitialized:    true,
		CloseSessionFunc: closeFunc,
		TracingEnabled:   false,
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if testSession != nil && testSession.CloseSessionFunc != nil {
		testSession.CloseSessionFunc()
	}

	if err := cassandraPool.Purge(cassandraResource); err != nil {
		log.Fatalf("Could not purge Cassandra resource: %s", err)
	}

	os.Exit(code)
}

// setupTestSchema creates a test keyspace and tables for testing
func setupTestSchema(hostPort string) error {
	cluster := gocql.NewCluster(hostPort)
	cluster.Consistency = gocql.One
	cluster.Timeout = 10 * time.Second
	cluster.ProtoVersion = 3
	cluster.IgnorePeerAddr = true
	cluster.DisableInitialHostLookup = true

	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Create test keyspace
	if err := session.Query(`
		CREATE KEYSPACE IF NOT EXISTS test_keyspace
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
	`).Exec(); err != nil {
		return fmt.Errorf("failed to create test keyspace: %w", err)
	}

	// Use the test keyspace
	session.Close()
	cluster.Keyspace = "test_keyspace"
	session, err = cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create session with keyspace: %w", err)
	}
	defer session.Close()

	// Create test tables
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			name TEXT,
			email TEXT,
			age INT,
			created_at TIMESTAMP
		)
	`).Exec(); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY,
			name TEXT,
			price DECIMAL,
			stock INT
		)
	`).Exec(); err != nil {
		return fmt.Errorf("failed to create products table: %w", err)
	}

	// Insert test data
	userID := gocql.TimeUUID()
	if err := session.Query(`
		INSERT INTO users (id, name, email, age, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, userID, "John Doe", "john@example.com", 30, time.Now()).Exec(); err != nil {
		return fmt.Errorf("failed to insert test user: %w", err)
	}

	productID := gocql.TimeUUID()
	if err := session.Query(`
		INSERT INTO products (id, name, price, stock)
		VALUES (?, ?, ?, ?)
	`, productID, "Test Product", 19.99, 100).Exec(); err != nil {
		return fmt.Errorf("failed to insert test product: %w", err)
	}

	return nil
}

// skipIfDockerUnavailable skips a test if Docker is not available
func skipIfDockerUnavailable(t *testing.T) {
	if !dockerAvailable {
		t.Skip("Skipping test because Docker is not available")
	}
}
