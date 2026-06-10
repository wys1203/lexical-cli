package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wys1203/lexical-cli/internal/lexical"
)

func load(t *testing.T, name string) *lexical.Document {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	doc, err := lexical.Parse(data)
	if err != nil {
		t.Fatalf("parse fixture %s: %v", name, err)
	}
	return doc
}

func convertMD(t *testing.T, name string) Result {
	t.Helper()
	res, err := Convert(load(t, name), Options{Format: FormatMarkdown})
	if err != nil {
		t.Fatalf("convert %s: %v", name, err)
	}
	return res
}

func TestParagraph(t *testing.T) {
	res := convertMD(t, "paragraph.json")
	if res.Output != "Hello world.\n" {
		t.Errorf("got %q", res.Output)
	}
}

func TestHeadingAndFormats(t *testing.T) {
	res := convertMD(t, "formats.json")
	want := "# Title\n\n**bold** *italic* ~~strike~~ <u>underline</u> `code` ***bolditalic***\n"
	if res.Output != want {
		t.Errorf("got:\n%q\nwant:\n%q", res.Output, want)
	}
}

func TestHeadingLevels(t *testing.T) {
	doc, _ := lexical.Parse([]byte(`{"root":{"type":"root","children":[
		{"type":"heading","tag":"h3","children":[{"type":"text","format":0,"text":"Sub"}]}]}}`))
	res, err := Convert(doc, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Output != "### Sub\n" {
		t.Errorf("got %q", res.Output)
	}
}

func TestUnorderedAndOrderedLists(t *testing.T) {
	res := convertMD(t, "lists.json")
	want := strings.Join([]string{
		"- Item one",
		"- Item two",
		"  1. Nested one",
		"  2. Nested two",
		"",
		"1. First",
		"2. Second",
		"",
	}, "\n")
	if res.Output != want {
		t.Errorf("got:\n%q\nwant:\n%q", res.Output, want)
	}
}

func TestLinks(t *testing.T) {
	res := convertMD(t, "links.json")
	want := "See [Lexical](https://lexical.dev) for details.\n"
	if res.Output != want {
		t.Errorf("got %q want %q", res.Output, want)
	}
}

func TestCodeBlock(t *testing.T) {
	res := convertMD(t, "code.json")
	want := "```go\nfunc main() {\n    println(\"hi\")\n}\n```\n"
	if res.Output != want {
		t.Errorf("got:\n%q\nwant:\n%q", res.Output, want)
	}
}

func TestQuoteAndLineBreaks(t *testing.T) {
	res := convertMD(t, "quote.json")
	want := "> To be, or not to be.  \n> That is the question.\n"
	if res.Output != want {
		t.Errorf("got:\n%q\nwant:\n%q", res.Output, want)
	}
}

func TestUnsupportedNonStrict(t *testing.T) {
	res, err := Convert(load(t, "unsupported.json"), Options{Format: FormatMarkdown})
	if err != nil {
		t.Fatalf("non-strict should not error: %v", err)
	}
	if len(res.Warnings) != 1 {
		t.Fatalf("want 1 warning, got %d: %v", len(res.Warnings), res.Warnings)
	}
	if !strings.Contains(res.Warnings[0].Path, "children[1]") {
		t.Errorf("warning path = %q, want it to mention children[1]", res.Warnings[0].Path)
	}
	// Best-effort text preservation.
	if !strings.Contains(res.Output, "salvageable text") {
		t.Errorf("expected salvaged text in output, got %q", res.Output)
	}
	if !strings.Contains(res.Output, "Before.") || !strings.Contains(res.Output, "After.") {
		t.Errorf("expected surrounding paragraphs preserved, got %q", res.Output)
	}
}

func TestUnsupportedStrict(t *testing.T) {
	_, err := Convert(load(t, "unsupported.json"), Options{Format: FormatMarkdown, Strict: true})
	if err == nil {
		t.Fatal("strict mode should error on unsupported node")
	}
	if !strings.Contains(err.Error(), "children[1]") {
		t.Errorf("strict error should include path, got %v", err)
	}
}

func TestDeterministicOutput(t *testing.T) {
	first := convertMD(t, "lists.json").Output
	for i := 0; i < 5; i++ {
		if got := convertMD(t, "lists.json").Output; got != first {
			t.Fatalf("non-deterministic output on run %d", i)
		}
	}
}

func TestTextFormat(t *testing.T) {
	res, err := Convert(load(t, "formats.json"), Options{Format: FormatText})
	if err != nil {
		t.Fatal(err)
	}
	want := "Title\n\nbold italic strike underline code bolditalic\n"
	if res.Output != want {
		t.Errorf("got:\n%q\nwant:\n%q", res.Output, want)
	}
}

func TestHTMLFormat(t *testing.T) {
	res, err := Convert(load(t, "formats.json"), Options{Format: FormatHTML})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Output, "<h1>Title</h1>") {
		t.Errorf("missing heading, got %q", res.Output)
	}
	if !strings.Contains(res.Output, "<strong>bold</strong>") {
		t.Errorf("missing bold, got %q", res.Output)
	}
	if !strings.Contains(res.Output, "<code>code</code>") {
		t.Errorf("missing code, got %q", res.Output)
	}
}

func TestInvalidFormat(t *testing.T) {
	_, err := Convert(load(t, "paragraph.json"), Options{Format: "yaml"})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestEscaping(t *testing.T) {
	doc, _ := lexical.Parse([]byte(`{"root":{"type":"root","children":[
		{"type":"paragraph","children":[{"type":"text","format":0,"text":"a*b_c[d]"}]}]}}`))
	res, _ := Convert(doc, Options{})
	if res.Output != `a\*b\_c\[d\]`+"\n" {
		t.Errorf("got %q", res.Output)
	}
}
