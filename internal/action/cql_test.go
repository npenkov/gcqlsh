package action

import (
	"testing"
	"time"

	"github.com/gocql/gocql"
)

func TestProcessCommand_Exit(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	breakLoop, continueLoop, err := ProcessCommand("exit", testSession)

	if !breakLoop {
		t.Error("Expected breakLoop to be true for 'exit' command")
	}

	if continueLoop {
		t.Error("Expected continueLoop to be false for 'exit' command")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestProcessCommand_EmptyString(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	breakLoop, continueLoop, err := ProcessCommand("", testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for empty command")
	}

	if !continueLoop {
		t.Error("Expected continueLoop to be true for empty command")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestProcessCommand_Comment(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	breakLoop, continueLoop, err := ProcessCommand("-- This is a comment", testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for comment")
	}

	if !continueLoop {
		t.Error("Expected continueLoop to be true for comment")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestProcessCommand_UseKeyspace(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	// Test USE command (lowercase)
	breakLoop, continueLoop, err := ProcessCommand("use system", testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for USE command")
	}

	if !continueLoop {
		t.Error("Expected continueLoop to be true for USE command")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify keyspace was changed
	if testSession.ActiveKeyspace != "system" {
		t.Errorf("Expected active keyspace to be 'system', got: %s", testSession.ActiveKeyspace)
	}

	// Switch back to test_keyspace
	_, _, _ = ProcessCommand("USE test_keyspace", testSession)
	if testSession.ActiveKeyspace != "test_keyspace" {
		t.Errorf("Expected active keyspace to be 'test_keyspace', got: %s", testSession.ActiveKeyspace)
	}
}

func TestProcessCommand_UseKeyspaceUppercase(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	// Test USE command (uppercase)
	breakLoop, continueLoop, err := ProcessCommand("USE system;", testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for USE command")
	}

	if !continueLoop {
		t.Error("Expected continueLoop to be true for USE command")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify keyspace was changed
	if testSession.ActiveKeyspace != "system" {
		t.Errorf("Expected active keyspace to be 'system', got: %s", testSession.ActiveKeyspace)
	}

	// Switch back to test_keyspace
	_, _, _ = ProcessCommand("use test_keyspace;", testSession)
}

func TestProcessCommand_DescribeKeyspaces(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	breakLoop, continueLoop, err := ProcessCommand("desc keyspaces", testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for DESC command")
	}

	if continueLoop {
		t.Error("Expected continueLoop to be false for DESC command")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestProcessCommand_DescribeTables(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	breakLoop, continueLoop, err := ProcessCommand("DESC tables", testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for DESC command")
	}

	if continueLoop {
		t.Error("Expected continueLoop to be false for DESC command")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestProcessCommand_DescribeTable(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	breakLoop, continueLoop, err := ProcessCommand("desc table users", testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for DESC TABLE command")
	}

	if continueLoop {
		t.Error("Expected continueLoop to be false for DESC TABLE command")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestProcessCommand_SelectQuery(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	// Insert a test record first
	testID := gocql.TimeUUID()
	err := testSession.Session.Query(`
		INSERT INTO users (id, name, email, age, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, testID, "Test User", "test@example.com", 25, time.Now()).Exec()
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test SELECT query
	breakLoop, continueLoop, err := ProcessCommand("SELECT * FROM users LIMIT 1", testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for SELECT command")
	}

	if continueLoop {
		t.Error("Expected continueLoop to be false for SELECT command")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestProcessCommand_InsertQuery(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	testID := gocql.TimeUUID()
	query := "INSERT INTO users (id, name, email, age, created_at) VALUES (" +
		testID.String() + ", 'Jane Doe', 'jane@example.com', 28, toTimestamp(now()))"

	breakLoop, continueLoop, err := ProcessCommand(query, testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for INSERT command")
	}

	if continueLoop {
		t.Error("Expected continueLoop to be false for INSERT command")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify the insert
	var count int
	err = testSession.Session.Query(`SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		t.Errorf("Failed to verify insert: %v", err)
	}

	if count < 1 {
		t.Error("Expected at least one record in users table")
	}
}

func TestProcessCommand_UpdateQuery(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	// Insert a test record first
	testID := gocql.TimeUUID()
	err := testSession.Session.Query(`
		INSERT INTO users (id, name, email, age, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, testID, "Update Test", "update@example.com", 30, time.Now()).Exec()
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Update the record
	query := "UPDATE users SET age = 31 WHERE id = " + testID.String()
	breakLoop, continueLoop, err := ProcessCommand(query, testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for UPDATE command")
	}

	if continueLoop {
		t.Error("Expected continueLoop to be false for UPDATE command")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestProcessCommand_DeleteQuery(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	// Insert a test record first
	testID := gocql.TimeUUID()
	err := testSession.Session.Query(`
		INSERT INTO users (id, name, email, age, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, testID, "Delete Test", "delete@example.com", 35, time.Now()).Exec()
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Delete the record
	query := "DELETE FROM users WHERE id = " + testID.String()
	breakLoop, continueLoop, err := ProcessCommand(query, testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for DELETE command")
	}

	if continueLoop {
		t.Error("Expected continueLoop to be false for DELETE command")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestProcessCommand_InvalidQuery(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	// Test invalid query
	breakLoop, continueLoop, err := ProcessCommand("INVALID QUERY SYNTAX", testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for invalid query")
	}

	if continueLoop {
		t.Error("Expected continueLoop to be false for invalid query")
	}

	if err == nil {
		t.Error("Expected an error for invalid query")
	}
}

func TestProcessCommand_TracingOn(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	// Ensure tracing is off initially
	testSession.DisableTracing()

	breakLoop, continueLoop, err := ProcessCommand("tracing on", testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for TRACING command")
	}

	if continueLoop {
		t.Error("Expected continueLoop to be false for TRACING command")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !testSession.TracingEnabled {
		t.Error("Expected tracing to be enabled")
	}
}

func TestProcessCommand_TracingOff(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	// Enable tracing first
	testSession.EnableTracing()

	breakLoop, continueLoop, err := ProcessCommand("TRACING OFF", testSession)

	if breakLoop {
		t.Error("Expected breakLoop to be false for TRACING command")
	}

	if continueLoop {
		t.Error("Expected continueLoop to be false for TRACING command")
	}

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if testSession.TracingEnabled {
		t.Error("Expected tracing to be disabled")
	}
}
