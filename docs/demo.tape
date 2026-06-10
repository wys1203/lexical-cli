# VHS tape that records the lexical-cli TUI for the README.
# Regenerate with: go build -o /tmp/lexical-cli ./cmd/lexical-cli && vhs docs/demo.tape
Output docs/demo.gif

Set FontSize 16
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set Padding 10

# --- Main view: tree + markdown preview ---
Hide
Type "/tmp/lexical-cli tui testdata/rich.json"
Enter
Sleep 1.5s
Show
Sleep 1s
Screenshot docs/tui-main.png

# Navigate the tree
Down@200ms 6
Sleep 1s

# Switch focus to the preview pane
Tab
Sleep 500ms
Down@200ms 3
Sleep 1s

# --- Help overlay ---
Type "?"
Sleep 1s
Screenshot docs/tui-help.png
Type "x"
Sleep 500ms

# Quit first session
Type "q"
Sleep 500ms

# --- Warnings view (document with an unsupported node) ---
Hide
Type "/tmp/lexical-cli tui testdata/unsupported.json"
Enter
Sleep 1.5s
Show
Sleep 500ms
Type "w"
Sleep 1s
Screenshot docs/tui-warnings.png
Sleep 1s
Type "q"
Sleep 500ms
