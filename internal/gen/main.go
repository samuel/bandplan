// Command gen writes the built-in band sets to bandplan.json.
package main

import (
	"log/slog"
	"os"

	"samuelks.com/bandplan"
)

func main() {
	if err := bandplan.SaveFile("bandplan.json", bandplan.Builtins()); err != nil {
		slog.Error("writing bandplan.json", "err", err)
		os.Exit(1)
	}
}
