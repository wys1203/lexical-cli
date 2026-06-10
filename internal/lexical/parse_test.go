package lexical

import (
	"os"
	"path/filepath"
	"testing"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatalf("reading fixture %s: %v", name, err)
	}
	return data
}

func TestParseValidParagraph(t *testing.T) {
	doc, err := Parse(readFixture(t, "paragraph.json"))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if doc.Root.Type != "root" {
		t.Fatalf("root type = %q, want root", doc.Root.Type)
	}
	if len(doc.Root.Children) != 1 {
		t.Fatalf("root children = %d, want 1", len(doc.Root.Children))
	}
	p := doc.Root.Children[0]
	if p.Type != "paragraph" {
		t.Fatalf("child type = %q, want paragraph", p.Type)
	}
	if got := p.Children[0].Text; got != "Hello world." {
		t.Fatalf("text = %q, want %q", got, "Hello world.")
	}
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := Parse([]byte(`{ "root": { broken `))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseEmptyInput(t *testing.T) {
	if _, err := Parse([]byte("   \n  ")); err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

func TestParseMissingRoot(t *testing.T) {
	_, err := Parse([]byte(`{ "notroot": {} }`))
	if err == nil {
		t.Fatal("expected error for missing root, got nil")
	}
}

func TestParseWrongRootType(t *testing.T) {
	_, err := Parse([]byte(`{ "root": { "type": "paragraph", "children": [] } }`))
	if err == nil {
		t.Fatal("expected error for wrong root type, got nil")
	}
}

func TestTextFormatBitmask(t *testing.T) {
	doc, err := Parse(readFixture(t, "formats.json"))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	para := doc.Root.Children[1]
	bold := para.Children[0]
	if !bold.HasFormat(FormatBold) {
		t.Errorf("expected bold format set, TextFormat=%d", bold.TextFormat)
	}
	boldItalic := para.Children[len(para.Children)-1]
	if !boldItalic.HasFormat(FormatBold) || !boldItalic.HasFormat(FormatItalic) {
		t.Errorf("expected bold+italic, TextFormat=%d", boldItalic.TextFormat)
	}
}

func TestElementAlignmentFormat(t *testing.T) {
	// A paragraph with a string "center" format must not be misread as a bitmask.
	doc, err := Parse([]byte(`{"root":{"type":"root","children":[
		{"type":"paragraph","format":"center","children":[
			{"type":"text","format":0,"text":"x"}]}]}}`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	p := doc.Root.Children[0]
	if p.Align != "center" {
		t.Errorf("align = %q, want center", p.Align)
	}
	if p.TextFormat != 0 {
		t.Errorf("TextFormat = %d, want 0 for element node", p.TextFormat)
	}
}

func TestUnknownTypeDetection(t *testing.T) {
	if IsKnown("mystery-embed") {
		t.Error("mystery-embed should not be known")
	}
	if !IsKnown("paragraph") {
		t.Error("paragraph should be known")
	}
}
