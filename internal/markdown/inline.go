package markdown

import (
	"fmt"
	"strings"

	"github.com/wys1203/lexical-cli/internal/lexical"
)

// renderInlineChildren renders the inline children of an element node.
func (c *converter) renderInlineChildren(n *lexical.Node, path string) string {
	var sb strings.Builder
	for i, child := range n.Children {
		sb.WriteString(c.renderInline(child, childPath(path, i)))
	}
	return sb.String()
}

// renderInline renders a single inline node to Markdown.
func (c *converter) renderInline(n *lexical.Node, path string) string {
	switch n.Type {
	case "text", "code-highlight":
		return applyFormat(n)

	case "linebreak":
		// Markdown hard line break.
		return "  \n"

	case "tab":
		return "\t"

	case "hashtag":
		return escapeText(n.Text)

	case "link", "autolink":
		label := c.renderInlineChildren(n, path)
		if label == "" {
			label = n.URL
		}
		if n.Title != "" {
			return fmt.Sprintf("[%s](%s \"%s\")", label, n.URL, n.Title)
		}
		return fmt.Sprintf("[%s](%s)", label, n.URL)

	case "image":
		return fmt.Sprintf("![%s](%s)", escapeText(n.AltText), n.Src)

	default:
		if len(n.Children) > 0 {
			// Container appearing in an inline context: render its children.
			if !lexical.IsKnown(n.Type) {
				c.warn(path, fmt.Sprintf("unsupported node type %q", n.Type))
			}
			return c.renderInlineChildren(n, path)
		}
		if !lexical.IsKnown(n.Type) {
			c.warn(path, fmt.Sprintf("unsupported node type %q", n.Type))
		}
		return escapeText(n.Text)
	}
}

// applyFormat wraps text with Markdown emphasis markers per the format bitmask.
// Inline code is rendered with backticks and, because Markdown does not render
// emphasis inside code spans, other emphasis is applied around the code span.
func applyFormat(n *lexical.Node) string {
	if n.Text == "" {
		return ""
	}

	var out string
	if n.HasFormat(lexical.FormatCode) {
		out = wrapCode(n.Text)
	} else {
		out = escapeText(n.Text)
	}

	// Apply from innermost to outermost. Underline has no Markdown equivalent,
	// so it uses an HTML tag; highlight uses the extended-Markdown ==…== marker.
	if n.HasFormat(lexical.FormatHighlight) {
		out = "==" + out + "=="
	}
	if n.HasFormat(lexical.FormatStrikethrough) {
		out = "~~" + out + "~~"
	}
	if n.HasFormat(lexical.FormatUnderline) {
		out = "<u>" + out + "</u>"
	}
	if n.HasFormat(lexical.FormatItalic) {
		out = "*" + out + "*"
	}
	if n.HasFormat(lexical.FormatBold) {
		out = "**" + out + "**"
	}
	return out
}

// wrapCode wraps s in a backtick code span, widening the fence if s contains
// backticks.
func wrapCode(s string) string {
	fence := "`"
	for strings.Contains(s, fence) {
		fence += "`"
	}
	pad := ""
	if strings.HasPrefix(s, "`") || strings.HasSuffix(s, "`") {
		pad = " "
	}
	return fence + pad + s + pad + fence
}

// markdownSpecials are characters escaped in regular Markdown text to avoid
// accidental formatting.
var mdReplacer = strings.NewReplacer(
	`\`, `\\`,
	"`", "\\`",
	"*", `\*`,
	"_", `\_`,
	"[", `\[`,
	"]", `\]`,
)

// escapeText escapes Markdown-significant characters in plain text.
func escapeText(s string) string {
	return mdReplacer.Replace(s)
}
