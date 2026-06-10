package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	focusedBorder = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("12"))
	blurredBorder = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("12"))
	warnStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	hintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func (m *model) View() string {
	width, height := m.width, m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	if m.showHelp {
		return m.helpView(width, height)
	}

	header := titleStyle.Render("lexical-cli — Lexical Inspector")

	footer := m.footerView(width)
	footerHeight := lipgloss.Height(footer)

	// Account for header (1) and footer.
	bodyHeight := height - 1 - footerHeight
	if bodyHeight < 3 {
		bodyHeight = 3
	}

	leftWidth := width * 2 / 5
	if leftWidth < 20 {
		leftWidth = 20
	}
	if leftWidth > width-10 {
		leftWidth = width - 10
	}
	rightWidth := width - leftWidth - 4 // borders/gutter

	// Inner content height excludes the pane's top/bottom border (2).
	innerHeight := bodyHeight - 2
	if innerHeight < 1 {
		innerHeight = 1
	}

	tree := m.renderTreePane(leftWidth-2, innerHeight)
	preview := m.renderPreviewPane(rightWidth-2, innerHeight)

	leftStyle, rightStyle := blurredBorder, blurredBorder
	if m.focus == focusTree {
		leftStyle = focusedBorder
	} else {
		rightStyle = focusedBorder
	}

	left := leftStyle.Width(leftWidth - 2).Height(innerHeight).Render(tree)
	right := rightStyle.Width(rightWidth - 2).Height(innerHeight).Render(preview)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	return strings.Join([]string{header, body, footer}, "\n")
}

func (m *model) renderTreePane(width, height int) string {
	title := "Lexical Tree"
	m.clampTreeOffset(height - 1)

	var lines []string
	lines = append(lines, hintStyle.Render(title))

	visible := height - 1
	for i := 0; i < visible; i++ {
		idx := m.treeOffset + i
		if idx >= len(m.tree) {
			break
		}
		line := m.tree[idx]
		text := strings.Repeat("  ", line.Depth) + line.Label
		text = truncateDisplay(text, width)
		if idx == m.treeCursor {
			text = cursorStyle.Render(padRight(text, width))
		}
		lines = append(lines, text)
	}
	return strings.Join(lines, "\n")
}

func (m *model) renderPreviewPane(width, height int) string {
	visible := height - 1
	if m.showWarnings {
		return m.renderWarningsPane(width, height)
	}

	title := fmt.Sprintf("%s Preview", titleCase(string(m.opts.Format)))
	m.clampPreviewOffset(visible)

	var lines []string
	lines = append(lines, hintStyle.Render(title))
	for i := 0; i < visible; i++ {
		idx := m.previewOffset + i
		if idx >= len(m.preview) {
			break
		}
		lines = append(lines, truncateDisplay(m.preview[idx], width))
	}
	return strings.Join(lines, "\n")
}

func (m *model) renderWarningsPane(width, height int) string {
	var lines []string
	lines = append(lines, warnStyle.Render(fmt.Sprintf("Warnings (%d)", len(m.warnings))))
	if len(m.warnings) == 0 {
		lines = append(lines, "No conversion warnings.")
	}
	visible := height - 1
	for i := 0; i < len(m.warnings) && i < visible; i++ {
		w := m.warnings[i]
		lines = append(lines, truncateDisplay(fmt.Sprintf("• %s: %s", w.Path, w.Message), width))
	}
	return strings.Join(lines, "\n")
}

func (m *model) footerView(width int) string {
	warnSummary := "no warnings"
	if len(m.warnings) > 0 {
		warnSummary = warnStyle.Render(fmt.Sprintf("%d warning(s)", len(m.warnings)))
	}
	status := m.status
	if status == "" {
		status = warnSummary
	}

	hints := hintStyle.Render("q quit · tab focus · ↑/↓ move · w warnings · s save · ? help")
	statusLine := truncateDisplay(status, width)
	return statusLine + "\n" + hints
}

func (m *model) helpView(width, height int) string {
	help := strings.Join([]string{
		titleStyle.Render("lexical-cli TUI — Help"),
		"",
		"  ↑ / k        move selection up (tree) or scroll up (preview)",
		"  ↓ / j        move selection down or scroll down",
		"  tab          switch focus between tree and preview",
		"  w            toggle the warnings pane",
		"  s            save/export the preview to a file",
		"  ?            toggle this help",
		"  q / esc      quit",
		"",
		hintStyle.Render("Press any key to close help."),
	}, "\n")
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box.Render(help))
}

func (m *model) clampTreeOffset(visible int) {
	if visible < 1 {
		visible = 1
	}
	if m.treeCursor < m.treeOffset {
		m.treeOffset = m.treeCursor
	}
	if m.treeCursor >= m.treeOffset+visible {
		m.treeOffset = m.treeCursor - visible + 1
	}
	if m.treeOffset < 0 {
		m.treeOffset = 0
	}
}

func (m *model) clampPreviewOffset(visible int) {
	max := len(m.preview) - visible
	if max < 0 {
		max = 0
	}
	if m.previewOffset > max {
		m.previewOffset = max
	}
	if m.previewOffset < 0 {
		m.previewOffset = 0
	}
}

func truncateDisplay(s string, width int) string {
	if width <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	return string(r[:width-1]) + "…"
}

func padRight(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(r))
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
