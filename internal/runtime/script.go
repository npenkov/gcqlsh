package runtime

import (
	"bufio"
	"fmt"
	"os"

	"github.com/npenkov/gcqlsh/internal/action"
	"github.com/npenkov/gcqlsh/internal/db"
)

func ProcessScriptFile(scriptFile string, cks *db.CQLKeyspaceSession, printCQL bool, failOnError bool) {
	file, err := os.Open(scriptFile)
	if err != nil {
		fmt.Printf("error opening file %s: %v\n", scriptFile, err)
		os.Exit(-2)
	}

	r := bufio.NewReader(file)
	for {
		cql, e := r.ReadString(';')
		if e != nil {
			break
		}
		breakLoop, continueLoop, closeFunc, err := action.ProcessCommand(cql, cks)
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
	}
}
