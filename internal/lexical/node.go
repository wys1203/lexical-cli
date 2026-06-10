// Package lexical models Meta Lexical EditorState JSON and provides parsing
// and validation independent of any output format or terminal UI.
package lexical

import (
	"encoding/json"
	"strconv"
)

// Text format bitmask values used by Lexical TextNode.format.
// See https://github.com/facebook/lexical packages/lexical/src/nodes/LexicalTextNode.ts
const (
	FormatBold          = 1 << 0 // 1
	FormatItalic        = 1 << 1 // 2
	FormatStrikethrough = 1 << 2 // 4
	FormatUnderline     = 1 << 3 // 8
	FormatCode          = 1 << 4 // 16
	FormatSubscript     = 1 << 5 // 32
	FormatSuperscript   = 1 << 6 // 64
	FormatHighlight     = 1 << 7 // 128
)

// Node is a single node in a Lexical tree. Lexical nodes are heterogeneous, so
// Node carries the union of fields used by the supported node types. Fields not
// relevant to a given node type are left at their zero value.
type Node struct {
	Type    string
	Version int

	Children []*Node

	// TextNode fields.
	Text       string
	TextFormat int // bitmask, see Format* constants
	Style      string
	Mode       string
	Detail     int

	// ElementNode alignment ("", "left", "center", "right", "justify").
	Align string

	// HeadingNode.
	Tag string

	// ListNode / ListItemNode.
	ListType string // "bullet", "number", "check"
	Start    int
	Value    int
	Checked  *bool

	// LinkNode.
	URL    string
	Rel    string
	Target string
	Title  string

	// CodeNode.
	Language string

	// ImageNode and similar.
	Src     string
	AltText string

	// Raw holds the original decoded object so unknown node types can still be
	// inspected (for warnings and best-effort rendering).
	Raw map[string]json.RawMessage
}

// rawNode mirrors the JSON shape. Format is decoded separately because it is a
// number on text nodes but a string (alignment) on element nodes.
type rawNode struct {
	Type     string          `json:"type"`
	Version  int             `json:"version"`
	Children []*Node         `json:"children"`
	Text     string          `json:"text"`
	Format   json.RawMessage `json:"format"`
	Style    string          `json:"style"`
	Mode     string          `json:"mode"`
	Detail   int             `json:"detail"`
	Tag      string          `json:"tag"`
	ListType string          `json:"listType"`
	Start    int             `json:"start"`
	Value    int             `json:"value"`
	Checked  *bool           `json:"checked"`
	URL      string          `json:"url"`
	Rel      string          `json:"rel"`
	Target   string          `json:"target"`
	Title    string          `json:"title"`
	Language string          `json:"language"`
	Src      string          `json:"src"`
	AltText  string          `json:"altText"`
}

// UnmarshalJSON decodes a Lexical node, resolving the polymorphic format field
// and preserving the raw object for unknown node types.
func (n *Node) UnmarshalJSON(data []byte) error {
	var r rawNode
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}

	n.Type = r.Type
	n.Version = r.Version
	n.Children = r.Children
	n.Text = r.Text
	n.Style = r.Style
	n.Mode = r.Mode
	n.Detail = r.Detail
	n.Tag = r.Tag
	n.ListType = r.ListType
	n.Start = r.Start
	n.Value = r.Value
	n.Checked = r.Checked
	n.URL = r.URL
	n.Rel = r.Rel
	n.Target = r.Target
	n.Title = r.Title
	n.Language = r.Language
	n.Src = r.Src
	n.AltText = r.AltText

	// format is a number (text bitmask) or a string (element alignment).
	if len(r.Format) > 0 && string(r.Format) != "null" {
		if num, err := strconv.Atoi(string(r.Format)); err == nil {
			n.TextFormat = num
		} else {
			var s string
			if err := json.Unmarshal(r.Format, &s); err == nil {
				n.Align = s
			}
		}
	}

	// Keep the raw object so unknown nodes remain inspectable.
	_ = json.Unmarshal(data, &n.Raw)

	return nil
}

// HasFormat reports whether the given text format bit is set.
func (n *Node) HasFormat(bit int) bool {
	return n.TextFormat&bit != 0
}
