package markdown

import (
	"fmt"
	"html"
	"strings"

	"github.com/wys1203/lexical-cli/internal/lexical"
)

// renderHTMLChildren renders the block-level children of a container as HTML.
func (c *converter) renderHTMLChildren(parent *lexical.Node, parentPath string) string {
	var blocks []string
	for i, child := range parent.Children {
		b := c.renderHTMLBlock(child, childPath(parentPath, i))
		if strings.TrimSpace(b) != "" {
			blocks = append(blocks, b)
		}
	}
	return strings.Join(blocks, "\n")
}

func (c *converter) renderHTMLBlock(n *lexical.Node, path string) string {
	switch n.Type {
	case "paragraph":
		return "<p>" + c.htmlInlineChildren(n, path) + "</p>"
	case "heading":
		tag := n.Tag
		if tag == "" {
			tag = "h1"
		}
		return fmt.Sprintf("<%s>%s</%s>", tag, c.htmlInlineChildren(n, path), tag)
	case "quote":
		return "<blockquote>" + c.htmlInlineChildren(n, path) + "</blockquote>"
	case "list":
		return c.renderHTMLList(n, path)
	case "code":
		lang := ""
		if n.Language != "" {
			lang = fmt.Sprintf(` class="language-%s"`, html.EscapeString(n.Language))
		}
		return "<pre><code" + lang + ">" + html.EscapeString(c.renderCodeText(n)) + "</code></pre>"
	case "horizontalrule":
		return "<hr>"
	case "image":
		return fmt.Sprintf("<img src=%q alt=%q>", n.Src, n.AltText)
	case "table":
		return c.renderHTMLTable(n, path)
	case "":
		return c.renderHTMLChildren(n, path)
	default:
		if !lexical.IsKnown(n.Type) {
			c.warn(path, fmt.Sprintf("unsupported node type %q", n.Type))
		}
		if len(n.Children) > 0 {
			return c.renderHTMLChildren(n, path)
		}
		return html.EscapeString(n.Text)
	}
}

func (c *converter) renderHTMLList(n *lexical.Node, path string) string {
	tag := "ul"
	if n.ListType == "number" {
		tag = "ol"
	}
	var sb strings.Builder
	sb.WriteString("<" + tag + ">")
	for i, item := range n.Children {
		itemPath := childPath(path, i)
		if item.Type != "listitem" {
			continue
		}
		sb.WriteString("<li>")
		for j, child := range item.Children {
			cp := childPath(itemPath, j)
			if child.Type == "list" {
				sb.WriteString(c.renderHTMLList(child, cp))
			} else {
				sb.WriteString(c.htmlInline(child, cp))
			}
		}
		sb.WriteString("</li>")
	}
	sb.WriteString("</" + tag + ">")
	return sb.String()
}

func (c *converter) renderHTMLTable(n *lexical.Node, path string) string {
	var sb strings.Builder
	sb.WriteString("<table>")
	for i, row := range n.Children {
		rp := childPath(path, i)
		if row.Type != "tablerow" {
			continue
		}
		sb.WriteString("<tr>")
		for j, cell := range row.Children {
			cp := childPath(rp, j)
			sb.WriteString("<td>" + c.htmlInlineChildren(cell, cp) + "</td>")
		}
		sb.WriteString("</tr>")
	}
	sb.WriteString("</table>")
	return sb.String()
}

func (c *converter) htmlInlineChildren(n *lexical.Node, path string) string {
	var sb strings.Builder
	for i, child := range n.Children {
		sb.WriteString(c.htmlInline(child, childPath(path, i)))
	}
	return sb.String()
}

func (c *converter) htmlInline(n *lexical.Node, path string) string {
	switch n.Type {
	case "text", "code-highlight":
		return htmlApplyFormat(n)
	case "hashtag":
		return html.EscapeString(n.Text)
	case "linebreak":
		return "<br>"
	case "tab":
		return "\t"
	case "link", "autolink":
		label := c.htmlInlineChildren(n, path)
		if label == "" {
			label = html.EscapeString(n.URL)
		}
		return fmt.Sprintf("<a href=%q>%s</a>", n.URL, label)
	case "image":
		return fmt.Sprintf("<img src=%q alt=%q>", n.Src, n.AltText)
	default:
		if len(n.Children) > 0 {
			if !lexical.IsKnown(n.Type) {
				c.warn(path, fmt.Sprintf("unsupported node type %q", n.Type))
			}
			return c.htmlInlineChildren(n, path)
		}
		if !lexical.IsKnown(n.Type) {
			c.warn(path, fmt.Sprintf("unsupported node type %q", n.Type))
		}
		return html.EscapeString(n.Text)
	}
}

func htmlApplyFormat(n *lexical.Node) string {
	if n.Text == "" {
		return ""
	}
	out := html.EscapeString(n.Text)
	if n.HasFormat(lexical.FormatCode) {
		out = "<code>" + out + "</code>"
	}
	if n.HasFormat(lexical.FormatStrikethrough) {
		out = "<del>" + out + "</del>"
	}
	if n.HasFormat(lexical.FormatUnderline) {
		out = "<u>" + out + "</u>"
	}
	if n.HasFormat(lexical.FormatItalic) {
		out = "<em>" + out + "</em>"
	}
	if n.HasFormat(lexical.FormatBold) {
		out = "<strong>" + out + "</strong>"
	}
	return out
}
