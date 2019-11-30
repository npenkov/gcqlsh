package action

import (
	"fmt"
	"strings"

	"github.com/npenkov/gcqlsh/internal/db"
)

func describeCmd(cks *db.CQLKeyspaceSession, cmd string) error {
	desc := strings.TrimPrefix(strings.TrimPrefix(cmd, "desc "), "DESC ")
	desc = strings.TrimSpace(desc)
	if strings.HasPrefix(desc, "keyspaces") || strings.HasPrefix(desc, "KEYSPACES") {
		keyspaces, _ := cks.FetchKeyspaces()
		for ksi := range keyspaces {
			fmt.Printf("%s\n", keyspaces[ksi])
		}
		return nil
	}

	if strings.HasPrefix(desc, "keyspace") || strings.HasPrefix(desc, "KEYSPACE") {
		// TODO:
		return nil
	}

	if strings.HasPrefix(desc, "tables") || strings.HasPrefix(desc, "TABLES") {
		tables, _ := cks.FetchTables()
		for ti := range tables {
			fmt.Printf("%s\n", tables[ti])
		}
		return nil
	}

	if strings.HasPrefix(desc, "table") || strings.HasPrefix(desc, "TABLE") {
		tableName := strings.TrimPrefix(strings.TrimPrefix(desc, "table"), "TABLE")
		tableName = strings.TrimSuffix(tableName, ";")
		tableName = strings.TrimSpace(tableName)

		columns, _ := cks.FetchColumns(tableName)
		colWidthName := 0
		colWidthType := 0

		colsData := make(map[string]string, len(columns))

		for _, col := range columns {
			colsData[col.Name] = fmt.Sprintf("%s", col.Type)
			if len(col.Name) > colWidthName {
				colWidthName = len(col.Name)
			}
			if len(col.Name) > colWidthType {
				colWidthType = len(col.Name)
			}
		}
		fmt.Printf(fmt.Sprintf("| %%%ds ", colWidthName), "Name")
		fmt.Printf(fmt.Sprintf("| %%%ds ", colWidthType), "Type")
		fmt.Printf("\n")

		// Print header delimeter
		fmt.Printf(fmt.Sprintf("+%%%ds", colWidthName), strings.Repeat("-", colWidthName+2))
		fmt.Printf(fmt.Sprintf("+%%%ds", colWidthType), strings.Repeat("-", colWidthType+2))
		fmt.Printf("\n")

		for colName, colType := range colsData {
			fmt.Printf(fmt.Sprintf("| %%%ds ", -colWidthName), colName)
			fmt.Printf(fmt.Sprintf("| %%%ds ", -colWidthType), colType)
			fmt.Printf("\n")
		}

		return nil
	}

	return nil
}
