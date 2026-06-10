package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/wys1203/lexical-cli/internal/lexical"
	"github.com/wys1203/lexical-cli/internal/markdown"
	"github.com/wys1203/lexical-cli/internal/tui"
)

// Version is overridable at build time via -ldflags "-X .../cli.Version=...".
var Version = "dev"

// Exit codes per the goal rubric.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// Env carries the runtime dependencies so Run is testable without touching
// global state.
type Env struct {
	Args        []string // arguments excluding the program name
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	StdinIsTTY  bool
	StdoutIsTTY bool
}

// Run executes the CLI and returns the process exit code.
func Run(env Env) int {
	cfg, err := parseArgs(env.Args)
	if err != nil {
		fmt.Fprintf(env.Stderr, "error: %v\n\n", err)
		fmt.Fprint(env.Stderr, usageText())
		return exitUsage
	}

	if cfg.ShowHelp {
		fmt.Fprint(env.Stdout, usageText())
		return exitOK
	}
	if cfg.ShowVersion {
		fmt.Fprintf(env.Stdout, "lexical-cli %s\n", Version)
		return exitOK
	}

	if !markdown.ValidFormat(cfg.Format) {
		fmt.Fprintf(env.Stderr, "error: unsupported --format %q (want markdown, text, or html)\n", cfg.Format)
		return exitUsage
	}

	if cfg.Command == "tui" && cfg.Input == "" {
		fmt.Fprint(env.Stderr, "error: tui mode requires an input file: lexical-cli tui input.json\n")
		return exitUsage
	}

	data, err := readInput(cfg, env)
	if err != nil {
		fmt.Fprintf(env.Stderr, "error: %v\n", err)
		return exitError
	}

	doc, err := lexical.Parse(data)
	if err != nil {
		fmt.Fprintf(env.Stderr, "error: %v\n", err)
		return exitError
	}

	if cfg.Command == "tui" {
		return runTUI(cfg, doc, env)
	}

	if cfg.ListUnsupported {
		return runListUnsupported(doc, env)
	}
	if cfg.DumpTree {
		fmt.Fprint(env.Stdout, lexical.DumpTree(doc.Root))
		return exitOK
	}

	res, err := markdown.Convert(doc, markdown.Options{
		Format: markdown.Format(cfg.Format),
		Strict: cfg.Strict,
	})
	if err != nil {
		fmt.Fprintf(env.Stderr, "error: %v\n", err)
		return exitError
	}

	for _, w := range res.Warnings {
		fmt.Fprintf(env.Stderr, "warning: %s\n", w)
	}

	if err := writeOutput(cfg, res.Output, env); err != nil {
		fmt.Fprintf(env.Stderr, "error: %v\n", err)
		return exitError
	}
	return exitOK
}

func readInput(cfg *Config, env Env) ([]byte, error) {
	if cfg.Input == "" {
		data, err := io.ReadAll(env.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading stdin: %w", err)
		}
		return data, nil
	}
	data, err := os.ReadFile(cfg.Input)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", cfg.Input, err)
	}
	return data, nil
}

func writeOutput(cfg *Config, content string, env Env) error {
	if cfg.Output == "" {
		_, err := io.WriteString(env.Stdout, content)
		return err
	}
	if err := os.WriteFile(cfg.Output, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", cfg.Output, err)
	}
	return nil
}

func runListUnsupported(doc *lexical.Document, env Env) int {
	items := lexical.ListUnsupported(doc.Root)
	if len(items) == 0 {
		fmt.Fprintln(env.Stdout, "No unsupported nodes found.")
		return exitOK
	}
	for _, it := range items {
		fmt.Fprintf(env.Stdout, "%s\t%s\n", it.Path, it.Type)
	}
	return exitOK
}

func runTUI(cfg *Config, doc *lexical.Document, env Env) int {
	if !env.StdoutIsTTY || !env.StdinIsTTY {
		fmt.Fprint(env.Stderr, "error: tui mode requires an interactive terminal (TTY). "+
			"Use plain CLI mode for non-interactive environments.\n")
		return exitError
	}
	err := tui.Run(tui.Options{
		Doc:        doc,
		Format:     markdown.Format(cfg.Format),
		Strict:     cfg.Strict,
		SourcePath: cfg.Input,
	})
	if err != nil {
		if errors.Is(err, tui.ErrNotInteractive) {
			fmt.Fprint(env.Stderr, "error: tui mode requires an interactive terminal (TTY).\n")
			return exitError
		}
		fmt.Fprintf(env.Stderr, "error: %v\n", err)
		return exitError
	}
	return exitOK
}
