package runtime

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/chzyer/readline"
	"github.com/npenkov/gcqlsh/internal/action"
	"github.com/npenkov/gcqlsh/internal/db"
)

const ProgramPromptPrefix = "gcqlsh"

func RunInteractiveSession(cks *db.CQLKeyspaceSession) error {
	var completer = readline.NewPrefixCompleter(
		readline.PcItem("use",
			readline.PcItemDynamic(action.ListKeyspaces(cks)),
		),
		readline.PcItem("select",
			readline.PcItem("*",
				readline.PcItem("from",
					readline.PcItemDynamic(action.ListTables(cks)),
				),
			),
		),
		readline.PcItem("insert",
			readline.PcItem("into",
				readline.PcItemDynamic(action.ListTables(cks)),
			),
		),
		readline.PcItem("delete",
			readline.PcItem("from",
				readline.PcItemDynamic(action.ListTables(cks),
					readline.PcItem(";"),
					readline.PcItemDynamic(action.ListColumns(cks, "delete from"),
						readline.PcItem("="),
					),
				),
			),
		),
		readline.PcItem("update",
			readline.PcItemDynamic(action.ListTables(cks),
				readline.PcItem("set",
					readline.PcItemDynamic(action.ListColumns(cks, "update"),
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
				readline.PcItemDynamic(action.ListKeyspaces(cks),
					readline.PcItem(";"),
				),
			),
			readline.PcItem("tables",
				readline.PcItem(";"),
			),
			readline.PcItem("table",
				readline.PcItemDynamic(action.ListTables(cks),
					readline.PcItem(";"),
				),
			),
		),
	)
	home := os.Getenv("HOME")
	config := &readline.Config{
		Prompt:                 fmt.Sprintf("%s:%s> ", ProgramPromptPrefix, cks.ActiveKeyspace),
		HistoryFile:            path.Join(home, ".gcqlsh-history"),
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
		breakLoop, _, closeFunction, err := action.ProcessCommand(cmd, cks)
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
