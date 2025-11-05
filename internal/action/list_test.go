package action

import (
	"testing"
)

func TestListKeyspaces(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	// Call the ListKeyspaces function
	listFunc := ListKeyspaces(testSession)
	keyspaces := listFunc("")

	// Verify that we got keyspaces
	if len(keyspaces) == 0 {
		t.Error("Expected at least one keyspace, got none")
	}

	// Check if system keyspaces are present
	hasSystemKeyspace := false
	hasTestKeyspace := false
	for _, ks := range keyspaces {
		if ks == "system" || ks == "system_schema" {
			hasSystemKeyspace = true
		}
		if ks == "test_keyspace" {
			hasTestKeyspace = true
		}
	}

	if !hasSystemKeyspace {
		t.Error("Expected to find system keyspace")
	}

	if !hasTestKeyspace {
		t.Error("Expected to find test_keyspace")
	}
}

func TestListTables(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	// Call the ListTables function
	listFunc := ListTables(testSession)
	tables := listFunc("")

	// Verify that we got tables
	if len(tables) == 0 {
		t.Error("Expected at least one table, got none")
	}

	// Check if our test tables are present
	hasUsersTable := false
	hasProductsTable := false
	for _, table := range tables {
		if table == "users" {
			hasUsersTable = true
		}
		if table == "products" {
			hasProductsTable = true
		}
	}

	if !hasUsersTable {
		t.Error("Expected to find users table")
	}

	if !hasProductsTable {
		t.Error("Expected to find products table")
	}
}

func TestListColumns(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	tests := []struct {
		name      string
		prefix    string
		line      string
		tableName string
		expected  []string
	}{
		{
			name:      "users table columns",
			prefix:    "SELECT ",
			line:      "SELECT users",
			tableName: "users",
			expected:  []string{"id", "name", "email", "age", "created_at"},
		},
		{
			name:      "products table columns",
			prefix:    "SELECT ",
			line:      "SELECT products",
			tableName: "products",
			expected:  []string{"id", "name", "price", "stock"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the ListColumns function
			listFunc := ListColumns(testSession, tt.prefix)
			columns := listFunc(tt.line)

			// Verify that we got columns
			if len(columns) == 0 {
				t.Errorf("Expected at least one column for table %s, got none", tt.tableName)
			}

			// Check if expected columns are present
			for _, expectedCol := range tt.expected {
				found := false
				for _, col := range columns {
					if col == expectedCol {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find column %s in table %s", expectedCol, tt.tableName)
				}
			}
		})
	}
}

func TestListColumnsInvalidTable(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	// Call the ListColumns function with a non-existent table
	listFunc := ListColumns(testSession, "SELECT ")
	columns := listFunc("SELECT nonexistent_table")

	// Should return empty or handle gracefully
	if len(columns) != 0 {
		t.Error("Expected no columns for non-existent table")
	}
}
