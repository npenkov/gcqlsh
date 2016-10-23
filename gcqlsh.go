package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/gocql/gocql"
)

const ProgramPromptPrefix = "gcqlsh"

type cqlKeyspaceSession struct {
	Host           string
	Port           int
	Session        *gocql.Session
	ActiveKeyspace string
}

func (cks *cqlKeyspaceSession) FetchKeyspaces() ([]string, error) {
	var keyspaceName string
	keyspaces := make([]string, 0)
	iter := cks.Session.Query("select keyspace_name from system.schema_keyspaces").Iter()
	for iter.Scan(&keyspaceName) {
		keyspaces = append(keyspaces, keyspaceName)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return keyspaces, nil
}

func (cks *cqlKeyspaceSession) FetchTables() ([]string, error) {
	var tableName string
	tables := make([]string, 0)
	iter := cks.Session.Query("select columnfamily_name from system.schema_columnfamilies where keyspace_name = ?", cks.ActiveKeyspace).Iter()
	for iter.Scan(&tableName) {
		tables = append(tables, tableName)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return tables, nil
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func printVal(col gocql.ColumnInfo, value interface{}) string {
	typeMod := "s"
	t := col.TypeInfo.Type()
	switch t {
	case gocql.TypeCustom:
	case gocql.TypeAscii:
	case gocql.TypeBigInt:
		typeMod = "d"
	case gocql.TypeBlob:
	case gocql.TypeBoolean:
		typeMod = "t"
	case gocql.TypeCounter:
	case gocql.TypeDecimal:
		typeMod = "d"
	case gocql.TypeDouble:
		typeMod = "f"
	case gocql.TypeFloat:
		typeMod = "f"
	case gocql.TypeInt:
		typeMod = "d"
	case gocql.TypeText:
	case gocql.TypeTimestamp:
	case gocql.TypeUUID:
	case gocql.TypeVarchar:
	case gocql.TypeVarint:
	case gocql.TypeTimeUUID:
	case gocql.TypeInet:
	case gocql.TypeDate:
	case gocql.TypeTime:
	case gocql.TypeSmallInt:
		typeMod = "d"
	case gocql.TypeTinyInt:
		typeMod = "d"
	case gocql.TypeList:
	case gocql.TypeMap:
	case gocql.TypeSet:
	case gocql.TypeUDT:
	case gocql.TypeTuple:
	}
	val := fmt.Sprintf("%"+typeMod, value)
	return val
}

func isPrimaryKeyColumn(col gocql.ColumnInfo, s *gocql.Session) bool {
	km, _ := s.KeyspaceMetadata(col.Keyspace)
	tm := km.Tables[col.Table]
	for pkc := range tm.PartitionKey {
		if tm.PartitionKey[pkc].Name == col.Name {
			return true
		}
	}
	return false
}

func isStringColumn(col gocql.ColumnInfo) bool {
	t := col.TypeInfo.Type()
	switch t {
	case gocql.TypeAscii:
		return true
	case gocql.TypeText:
		return true
	case gocql.TypeVarchar:
		return true
	default:
		return false
	}
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
			row[col.Name] = printVal(col, value)
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
		if isPrimaryKeyColumn(col, s) {
			colName = red(col.Name)
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
			if isStringColumn(col) {
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

func createCluster(host string, port int, keyspace string) *gocql.ClusterConfig {
	cluster := gocql.NewCluster(gocql.JoinHostPort(host, port))

	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.One
	cluster.Timeout = 10 * time.Second
	cluster.MaxWaitSchemaAgreement = 2 * time.Minute
	cluster.ProtoVersion = 3
	cluster.IgnorePeerAddr = true
	cluster.DisableInitialHostLookup = true

	cluster.NumConns = 3

	return cluster
}

func createSession(cluster *gocql.ClusterConfig) (*gocql.Session, func(), error) {
	session, err := cluster.CreateSession()
	return session, func() {
		session.Close()
	}, err
}

func describeCmd(cks *cqlKeyspaceSession, cmd string) error {
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
		return nil
	}

	return nil
}

func processCommand(cql string, cks *cqlKeyspaceSession) (breakLoop bool, continueLoop bool, closeFunc func(), errRet error) {
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
		s, closef, err := createSession(createCluster(cks.Host, cks.Port, scriptKeyspace))
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

func processScriptFile(scriptFile string, cks *cqlKeyspaceSession, printCQL bool, failOnError bool) {
	file, err := os.Open(scriptFile)
	if err != nil {
		fmt.Printf("error opening file %s: %v\n", scriptFile, err)
		os.Exit(-2)
	}

	r := bufio.NewReader(file)
	for cql, e := r.ReadString(";"[0]); e == nil; {
		breakLoop, continueLoop, closeFunc, err := processCommand(cql, cks)
		defer closeFunc()
		if printCQL {
			fmt.Println(cql)
		}
		if breakLoop {
			break
		}
		if continueLoop {
			continue
		}
		if err != nil {
			fmt.Println(err)
		}

		if err != nil && failOnError {
			os.Exit(-1)
		}
		cql, e = r.ReadString(";"[0])
	}
}

func listKeyspaces(cks *cqlKeyspaceSession) func(string) []string {
	return func(line string) []string {
		keyspaces, _ := cks.FetchKeyspaces()
		return keyspaces
	}
}

func listTables(cks *cqlKeyspaceSession) func(string) []string {
	return func(line string) []string {
		tables, _ := cks.FetchTables()
		return tables
	}
}

func listColumns(cks *cqlKeyspaceSession) func(string) []string {
	return func(line string) []string {
		// get table from the line and fetch the columns
		tables, _ := cks.FetchTables()
		return tables
	}
}

func runInteractiveSession(cks *cqlKeyspaceSession) error {
	var completer = readline.NewPrefixCompleter(
		readline.PcItem("use",
			readline.PcItemDynamic(listKeyspaces(cks)),
		),
		readline.PcItem("select",
			readline.PcItem("*",
				readline.PcItem("from",
					readline.PcItemDynamic(listTables(cks)),
				),
			),
		),
		readline.PcItem("insert",
			readline.PcItem("into",
				readline.PcItemDynamic(listTables(cks)),
			),
		),
		readline.PcItem("delete",
			readline.PcItem("from",
				readline.PcItemDynamic(listTables(cks),
					readline.PcItem(";"),
					readline.PcItemDynamic(listColumns(cks),
						readline.PcItem("="),
					),
				),
			),
		),
		readline.PcItem("update",
			readline.PcItemDynamic(listTables(cks),
				readline.PcItem("set",
					readline.PcItemDynamic(listColumns(cks),
						readline.PcItem("="),
					),
				),
			),
		),
		readline.PcItem("desc",
			readline.PcItem("keyspaces",
				readline.PcItem(";"),
			),
			readline.PcItem("keyspace",
				readline.PcItemDynamic(listKeyspaces(cks),
					readline.PcItem(";"),
				),
			),
			readline.PcItem("tables",
				readline.PcItem(";"),
			),
			readline.PcItem("table",
				readline.PcItemDynamic(listTables(cks),
					readline.PcItem(";"),
				),
			),
		),
	)
	config := &readline.Config{
		Prompt:                 fmt.Sprintf("%s:%s> ", ProgramPromptPrefix, cks.ActiveKeyspace),
		HistoryFile:            "/tmp/.readline-multiline",
		DisableAutoSaveHistory: true,
		AutoComplete:           completer,
		InterruptPrompt:        "^C",
	}

	rl, err := readline.NewEx(config)
	rl.SetPrompt(fmt.Sprintf("%s:%s> ", ProgramPromptPrefix, cks.ActiveKeyspace))
	if err != nil {
		return err
	}
	defer rl.Close()

	var cmds []string
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		cmds = append(cmds, line)
		if !strings.HasSuffix(line, ";") {
			rl.SetPrompt(">>> ")
			continue
		}
		cmd := strings.Join(cmds, " ")
		cmds = cmds[:0]
		breakLoop, _, closeFunction, err := processCommand(cmd, cks)
		defer closeFunction()

		if err != nil {
			fmt.Println(err)
		}
		if breakLoop {
			break
		}
		rl.SetPrompt(fmt.Sprintf("%s:%s> ", ProgramPromptPrefix, cks.ActiveKeyspace))
		rl.SaveHistory(cmd)
	}
	return nil
}
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var host string
	var port int
	var keyspace string
	var printConf bool
	var printCQL bool
	var failOnError bool
	var noColor bool
	var scriptFile string

	flag.StringVar(&host, "host", "127.0.0.1", "Cassandra host to connect to")
	flag.IntVar(&port, "port", 9042, "Cassandra RPC port")
	flag.BoolVar(&printConf, "print-confirmation", false, "Print 'ok' on successfuly executed cql statement from the file")
	flag.BoolVar(&printCQL, "print-cql", false, "Print Statements that are executed from a file")
	flag.BoolVar(&failOnError, "fail-on-error", false, "Stop execution if statement from file fails.")
	flag.BoolVar(&noColor, "no-color", false, "Console without colors")
	flag.StringVar(&keyspace, "k", "system", "Default keyspace to connect to")
	flag.StringVar(&scriptFile, "f", "", "Execute file containing cql statements instead of having interacive session")

	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "%s [options] CQL_SCRIPT_FILE\n", os.Args[0])
		flag.PrintDefaults()
	}

	color.NoColor = noColor

	// connect to the cluster
	session, closeFunc, sesErr := createSession(createCluster(host, port, keyspace))
	if sesErr != nil {
		fmt.Println(sesErr)
		os.Exit(-1)
	}
	defer closeFunc()

	keyspaceSession := &cqlKeyspaceSession{Session: session, ActiveKeyspace: keyspace, Host: host, Port: port}

	if scriptFile == "" {
		runInteractiveSession(keyspaceSession)
	} else {
		color.NoColor = true
		processScriptFile(scriptFile, keyspaceSession, printCQL, failOnError)
	}
}
