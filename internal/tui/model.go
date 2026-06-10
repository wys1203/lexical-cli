package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/wys1203/lexical-cli/internal/lexical"
	"github.com/wys1203/lexical-cli/internal/markdown"
)

type focusArea int

const (
	focusTree focusArea = iota
	focusPreview
)

type model struct {
	opts Options

	tree     []lexical.TreeLine
	preview  []string // preview split into logical lines
	warnings []markdown.Warning

	width, height int

	focus         focusArea
	treeCursor    int
	treeOffset    int
	previewOffset int

	showWarnings bool
	showHelp     bool
	status       string
}

func newModel(opts Options) (*model, error) {
	if opts.Doc == nil || opts.Doc.Root == nil {
		return nil, lexical.Validate(opts.Doc)
	}
	if opts.Format == "" {
		opts.Format = markdown.FormatMarkdown
	}

	res, err := markdown.Convert(opts.Doc, markdown.Options{
		Format: opts.Format,
		Strict: opts.Strict,
	})
	if err != nil {
		return nil, err
	}

	m := &model{
		opts:     opts,
		tree:     lexical.Flatten(opts.Doc.Root),
		preview:  strings.Split(strings.TrimRight(res.Output, "\n"), "\n"),
		warnings: res.Warnings,
		focus:    focusTree,
	}
	return m, nil
}

func (m *model) Init() tea.Cmd { return nil }
