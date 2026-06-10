package lexical

// Document is a parsed Lexical EditorState with a top-level root node.
type Document struct {
	Root *Node `json:"root"`
}

// KnownTypes is the set of node types this tool understands. Used to decide
// whether a node triggers an unsupported-node warning.
var KnownTypes = map[string]bool{
	"root":           true,
	"paragraph":      true,
	"text":           true,
	"heading":        true,
	"quote":          true,
	"list":           true,
	"listitem":       true,
	"linebreak":      true,
	"link":           true,
	"autolink":       true,
	"code":           true,
	"code-highlight": true,
	"tab":            true,
	"table":          true,
	"tablerow":       true,
	"tablecell":      true,
	"horizontalrule": true,
	"image":          true,
	"hashtag":        true,
}

// IsKnown reports whether the node type is supported by the converter.
func IsKnown(nodeType string) bool {
	return KnownTypes[nodeType]
}
