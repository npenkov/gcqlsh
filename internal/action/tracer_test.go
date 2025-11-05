package action

import (
	"testing"
	"time"

	"github.com/gocql/gocql"
)

func TestNewTracer(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	tests := []struct {
		name           string
		tracingEnabled bool
	}{
		{
			name:           "tracer with tracing disabled",
			tracingEnabled: false,
		},
		{
			name:           "tracer with tracing enabled",
			tracingEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tracingEnabled {
				testSession.EnableTracing()
			} else {
				testSession.DisableTracing()
			}

			tracer := NewTracer(testSession)
			if tracer == nil {
				t.Error("Expected tracer to be non-nil")
			}

			if tracer.cks == nil {
				t.Error("Expected tracer.cks to be non-nil")
			}

			if tt.tracingEnabled {
				if tracer.tw == nil {
					t.Error("Expected tracer.tw to be non-nil when tracing is enabled")
				}
				if tracer.buf == nil {
					t.Error("Expected tracer.buf to be non-nil when tracing is enabled")
				}
				if tracer.cf == nil {
					t.Error("Expected tracer.cf to be non-nil when tracing is enabled")
				}
			} else {
				if tracer.tw != nil {
					t.Error("Expected tracer.tw to be nil when tracing is disabled")
				}
				if tracer.buf != nil {
					t.Error("Expected tracer.buf to be nil when tracing is disabled")
				}
				if tracer.cf != nil {
					t.Error("Expected tracer.cf to be nil when tracing is disabled")
				}
			}

			// Clean up
			if tracer.cf != nil {
				tracer.cf()
			}
		})
	}
}

func TestTracerQuery(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	tests := []struct {
		name           string
		tracingEnabled bool
		query          string
	}{
		{
			name:           "query without tracing",
			tracingEnabled: false,
			query:          "SELECT * FROM users LIMIT 1",
		},
		{
			name:           "query with tracing",
			tracingEnabled: true,
			query:          "SELECT * FROM users LIMIT 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tracingEnabled {
				testSession.EnableTracing()
			} else {
				testSession.DisableTracing()
			}

			tracer := NewTracer(testSession)
			defer func() {
				if tracer.cf != nil {
					tracer.cf()
				}
			}()

			query := tracer.Query(tt.query)
			if query == nil {
				t.Error("Expected query to be non-nil")
			}

			// Execute the query to ensure it works
			iter := query.Iter()
			if iter == nil {
				t.Error("Expected iterator to be non-nil")
			}

			// Close the iterator
			if err := iter.Close(); err != nil {
				t.Errorf("Expected no error closing iterator, got: %v", err)
			}
		})
	}
}

func TestTracerClose(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	// Insert test data
	testID := gocql.TimeUUID()
	err := testSession.Session.Query(`
		INSERT INTO users (id, name, email, age, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, testID, "Tracer Test", "tracer@example.com", 40, time.Now()).Exec()
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	tests := []struct {
		name           string
		tracingEnabled bool
	}{
		{
			name:           "close tracer without tracing",
			tracingEnabled: false,
		},
		{
			name:           "close tracer with tracing",
			tracingEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tracingEnabled {
				testSession.EnableTracing()
			} else {
				testSession.DisableTracing()
			}

			tracer := NewTracer(testSession)

			// Execute a query
			query := tracer.Query("SELECT * FROM users WHERE id = ?", testID)
			iter := query.Iter()
			iter.Close()

			// Close should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Tracer.Close() panicked: %v", r)
				}
			}()
			tracer.Close()
		})
	}
}

func TestTracerQueryWithValues(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	testSession.DisableTracing()

	// Insert test data
	testID := gocql.TimeUUID()
	err := testSession.Session.Query(`
		INSERT INTO users (id, name, email, age, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, testID, "Query Test", "query@example.com", 45, time.Now()).Exec()
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	tracer := NewTracer(testSession)
	defer func() {
		if tracer.cf != nil {
			tracer.cf()
		}
	}()

	// Query with bind values
	query := tracer.Query("SELECT name FROM users WHERE id = ?", testID)
	var name string
	err = query.Scan(&name)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if name != "Query Test" {
		t.Errorf("Expected name to be 'Query Test', got: %s", name)
	}
}

func TestNewTraceWriter(t *testing.T) {
	skipIfDockerUnavailable(t)
	if testSession == nil {
		t.Fatal("Test session is not initialized")
	}

	// We can't directly test NewTraceWriter as it's used internally,
	// but we can test it indirectly through NewTracer with tracing enabled
	testSession.EnableTracing()
	defer testSession.DisableTracing()

	tracer := NewTracer(testSession)
	defer func() {
		if tracer.cf != nil {
			tracer.cf()
		}
	}()

	if tracer.tw == nil {
		t.Error("Expected trace writer to be initialized when tracing is enabled")
	}

	if tracer.tw.session == nil {
		t.Error("Expected trace writer session to be non-nil")
	}

	if tracer.tw.w == nil {
		t.Error("Expected trace writer io.Writer to be non-nil")
	}
}
