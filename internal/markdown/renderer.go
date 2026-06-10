// Package markdown converts a parsed Lexical document into Markdown (default),
// plain text, or HTML. It is independent of the CLI and TUI so conversion can
// be unit-tested without terminal interaction.
package markdown

import (
	"fmt"
	"strings"

	"github.com/wys1203/lexical-cli/internal/lexical"
)

// Format selects the output representation.
type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatText     Format = "text"
	FormatHTML     Format = "html"
)

// ValidFormat reports whether s names a supported output format.
func ValidFormat(s string) bool {
	switch Format(s) {
	case FormatMarkdown, FormatText, FormatHTML:
		return true
	}
	return false
}

// Options controls conversion behavior.
type Options struct {
	Format Format
	// Strict makes unknown/unsupported nodes a hard error instead of a warning.
	Strict bool
}

// Warning records a non-fatal conversion issue with the node path where it
// occurred, e.g. root.children[2].children[0].
type Warning struct {
	Path    string
	Message string
}

func (w Warning) String() string {
	return fmt.Sprintf("%s: %s", w.Path, w.Message)
}

// Result is the outcome of a conversion.
type Result struct {
	Output   string
	Warnings []Warning
}

// converter holds per-conversion state while walking the tree.
type converter struct {
	opts     Options
	warnings []Warning
	strict   *strictError
}

type strictError struct {
	path    string
	message string
}

func (e *strictError) Error() string {
	return fmt.Sprintf("unsupported node at %s: %s (strict mode)", e.path, e.message)
}

// Convert renders the document according to opts. In strict mode it returns an
// error on the first unsupported node. In non-strict mode unsupported nodes are
// recorded as warnings and the converter renders best-effort readable text.
func Convert(doc *lexical.Document, opts Options) (Result, error) {
	if opts.Format == "" {
		opts.Format = FormatMarkdown
	}
	if !ValidFormat(string(opts.Format)) {
		return Result{}, fmt.Errorf("unsupported format %q (want markdown, text, or html)", opts.Format)
	}
	c := &converter{opts: opts}

	var output string
	switch opts.Format {
	case FormatHTML:
		output = c.renderHTMLChildren(doc.Root, "root")
	case FormatText:
		output = c.renderBlocksText(doc.Root, "root")
	default:
		output = c.renderBlocks(doc.Root, "root")
	}
	if c.strict != nil {
		return Result{}, c.strict
	}

	output = strings.TrimRight(output, "\n")
	if output != "" {
		output += "\n"
	}
	return Result{Output: output, Warnings: c.warnings}, nil
}

// warn records a warning, or trips strict mode if enabled.
func (c *converter) warn(path, message string) {
	if c.opts.Strict {
		if c.strict == nil {
			c.strict = &strictError{path: path, message: message}
		}
		return
	}
	c.warnings = append(c.warnings, Warning{Path: path, Message: message})
}

// childPath builds the path for the i-th child of parent.
func childPath(parent string, i int) string {
	return fmt.Sprintf("%s.children[%d]", parent, i)
}
