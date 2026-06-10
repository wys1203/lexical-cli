// Package cli wires argument parsing, input/output, and exit codes around the
// lexical parser and renderer.
package cli

import "fmt"

// Config is the parsed command-line configuration.
type Config struct {
	Command string // "", "tui"
	Input   string // path to input file; "" means stdin
	Output  string // path to output file; "" means stdout

	Format string
	Strict bool

	ShowVersion bool
	ShowHelp    bool

	ListUnsupported bool
	DumpTree        bool
}

// usageError signals invalid CLI usage and maps to exit code 2.
type usageError struct{ msg string }

func (e *usageError) Error() string { return e.msg }

func usagef(format string, args ...any) error {
	return &usageError{msg: fmt.Sprintf(format, args...)}
}

// parseArgs parses the argument slice (excluding the program name) into Config.
// It is tolerant of flag/positional ordering.
func parseArgs(args []string) (*Config, error) {
	cfg := &Config{Format: "markdown"}

	i := 0
	// Optional leading subcommand.
	if len(args) > 0 && args[0] == "tui" {
		cfg.Command = "tui"
		i = 1
	}

	var positionals []string
	for ; i < len(args); i++ {
		arg := args[i]

		// Split --flag=value form.
		name, inlineVal, hasInline := splitFlag(arg)

		takeValue := func() (string, error) {
			if hasInline {
				return inlineVal, nil
			}
			if i+1 >= len(args) {
				return "", usagef("flag %s requires a value", name)
			}
			i++
			return args[i], nil
		}

		switch name {
		case "-o", "--output":
			v, err := takeValue()
			if err != nil {
				return nil, err
			}
			cfg.Output = v
		case "--format", "-f":
			v, err := takeValue()
			if err != nil {
				return nil, err
			}
			cfg.Format = v
		case "--strict":
			cfg.Strict = true
		case "--version", "-v":
			cfg.ShowVersion = true
		case "--help", "-h":
			cfg.ShowHelp = true
		case "--list-unsupported":
			cfg.ListUnsupported = true
		case "--dump-tree":
			cfg.DumpTree = true
		case "--":
			// Everything after -- is positional.
			positionals = append(positionals, args[i+1:]...)
			i = len(args)
		default:
			if len(arg) > 1 && arg[0] == '-' {
				return nil, usagef("unknown flag %q", arg)
			}
			positionals = append(positionals, arg)
		}
	}

	if len(positionals) > 1 {
		return nil, usagef("expected at most one input file, got %d", len(positionals))
	}
	if len(positionals) == 1 {
		cfg.Input = positionals[0]
	}

	return cfg, nil
}

// splitFlag splits "--flag=value" into ("--flag", "value", true). For other
// forms it returns (arg, "", false).
func splitFlag(arg string) (name, value string, hasInline bool) {
	if len(arg) >= 2 && arg[0] == '-' {
		for j := 1; j < len(arg); j++ {
			if arg[j] == '=' {
				return arg[:j], arg[j+1:], true
			}
		}
	}
	return arg, "", false
}
