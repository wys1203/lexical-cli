# lexical-cli

A Go-native command-line tool that converts pure Meta [Lexical](https://lexical.dev)
`EditorState` JSON into Markdown (and text or HTML). It ships as a single
standalone binary — **no Node.js, npm, or `@lexical/markdown` required at
runtime** — and includes an optional terminal UI for inspecting the Lexical
tree, previewing output, reviewing conversion warnings, and exporting.

Markdown is the default export format because it is readable, diffable, and
useful for documentation and AI workflows. Markdown output is **not lossless**;
the original Lexical JSON remains the most faithful storage format.

## Install

### From source

```sh
go install github.com/wys1203/lexical-cli/cmd/lexical-cli@latest
```

### Build locally

```sh
git clone https://github.com/wys1203/lexical-cli
cd lexical-cli
go build -o lexical-cli ./cmd/lexical-cli
```

### Homebrew

```sh
brew install wys1203/tap/lexical-cli
```

## Usage

```text
lexical-cli [flags] [input.json]
lexical-cli tui [flags] input.json
cat input.json | lexical-cli [flags]
```

### Flags

| Flag | Description |
| --- | --- |
| `-o`, `--output <file>` | Write output to a file instead of stdout. |
| `-f`, `--format <fmt>` | Output format: `markdown` (default), `text`, or `html`. |
| `--strict` | Fail on unsupported nodes instead of warning. |
| `--list-unsupported` | List unknown node types and their paths, then exit. |
| `--dump-tree` | Print a simplified Lexical tree for debugging, then exit. |
| `-v`, `--version` | Print version information. |
| `-h`, `--help` | Print usage. |

### Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Successful conversion or TUI export. |
| `1` | Invalid JSON, invalid root shape, unsupported node in strict mode, or I/O failure. |
| `2` | Invalid CLI usage (unknown flag, bad `--format`, missing argument). |

## Examples

Convert a file to Markdown on stdout:

```sh
lexical-cli input.json
```

Write Markdown to a file:

```sh
lexical-cli --output output.md input.json
```

Pipe JSON through stdin:

```sh
cat input.json | lexical-cli
```

Render HTML:

```sh
lexical-cli --format html input.json
```

Fail fast on anything unsupported:

```sh
lexical-cli --strict input.json
```

Inspect what a document contains without converting:

```sh
lexical-cli --dump-tree input.json
lexical-cli --list-unsupported input.json
```

### Example conversion

Input (abbreviated Lexical JSON):

```json
{
  "root": {
    "type": "root",
    "children": [
      { "type": "heading", "tag": "h1",
        "children": [{ "type": "text", "format": 0, "text": "Title" }] },
      { "type": "paragraph",
        "children": [{ "type": "text", "format": 1, "text": "bold" }] }
    ]
  }
}
```

Output:

```markdown
# Title

**bold**
```

## TUI

```sh
lexical-cli tui input.json
```

Opens an interactive inspector:

```text
Lexical Tree                    Markdown Preview
-------------------------------- ----------------------------------
root                             # Heading
|-- heading                      
|   `-- text "Heading"           Paragraph with **bold** text.
`-- paragraph                    
    `-- text "Paragraph..."      - Item one

q quit · tab focus · ↑/↓ move · w warnings · s save · ? help
```

### Key bindings

| Key | Action |
| --- | --- |
| `q` / `esc` | Quit. |
| `tab` | Switch focus between tree and preview. |
| `↑` / `↓` (`k` / `j`) | Move selection in tree or scroll preview. |
| `pgup` / `pgdown` | Page through the focused pane. |
| `w` | Toggle the warnings pane. |
| `s` | Save/export the preview to a file. |
| `?` | Toggle help. |

Export writes the current preview to a file derived from the input name and
chosen format (e.g. `input.json` → `input.md`). The TUI detects non-interactive
(non-TTY) environments and exits with a clear error rather than hanging; plain
CLI mode continues to work everywhere.

## Supported nodes

**Block:** `root`, `paragraph`, `heading` (`h1`–`h6`), `quote`, `list`
(bullet / number / check), `listitem`, `code`, `horizontalrule`, `table` /
`tablerow` / `tablecell`, `image`.

**Inline:** `text`, `linebreak`, `tab`, `link` / `autolink`, `hashtag`,
`code-highlight`.

**Inline formatting** (Lexical `format` bitmask): bold, italic, strikethrough,
underline (rendered as `<u>…</u>` in Markdown), and inline code.

## Known limitations

- Markdown output is not lossless and does not round-trip back to identical
  Lexical JSON.
- Underline has no standard Markdown syntax and is emitted as inline HTML.
- CMS-wrapped shapes (e.g. Payload documents with `content.root`) are **not**
  supported — input must have a top-level `root`.
- Selection, editor history, and collaboration metadata are ignored.
- In non-strict mode (default), unknown nodes are skipped with a warning and any
  readable child text is preserved. Use `--strict` to fail instead.

## Development

```sh
go fmt ./...
go vet ./...
go test ./...
go build ./cmd/lexical-cli
```

Conversion snapshots live in `testdata/golden/`. Regenerate them after an
intentional output change:

```sh
go test ./internal/markdown -update
```

The parser (`internal/lexical`) and renderer (`internal/markdown`) are
independent of the TUI (`internal/tui`), so conversion is fully testable without
a terminal.

## License

MIT
