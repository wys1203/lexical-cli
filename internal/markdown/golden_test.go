package markdown

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// update regenerates golden files: go test ./internal/markdown -update
var update = flag.Bool("update", false, "update golden snapshot files")

func TestGoldenSnapshots(t *testing.T) {
	cases := []struct {
		fixture string
		format  Format
		golden  string
	}{
		{"rich.json", FormatMarkdown, "rich.md"},
		{"rich.json", FormatHTML, "rich.html"},
		{"lists.json", FormatMarkdown, "lists.md"},
		{"formats.json", FormatMarkdown, "formats.md"},
		{"code.json", FormatMarkdown, "code.md"},
	}

	for _, tc := range cases {
		t.Run(tc.golden, func(t *testing.T) {
			res, err := Convert(load(t, tc.fixture), Options{Format: tc.format})
			if err != nil {
				t.Fatal(err)
			}
			goldenPath := filepath.Join("..", "..", "testdata", "golden", tc.golden)
			if *update {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(goldenPath, []byte(res.Output), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden (run with -update to create): %v", err)
			}
			if res.Output != string(want) {
				t.Errorf("output mismatch for %s\n got:\n%q\nwant:\n%q", tc.golden, res.Output, want)
			}
		})
	}
}
