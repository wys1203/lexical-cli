package cli

import "testing"

func TestParseBasicFile(t *testing.T) {
	cfg, err := parseArgs([]string{"input.json"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Input != "input.json" || cfg.Command != "" {
		t.Errorf("got %+v", cfg)
	}
	if cfg.Format != "markdown" {
		t.Errorf("default format = %q", cfg.Format)
	}
}

func TestParseOutputAndFormat(t *testing.T) {
	cfg, err := parseArgs([]string{"-o", "out.md", "--format", "html", "in.json"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Output != "out.md" || cfg.Format != "html" || cfg.Input != "in.json" {
		t.Errorf("got %+v", cfg)
	}
}

func TestParseInlineValue(t *testing.T) {
	cfg, err := parseArgs([]string{"--format=text", "--output=x.txt", "in.json"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Format != "text" || cfg.Output != "x.txt" {
		t.Errorf("got %+v", cfg)
	}
}

func TestParseTUISubcommand(t *testing.T) {
	cfg, err := parseArgs([]string{"tui", "in.json"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Command != "tui" || cfg.Input != "in.json" {
		t.Errorf("got %+v", cfg)
	}
}

func TestParseStrictAndBools(t *testing.T) {
	cfg, err := parseArgs([]string{"--strict", "--dump-tree", "in.json"})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Strict || !cfg.DumpTree {
		t.Errorf("got %+v", cfg)
	}
}

func TestParseUnknownFlag(t *testing.T) {
	if _, err := parseArgs([]string{"--nope", "in.json"}); err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestParseTooManyPositionals(t *testing.T) {
	if _, err := parseArgs([]string{"a.json", "b.json"}); err == nil {
		t.Fatal("expected error for two input files")
	}
}

func TestParseMissingFlagValue(t *testing.T) {
	if _, err := parseArgs([]string{"--output"}); err == nil {
		t.Fatal("expected error for missing flag value")
	}
}
