package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/wys1203/lexical-cli/internal/markdown"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Help and warnings overlays close on any key.
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	switch key {
	case "q", "ctrl+c", "esc":
		return m, tea.Quit

	case "?":
		m.showHelp = true
		return m, nil

	case "w":
		m.showWarnings = !m.showWarnings
		return m, nil

	case "tab":
		if m.focus == focusTree {
			m.focus = focusPreview
		} else {
			m.focus = focusTree
		}
		return m, nil

	case "s":
		m.save()
		return m, nil

	case "up", "k":
		m.moveUp()
		return m, nil

	case "down", "j":
		m.moveDown()
		return m, nil

	case "pgup":
		for i := 0; i < 5; i++ {
			m.moveUp()
		}
		return m, nil

	case "pgdown":
		for i := 0; i < 5; i++ {
			m.moveDown()
		}
		return m, nil
	}
	return m, nil
}

func (m *model) moveUp() {
	if m.focus == focusTree {
		if m.treeCursor > 0 {
			m.treeCursor--
		}
		return
	}
	if m.previewOffset > 0 {
		m.previewOffset--
	}
}

func (m *model) moveDown() {
	if m.focus == focusTree {
		if m.treeCursor < len(m.tree)-1 {
			m.treeCursor++
		}
		return
	}
	if m.previewOffset < len(m.preview)-1 {
		m.previewOffset++
	}
}

// save writes the current preview content to a file derived from the source
// path and chosen format.
func (m *model) save() {
	target := m.outputPath()
	content := strings.Join(m.preview, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		m.status = fmt.Sprintf("save failed: %v", err)
		return
	}
	m.status = fmt.Sprintf("saved %d lines to %s", len(m.preview), target)
}

func (m *model) outputPath() string {
	ext := ".md"
	switch m.opts.Format {
	case markdown.FormatText:
		ext = ".txt"
	case markdown.FormatHTML:
		ext = ".html"
	}
	if m.opts.SourcePath == "" {
		return "output" + ext
	}
	base := strings.TrimSuffix(m.opts.SourcePath, filepath.Ext(m.opts.SourcePath))
	return base + ext
}
