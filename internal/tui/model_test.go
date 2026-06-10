package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/wys1203/lexical-cli/internal/lexical"
	"github.com/wys1203/lexical-cli/internal/markdown"
)

func loadDoc(t *testing.T, name string) *lexical.Document {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := lexical.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestModelBuildsTreeAndPreview(t *testing.T) {
	m, err := newModel(Options{Doc: loadDoc(t, "formats.json"), Format: markdown.FormatMarkdown})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.tree) == 0 {
		t.Error("expected tree lines")
	}
	if len(m.preview) == 0 {
		t.Error("expected preview lines")
	}
	joined := strings.Join(m.preview, "\n")
	if !strings.Contains(joined, "# Title") {
		t.Errorf("preview missing heading: %q", joined)
	}
}

func TestModelCollectsWarnings(t *testing.T) {
	m, err := newModel(Options{Doc: loadDoc(t, "unsupported.json"), Format: markdown.FormatMarkdown})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.warnings) != 1 {
		t.Fatalf("want 1 warning, got %d", len(m.warnings))
	}
}

func TestModelStrictErrors(t *testing.T) {
	_, err := newModel(Options{Doc: loadDoc(t, "unsupported.json"), Format: markdown.FormatMarkdown, Strict: true})
	if err == nil {
		t.Fatal("strict model should fail to build on unsupported node")
	}
}

func TestModelFocusToggle(t *testing.T) {
	m, _ := newModel(Options{Doc: loadDoc(t, "formats.json")})
	if m.focus != focusTree {
		t.Fatal("default focus should be tree")
	}
	m.handleKey(keyMsg("tab"))
	if m.focus != focusPreview {
		t.Error("tab should switch focus to preview")
	}
	m.handleKey(keyMsg("tab"))
	if m.focus != focusTree {
		t.Error("tab should switch back to tree")
	}
}

func TestModelNavigation(t *testing.T) {
	m, _ := newModel(Options{Doc: loadDoc(t, "lists.json")})
	start := m.treeCursor
	m.handleKey(keyMsg("down"))
	if m.treeCursor != start+1 {
		t.Errorf("down should advance cursor, got %d", m.treeCursor)
	}
	m.handleKey(keyMsg("up"))
	if m.treeCursor != start {
		t.Errorf("up should restore cursor, got %d", m.treeCursor)
	}
}

func TestModelQuit(t *testing.T) {
	m, _ := newModel(Options{Doc: loadDoc(t, "formats.json")})
	_, cmd := m.handleKey(keyMsg("q"))
	if cmd == nil {
		t.Fatal("q should return a quit command")
	}
	if msg := cmd(); msg == nil {
		t.Error("quit command should produce a message")
	}
}

func TestModelSaveExport(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "doc.json")
	data, _ := os.ReadFile(filepath.Join("..", "..", "testdata", "formats.json"))
	if err := os.WriteFile(src, data, 0o644); err != nil {
		t.Fatal(err)
	}
	m, _ := newModel(Options{Doc: loadDoc(t, "formats.json"), Format: markdown.FormatMarkdown, SourcePath: src})
	m.handleKey(keyMsg("s"))

	target := filepath.Join(dir, "doc.md")
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected exported file: %v", err)
	}
	if !strings.Contains(string(got), "# Title") {
		t.Errorf("exported content = %q", got)
	}
	if !strings.Contains(m.status, "saved") {
		t.Errorf("status = %q", m.status)
	}
}

func TestModelHelpToggle(t *testing.T) {
	m, _ := newModel(Options{Doc: loadDoc(t, "formats.json")})
	m.handleKey(keyMsg("?"))
	if !m.showHelp {
		t.Error("? should show help")
	}
	m.handleKey(keyMsg("x")) // any key closes help
	if m.showHelp {
		t.Error("any key should close help")
	}
}

func TestModelViewRenders(t *testing.T) {
	m, _ := newModel(Options{Doc: loadDoc(t, "formats.json")})
	m.width, m.height = 100, 30
	view := m.View()
	if !strings.Contains(view, "Lexical Tree") {
		t.Errorf("view missing tree title")
	}
	if !strings.Contains(view, "quit") {
		t.Errorf("view missing footer hints")
	}
}

// keyMsg builds a tea.KeyMsg from a key name for tests.
func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}
