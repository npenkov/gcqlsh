package action

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/gocql/gocql"
	"github.com/npenkov/gcqlsh/internal/db"
)

func ProcessCommand(cql string, cks *db.CQLKeyspaceSession) (breakLoop bool, continueLoop bool, closeFunc func(), errRet error) {
	breakLoop = false
	continueLoop = false
	errRet = nil
	closeFunc = func() {}

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
		s, closef, err := db.NewSession(cks.Host, cks.Port, scriptKeyspace)
		if err == nil {
			cks.Session = s
			cks.ActiveKeyspace = scriptKeyspace
			closeFunc = closef
		}
		continueLoop = true
		return
	}

	if strings.HasPrefix(cql, "desc ") || strings.HasPrefix(cql, "DESC ") {
		errRet = describeCmd(cks, cql)
		return
	}
	errRet = execCQL(cks.Session, cql)
	return
}

func execSelectCQL(s *gocql.Session, cql string) error {
	iter := s.Query(cql).Iter()

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

	// Print header
	red := color.New(color.FgRed).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	var addSpaceColor = 0
	if !color.NoColor {
		addSpaceColor = 9
	}

	for colIdx := range iter.Columns() {
		col := iter.Columns()[colIdx]
		if len(col.Name) > cellWidths[col.Name] {
			cellWidths[col.Name] = len(col.Name)
		}
		var colName string
		if db.IsPartitionKeyColumn(col, s) {
			colName = red(col.Name)
		} else if db.IsClusterKeyColumn(col, s) {
			colName = blue(col.Name)
		} else {
			colName = magenta(col.Name)
		}

		fmt.Printf(fmt.Sprintf("| %%%ds ", cellWidths[col.Name]+addSpaceColor), colName)
	}
	fmt.Printf("\n")

	// Print header delimeter
	for colIdx := range iter.Columns() {
		col := iter.Columns()[colIdx]
		fmt.Printf(fmt.Sprintf("+%%%ds", cellWidths[col.Name]+2), strings.Repeat("-", cellWidths[col.Name]+2))
	}
	fmt.Printf("\n")

	// Print row data
	for rowIdx := range rows {
		row := rows[rowIdx]
		for colIdx := range iter.Columns() {
			col := iter.Columns()[colIdx]

			var valueStr string
			width := cellWidths[col.Name] + addSpaceColor
			if db.IsStringColumn(col) {
				valueStr = yellow(row[col.Name])
			} else {
				valueStr = green(row[col.Name])
			}
			fmt.Printf(fmt.Sprintf("| %%%ds ", width), valueStr)
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

func execCQL(s *gocql.Session, cql string) error {
	if strings.HasPrefix(cql, "select") || strings.HasPrefix(cql, "SELECT") {
		return execSelectCQL(s, cql)
	} else {
		if err := s.Query(cql).RetryPolicy(nil).Exec(); err != nil {
			fmt.Printf("error executing cql cql=%q err=%v\n", cql, err)
			return err
		}
	}

	return nil
}
