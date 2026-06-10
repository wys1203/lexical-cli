package lexical

import (
	"fmt"
	"strings"
)

// TreeLine is a single flattened tree row for display in the TUI or --dump-tree.
type TreeLine struct {
	Depth int
	Label string
	Path  string
	Node  *Node
}

// Unsupported reports an unknown node type found in a document.
type Unsupported struct {
	Path string
	Type string
}

// NodeLabel returns a short human-readable label for a node.
func NodeLabel(n *Node) string {
	switch n.Type {
	case "text", "code-highlight", "hashtag":
		return fmt.Sprintf("%s %q", n.Type, truncate(n.Text, 40))
	case "heading":
		return fmt.Sprintf("heading (%s)", n.Tag)
	case "list":
		return fmt.Sprintf("list (%s)", n.ListType)
	case "link", "autolink":
		return fmt.Sprintf("%s → %s", n.Type, n.URL)
	case "code":
		if n.Language != "" {
			return fmt.Sprintf("code (%s)", n.Language)
		}
		return "code"
	case "image":
		return fmt.Sprintf("image %q", n.Src)
	case "linebreak", "tab":
		return n.Type
	default:
		return n.Type
	}
}

// Flatten produces display rows for the whole tree in document order.
func Flatten(root *Node) []TreeLine {
	var out []TreeLine
	var walk func(n *Node, depth int, path string)
	walk = func(n *Node, depth int, path string) {
		out = append(out, TreeLine{Depth: depth, Label: NodeLabel(n), Path: path, Node: n})
		for i, c := range n.Children {
			walk(c, depth+1, fmt.Sprintf("%s.children[%d]", path, i))
		}
	}
	walk(root, 0, "root")
	return out
}

// DumpTree renders a simplified ASCII tree for debugging.
func DumpTree(root *Node) string {
	var sb strings.Builder
	for _, line := range Flatten(root) {
		sb.WriteString(strings.Repeat("  ", line.Depth))
		sb.WriteString(line.Label)
		sb.WriteString("\n")
	}
	return sb.String()
}

// ListUnsupported returns every unknown node type in the document with its path.
func ListUnsupported(root *Node) []Unsupported {
	var out []Unsupported
	var walk func(n *Node, path string)
	walk = func(n *Node, path string) {
		if path != "root" && !IsKnown(n.Type) {
			out = append(out, Unsupported{Path: path, Type: n.Type})
		}
		for i, c := range n.Children {
			walk(c, fmt.Sprintf("%s.children[%d]", path, i))
		}
	}
	walk(root, "root")
	return out
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
