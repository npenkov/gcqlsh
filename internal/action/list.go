package action

import (
	"strings"

	"github.com/npenkov/gcqlsh/internal/db"
)

func ListKeyspaces(cks *db.CQLKeyspaceSession) func(string) []string {
	return func(line string) []string {
		keyspaces, _ := cks.FetchKeyspaces()
		return keyspaces
	}
}

func ListTables(cks *db.CQLKeyspaceSession) func(string) []string {
	return func(line string) []string {
		tables, _ := cks.FetchTables()
		return tables
	}
}

func ListColumns(cks *db.CQLKeyspaceSession, prefix string) func(string) []string {
	return func(line string) []string {
		// get table from the line and fetch the columns
		tableName := strings.TrimPrefix(line, prefix)
		tableName = strings.Fields(tableName)[0]
		columns, _ := cks.FetchColumns(tableName)
		cols := make([]string, 0)
		for col := range columns {
			cols = append(cols, col)
		}
		return cols
	}
}
