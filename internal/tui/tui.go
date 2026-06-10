// Package tui provides an interactive terminal interface for inspecting a
// Lexical document: a node tree, a rendered preview, conversion warnings, and
// export. It depends on the lexical and markdown packages but those packages do
// not depend on it, so conversion stays testable without a terminal.
package tui

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/wys1203/lexical-cli/internal/lexical"
	"github.com/wys1203/lexical-cli/internal/markdown"
)

// ErrNotInteractive is returned when the TUI cannot run because there is no
// interactive terminal available.
var ErrNotInteractive = errors.New("tui requires an interactive terminal")

// Options configures a TUI session.
type Options struct {
	Doc        *lexical.Document
	Format     markdown.Format
	Strict     bool
	SourcePath string
}

// Run starts the TUI event loop and blocks until the user quits.
func Run(opts Options) error {
	m, err := newModel(opts)
	if err != nil {
		return err
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
