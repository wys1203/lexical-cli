package cli

func usageText() string {
	return `lexical-cli — convert Meta Lexical EditorState JSON to Markdown.

USAGE:
  lexical-cli [flags] [input.json]
  lexical-cli tui [flags] input.json
  cat input.json | lexical-cli [flags]

FLAGS:
  -o, --output <file>   Write output to <file> instead of stdout.
  -f, --format <fmt>    Output format: markdown (default), text, or html.
      --strict          Fail on unsupported nodes instead of warning.
      --list-unsupported  List unknown node types and their paths, then exit.
      --dump-tree       Print a simplified Lexical tree for debugging, then exit.
  -v, --version         Print version information.
  -h, --help            Print this help.

EXAMPLES:
  lexical-cli input.json
  lexical-cli --output out.md input.json
  lexical-cli --format html input.json
  cat input.json | lexical-cli
  lexical-cli tui input.json

EXIT CODES:
  0  success
  1  invalid input, conversion failure, or I/O error
  2  invalid CLI usage
`
}
