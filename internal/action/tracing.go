package action

import (
	"fmt"
	"strings"

	"github.com/npenkov/gcqlsh/internal/db"
	"github.com/npenkov/gcqlsh/internal/output"
)

func tracingCmd(cks *db.CQLKeyspaceSession, cmd string) error {
	desc := strings.TrimPrefix(strings.TrimPrefix(cmd, "tracing "), "TRACING ")
	desc = strings.TrimSpace(desc)
	if strings.HasPrefix(desc, "on") || strings.HasPrefix(desc, "ON") {
		if cks.TracingEnabled {
			output.PrintError("Tracing is already enabled. Use TRACING OFF to disable.")
			return nil
		}
		cks.EnableTracing()
		fmt.Print("Now Tracing is enabled.\n")
		return nil
	}

	if strings.HasPrefix(desc, "off") || strings.HasPrefix(desc, "OFF") {
		if !cks.TracingEnabled {
			output.PrintError("Tracing is not enabled.")
			return nil
		}
		cks.DisableTracing()
		fmt.Print("Disabled Tracing.\n")
		return nil
	}

	output.PrintError("Improper tracing command.")

	return nil
}
