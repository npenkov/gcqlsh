package db

import (
	"fmt"

	"github.com/gocql/gocql"
)

type CQLKeyspaceSession struct {
	Host           string
	Port           int
	Session        *gocql.Session
	ActiveKeyspace string
	NewSchema      bool
	IsInitialized  bool
}

// FetchKeyspaces obtains the list of all keyspaces available
func (cks *CQLKeyspaceSession) FetchKeyspaces() ([]string, error) {
	var keyspaceName string
	keyspaces := make([]string, 0)
	// We need to have info on what type of schema
	if !cks.IsInitialized {
		cks.Init()
	}
	var cqlSchemaSelect string
	if cks.NewSchema {
		cqlSchemaSelect = "select keyspace_name from system_schema.keyspaces"
	} else {
		cqlSchemaSelect = "select keyspace_name from system.schema_keyspaces"
	}
	iter := cks.Session.Query(cqlSchemaSelect).Iter()
	for iter.Scan(&keyspaceName) {
		keyspaces = append(keyspaces, keyspaceName)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return keyspaces, nil
}

// FetchTables returns a list of all tables in the Active keyspace
func (cks *CQLKeyspaceSession) FetchTables() ([]string, error) {
	tables := make([]string, 0)

	if schema, err := cks.Session.KeyspaceMetadata(cks.ActiveKeyspace); err == nil {
		for table := range schema.Tables {
			tables = append(tables, table)
		}
	}

	return tables, nil
}

func (cks *CQLKeyspaceSession) FetchColumns(tableName string) (map[string]*gocql.ColumnMetadata, error) {
	schema, err := cks.Session.KeyspaceMetadata(cks.ActiveKeyspace)
	if err != nil {
		return nil, err
	}

	tm, ok := schema.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("Table %s not in schema", tableName)
	}

	return tm.Columns, nil
}

// Init method for switching the new/old schema detection
func (cks *CQLKeyspaceSession) Init() {
	cks.NewSchema = false
	if schemaKS, err := cks.Session.KeyspaceMetadata("system"); err == nil {
		if _, ok := schemaKS.Tables["schema_keyspaces"]; !ok {
			cks.NewSchema = true
		}
	}
	// newSchema := cks.Session.KeyspaceMetadata("system_schema").Tables["keyspaces"]
	cks.IsInitialized = true
}
