package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(name string) string {
	return filepath.Join("..", "..", "testdata", name)
}

// runWith executes Run with buffered I/O and returns exit code, stdout, stderr.
func runWith(args []string, stdin string, ttyOut bool) (int, string, string) {
	var out, errBuf bytes.Buffer
	code := Run(Env{
		Args:        args,
		Stdin:       strings.NewReader(stdin),
		Stdout:      &out,
		Stderr:      &errBuf,
		StdinIsTTY:  false,
		StdoutIsTTY: ttyOut,
	})
	return code, out.String(), errBuf.String()
}

func TestRunFileToStdout(t *testing.T) {
	code, out, _ := runWith([]string{fixture("paragraph.json")}, "", false)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if out != "Hello world.\n" {
		t.Errorf("stdout = %q", out)
	}
}

func TestRunStdin(t *testing.T) {
	data, _ := os.ReadFile(fixture("paragraph.json"))
	code, out, _ := runWith(nil, string(data), false)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if out != "Hello world.\n" {
		t.Errorf("stdout = %q", out)
	}
}

func TestRunOutputFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.md")
	code, _, _ := runWith([]string{"-o", target, fixture("paragraph.json")}, "", false)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "Hello world.\n" {
		t.Errorf("file = %q", got)
	}
}

func TestRunInvalidJSON(t *testing.T) {
	code, _, errOut := runWith(nil, "{ broken", false)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut, "error:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunMissingFile(t *testing.T) {
	code, _, _ := runWith([]string{"/no/such/file.json"}, "", false)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
}

func TestRunMissingRoot(t *testing.T) {
	code, _, _ := runWith(nil, `{"notroot":{}}`, false)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
}

func TestRunStrictUnsupported(t *testing.T) {
	code, _, errOut := runWith([]string{"--strict", fixture("unsupported.json")}, "", false)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut, "children[1]") {
		t.Errorf("stderr should mention path, got %q", errOut)
	}
}

func TestRunNonStrictWarns(t *testing.T) {
	code, out, errOut := runWith([]string{fixture("unsupported.json")}, "", false)
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(errOut, "warning:") {
		t.Errorf("expected warning on stderr, got %q", errOut)
	}
	if !strings.Contains(out, "Before.") {
		t.Errorf("expected content on stdout, got %q", out)
	}
}

func TestRunVersion(t *testing.T) {
	code, out, _ := runWith([]string{"--version"}, "", false)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(out, "lexical-cli") {
		t.Errorf("version output = %q", out)
	}
}

func TestRunHelp(t *testing.T) {
	code, out, _ := runWith([]string{"--help"}, "", false)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(out, "USAGE") {
		t.Errorf("help output = %q", out)
	}
}

func TestRunUnknownFlag(t *testing.T) {
	code, _, _ := runWith([]string{"--bogus"}, "", false)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}

func TestRunInvalidFormat(t *testing.T) {
	code, _, _ := runWith([]string{"--format", "yaml", fixture("paragraph.json")}, "", false)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}

func TestRunHTMLFormat(t *testing.T) {
	code, out, _ := runWith([]string{"--format", "html", fixture("paragraph.json")}, "", false)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(out, "<p>Hello world.</p>") {
		t.Errorf("html = %q", out)
	}
}

func TestRunDumpTree(t *testing.T) {
	code, out, _ := runWith([]string{"--dump-tree", fixture("formats.json")}, "", false)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(out, "root") || !strings.Contains(out, "heading (h1)") {
		t.Errorf("dump-tree = %q", out)
	}
}

func TestRunListUnsupported(t *testing.T) {
	code, out, _ := runWith([]string{"--list-unsupported", fixture("unsupported.json")}, "", false)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(out, "mystery-embed") {
		t.Errorf("list-unsupported = %q", out)
	}
}

func TestRunListUnsupportedNone(t *testing.T) {
	code, out, _ := runWith([]string{"--list-unsupported", fixture("paragraph.json")}, "", false)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(out, "No unsupported nodes") {
		t.Errorf("got %q", out)
	}
}

func TestRunTUIRequiresFile(t *testing.T) {
	code, _, errOut := runWith([]string{"tui"}, "", true)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "requires an input file") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunTUINonTTY(t *testing.T) {
	code, _, errOut := runWith([]string{"tui", fixture("paragraph.json")}, "", false)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut, "interactive terminal") {
		t.Errorf("stderr = %q", errOut)
	}
}
