package markdown

import (
	"fmt"
	"strings"

	"github.com/wys1203/lexical-cli/internal/lexical"
)

// renderBlocks renders the block-level children of a container node (root,
// quote, table cell, ...) joined by blank lines.
func (c *converter) renderBlocks(parent *lexical.Node, parentPath string) string {
	var blocks []string
	for i, child := range parent.Children {
		b := c.renderBlock(child, childPath(parentPath, i))
		if strings.TrimSpace(b) != "" {
			blocks = append(blocks, b)
		}
	}
	return strings.Join(blocks, "\n\n")
}

// renderBlock renders a single block-level node to Markdown.
func (c *converter) renderBlock(n *lexical.Node, path string) string {
	switch n.Type {
	case "paragraph":
		return c.renderInlineChildren(n, path)

	case "heading":
		level := headingLevel(n.Tag)
		return strings.Repeat("#", level) + " " + c.renderInlineChildren(n, path)

	case "quote":
		inner := c.renderInlineChildren(n, path)
		return prefixLines(inner, "> ")

	case "list":
		return c.renderList(n, path, 0)

	case "code":
		return c.renderCode(n)

	case "horizontalrule":
		return "---"

	case "image":
		return fmt.Sprintf("![%s](%s)", escapeText(n.AltText), n.Src)

	case "table":
		return c.renderTable(n, path)

	case "":
		// Defensive: a node with no type. Treat as a container.
		return c.renderBlocks(n, path)

	default:
		if lexical.IsKnown(n.Type) {
			// Known but not block-level here (e.g. stray text): render inline.
			return c.renderInline(n, path)
		}
		c.warn(path, fmt.Sprintf("unsupported node type %q", n.Type))
		// Best-effort: render any child content so text is preserved.
		if len(n.Children) > 0 {
			return c.renderBlocks(n, path)
		}
		if n.Text != "" {
			return escapeText(n.Text)
		}
		return ""
	}
}

func headingLevel(tag string) int {
	switch tag {
	case "h1":
		return 1
	case "h2":
		return 2
	case "h3":
		return 3
	case "h4":
		return 4
	case "h5":
		return 5
	case "h6":
		return 6
	default:
		return 1
	}
}

// renderCode renders a Lexical code block as a fenced Markdown block.
func (c *converter) renderCode(n *lexical.Node) string {
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
	fence := "```"
	body := sb.String()
	// Use a longer fence if the body contains a triple backtick.
	for strings.Contains(body, fence) {
		fence += "`"
	}
	return fence + n.Language + "\n" + body + "\n" + fence
}

// renderList renders a list (and nested lists) with indentation by depth.
func (c *converter) renderList(n *lexical.Node, path string, depth int) string {
	indent := strings.Repeat("  ", depth)
	var lines []string
	counter := n.Start
	if counter == 0 {
		counter = 1
	}

	for i, item := range n.Children {
		itemPath := childPath(path, i)
		if item.Type != "listitem" {
			// Unexpected child; render best-effort.
			content := c.renderInline(item, itemPath)
			if content != "" {
				lines = append(lines, indent+content)
			}
			continue
		}

		marker := c.listMarker(n, item, counter)
		counter++

		inline, nested := c.splitListItem(item, itemPath, depth)
		lines = append(lines, indent+marker+inline)
		lines = append(lines, nested...)
	}
	return strings.Join(lines, "\n")
}

func (c *converter) listMarker(list, item *lexical.Node, counter int) string {
	switch list.ListType {
	case "number":
		val := counter
		if item.Value != 0 {
			val = item.Value
		}
		return fmt.Sprintf("%d. ", val)
	case "check":
		if item.Checked != nil && *item.Checked {
			return "- [x] "
		}
		return "- [ ] "
	default:
		return "- "
	}
}

// splitListItem returns the inline content of a list item plus any rendered
// nested lists (indented one level deeper than the current item).
func (c *converter) splitListItem(item *lexical.Node, itemPath string, depth int) (string, []string) {
	var inlineParts []string
	var nested []string
	for i, child := range item.Children {
		cp := childPath(itemPath, i)
		if child.Type == "list" {
			nested = append(nested, c.renderList(child, cp, depth+1))
			continue
		}
		inlineParts = append(inlineParts, c.renderInline(child, cp))
	}
	return strings.Join(inlineParts, ""), nested
}

// renderTable renders a Lexical table as a GitHub-flavored Markdown table.
func (c *converter) renderTable(n *lexical.Node, path string) string {
	var rows [][]string
	for i, row := range n.Children {
		rp := childPath(path, i)
		if row.Type != "tablerow" {
			continue
		}
		var cells []string
		for j, cell := range row.Children {
			cp := childPath(rp, j)
			cells = append(cells, strings.TrimSpace(c.renderInlineChildren(cell, cp)))
		}
		rows = append(rows, cells)
	}
	if len(rows) == 0 {
		return ""
	}

	width := 0
	for _, r := range rows {
		if len(r) > width {
			width = len(r)
		}
	}

	var sb strings.Builder
	writeRow := func(cells []string) {
		sb.WriteString("|")
		for k := 0; k < width; k++ {
			val := ""
			if k < len(cells) {
				val = cells[k]
			}
			sb.WriteString(" " + val + " |")
		}
		sb.WriteString("\n")
	}

	writeRow(rows[0])
	sb.WriteString("|")
	for k := 0; k < width; k++ {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")
	for _, r := range rows[1:] {
		writeRow(r)
	}
	return strings.TrimRight(sb.String(), "\n")
}

// prefixLines prefixes every line of s with prefix.
func prefixLines(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
