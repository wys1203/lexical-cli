// Command lexical-cli converts Meta Lexical EditorState JSON to Markdown (and
// other formats) with both a script-friendly CLI and an interactive TUI.
package main

import (
	"os"

	"github.com/wys1203/lexical-cli/internal/cli"
)

func main() {
	os.Exit(cli.Run(cli.Env{
		Args:        os.Args[1:],
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		StdinIsTTY:  isTerminal(os.Stdin),
		StdoutIsTTY: isTerminal(os.Stdout),
	}))
}

// isTerminal reports whether f is connected to an interactive terminal.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
