package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/fatih/color"

	"github.com/npenkov/gcqlsh/internal/db"
	r "github.com/npenkov/gcqlsh/internal/runtime"
)

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
	session, closeFunc, sesErr := db.NewSession(host, port, keyspace)
	if sesErr != nil {
		fmt.Println(sesErr)
		os.Exit(-1)
	}

	keyspaceSession := &db.CQLKeyspaceSession{
		Session: session, ActiveKeyspace: keyspace, Host: host, Port: port, CloseSessionFunc: closeFunc}

	defer func() {
		keyspaceSession.CloseSessionFunc()
	}()

	if scriptFile == "" {
		if err := r.RunInteractiveSession(keyspaceSession); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	} else {
		color.NoColor = true
		r.ProcessScriptFile(scriptFile, keyspaceSession, printCQL, failOnError)
	}
}
