package action

import (
	"fmt"
	"strings"

	"github.com/npenkov/gcqlsh/internal/output"

	"github.com/npenkov/gcqlsh/internal/db"
)

func ProcessCommand(cql string, cks *db.CQLKeyspaceSession) (breakLoop bool, continueLoop bool, errRet error) {
	breakLoop = false
	continueLoop = false
	errRet = nil

	cql = strings.TrimSpace(cql)

	if strings.HasPrefix(cql, "exit") {
		breakLoop = true
		return
	}

	if cql == "" || strings.HasPrefix(cql, "--") {
		continueLoop = true
		return
	}

	if strings.HasPrefix(cql, "use ") || strings.HasPrefix(cql, "USE ") {
		// Remove leading use keyword and trailing semicolon
		scriptKeyspace := strings.TrimPrefix(cql, "use ")
		scriptKeyspace = strings.TrimPrefix(scriptKeyspace, "USE ")
		scriptKeyspace = strings.TrimSuffix(scriptKeyspace, ";")
		scriptKeyspace = strings.TrimPrefix(scriptKeyspace, "\"")
		scriptKeyspace = strings.TrimSuffix(scriptKeyspace, "\"")

		scriptKeyspace = strings.TrimSpace(scriptKeyspace)
		// Create new session as gocql does not support changing keyspaces in session
		s, closef, err := db.NewSession(cks.Host, cks.Port, cks.Username, cks.Password, scriptKeyspace)
		if err == nil {
			if cks.CloseSessionFunc != nil {
				cks.CloseSessionFunc()
			}
			cks.Session = s
			cks.ActiveKeyspace = scriptKeyspace
			cks.CloseSessionFunc = closef
		}
		continueLoop = true
		return
	}

	if strings.HasPrefix(cql, "desc ") || strings.HasPrefix(cql, "DESC ") {
		errRet = describeCmd(cks, cql)
		return
	}

	if strings.HasPrefix(cql, "tracing ") || strings.HasPrefix(cql, "TRACING ") {
		errRet = tracingCmd(cks, cql)
		return
	}
	errRet = execCQL(cks, cql)
	return
}

func execSelectCQL(cks *db.CQLKeyspaceSession, cql string) error {
	tracer := NewTracer(cks)
	defer tracer.Close()
	qry := tracer.Query(cql)
	iter := qry.Iter()

	cntCols := len(iter.Columns())
	cellWidths := make(map[string]int, cntCols)

	res := make(map[string]interface{}, cntCols)
	var rows []map[string]string
	// Print header
	fmt.Println("")

	rowIdx := 0
	for iter.MapScan(res) {
		row := make(map[string]string, cntCols)
		for colIdx := range iter.Columns() {
			col := iter.Columns()[colIdx]
			value := res[col.Name]
			row[col.Name] = printRowValue(col, value)
			if len(row[col.Name]) > cellWidths[col.Name] {
				cellWidths[col.Name] = len(row[col.Name])
			}
		}
		rows = append(rows, row)
		res = make(map[string]interface{}, cntCols)
		rowIdx++
	}
	if err := iter.Close(); err != nil {
		fmt.Printf("error executing cql cql=%q err=%v\n", cql, err)
		return err
	}

	for colIdx := range iter.Columns() {
		col := iter.Columns()[colIdx]
		if len(col.Name) > cellWidths[col.Name] {
			cellWidths[col.Name] = len(col.Name)
		}
		var color func(a ...interface{}) string
		if db.IsPartitionKeyColumn(col, cks.Session) {
			color = output.Red
		} else if db.IsClusterKeyColumn(col, cks.Session) {
			color = output.Blue
		} else {
			color = output.Magenta
		}

		output.PrintColoredColumnVal(cellWidths[col.Name], col.Name, color)
	}
	fmt.Printf("\n")

	// Print header delimeter
	for colIdx := range iter.Columns() {
		col := iter.Columns()[colIdx]
		//output.PrintColumnVal(cellWidths[col.Name]+2, strings.Repeat("-", cellWidths[col.Name]+2))
		output.PrintHeaderSeparator(cellWidths[col.Name])
	}
	fmt.Printf("\n")

	// Print row data
	for rowIdx := range rows {
		row := rows[rowIdx]
		for colIdx := range iter.Columns() {
			col := iter.Columns()[colIdx]

			var color func(a ...interface{}) string
			if db.IsStringColumn(col) {
				color = output.Yellow
			} else {
				color = output.Green
			}
			output.PrintColoredColumnVal(cellWidths[col.Name], row[col.Name], color)
		}
		fmt.Printf("\n")
	}
	if len(rows) == 1 {
		fmt.Printf("\n (%d row)\n", len(rows))
	} else {
		fmt.Printf("\n (%d rows)\n", len(rows))
	}

	return nil
}

func execCQL(cks *db.CQLKeyspaceSession, cql string) error {
	if strings.HasPrefix(cql, "select") || strings.HasPrefix(cql, "SELECT") {
		return execSelectCQL(cks, cql)
	} else {
		tracer := NewTracer(cks)
		defer tracer.Close()
		if err := tracer.Query(cql).RetryPolicy(nil).Exec(); err != nil {
			fmt.Printf("error executing cql cql=%q err=%v\n", cql, err)
			return err
		}
	}

	return nil
}
