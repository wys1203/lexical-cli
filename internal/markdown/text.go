package markdown

import (
	"fmt"
	"strings"

	"github.com/wys1203/lexical-cli/internal/lexical"
)

// renderBlocksText renders block children as plain text, dropping inline
// emphasis markers but keeping readable block structure.
func (c *converter) renderBlocksText(parent *lexical.Node, parentPath string) string {
	var blocks []string
	for i, child := range parent.Children {
		b := c.renderBlockText(child, childPath(parentPath, i))
		if strings.TrimSpace(b) != "" {
			blocks = append(blocks, b)
		}
	}
	return strings.Join(blocks, "\n\n")
}

func (c *converter) renderBlockText(n *lexical.Node, path string) string {
	switch n.Type {
	case "paragraph", "heading", "quote":
		return c.inlineText(n, path)
	case "list":
		return c.renderListText(n, path, 0)
	case "code":
		return c.renderCodeText(n)
	case "horizontalrule":
		return strings.Repeat("-", 40)
	case "image":
		if n.AltText != "" {
			return fmt.Sprintf("%s (%s)", n.AltText, n.Src)
		}
		return n.Src
	case "table":
		return c.renderTableText(n, path)
	case "":
		return c.renderBlocksText(n, path)
	default:
		if !lexical.IsKnown(n.Type) {
			c.warn(path, fmt.Sprintf("unsupported node type %q", n.Type))
		}
		if len(n.Children) > 0 {
			return c.renderBlocksText(n, path)
		}
		return n.Text
	}
}

func (c *converter) inlineText(n *lexical.Node, path string) string {
	var sb strings.Builder
	for i, child := range n.Children {
		sb.WriteString(c.inlineTextNode(child, childPath(path, i)))
	}
	return sb.String()
}

// inlineTextNode renders a single inline node to plain text.
func (c *converter) inlineTextNode(n *lexical.Node, path string) string {
	switch n.Type {
	case "linebreak":
		return "\n"
	case "tab":
		return "\t"
	case "link", "autolink":
		label := c.inlineText(n, path)
		if label == "" {
			label = n.URL
		}
		return label
	case "text", "code-highlight", "hashtag":
		return n.Text
	default:
		if len(n.Children) > 0 {
			if !lexical.IsKnown(n.Type) {
				c.warn(path, fmt.Sprintf("unsupported node type %q", n.Type))
			}
			return c.inlineText(n, path)
		}
		if !lexical.IsKnown(n.Type) {
			c.warn(path, fmt.Sprintf("unsupported node type %q", n.Type))
		}
		return n.Text
	}
}

func (c *converter) renderListText(n *lexical.Node, path string, depth int) string {
	indent := strings.Repeat("  ", depth)
	var lines []string
	counter := n.Start
	if counter == 0 {
		counter = 1
	}
	for i, item := range n.Children {
		itemPath := childPath(path, i)
		if item.Type != "listitem" {
			continue
		}
		marker := c.listMarker(n, item, counter)
		counter++
		var inlineParts []string
		var nested []string
		for j, child := range item.Children {
			cp := childPath(itemPath, j)
			if child.Type == "list" {
				nested = append(nested, c.renderListText(child, cp, depth+1))
				continue
			}
			inlineParts = append(inlineParts, c.inlineTextNode(child, cp))
		}
		lines = append(lines, indent+marker+strings.Join(inlineParts, ""))
		lines = append(lines, nested...)
	}
	return strings.Join(lines, "\n")
}

func (c *converter) renderCodeText(n *lexical.Node) string {
	var sb strings.Builder
	for _, child := range n.Children {
		switch child.Type {
		case "linebreak":
			sb.WriteString("\n")
		case "tab":
			sb.WriteString("\t")
		default:
			sb.WriteString(child.Text)
		}
	}
	return sb.String()
}

func (c *converter) renderTableText(n *lexical.Node, path string) string {
	var lines []string
	for i, row := range n.Children {
		rp := childPath(path, i)
		if row.Type != "tablerow" {
			continue
		}
		var cells []string
		for j, cell := range row.Children {
			cp := childPath(rp, j)
			cells = append(cells, strings.TrimSpace(c.inlineText(cell, cp)))
		}
		lines = append(lines, strings.Join(cells, "\t"))
	}
	return strings.Join(lines, "\n")
}
