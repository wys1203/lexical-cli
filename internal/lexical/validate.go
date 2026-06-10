package lexical

import "fmt"

// Validate checks that a decoded document has the expected Lexical root shape.
func Validate(doc *Document) error {
	if doc == nil || doc.Root == nil {
		return fmt.Errorf("invalid Lexical document: missing top-level \"root\" object")
	}
	if doc.Root.Type != "" && doc.Root.Type != "root" {
		return fmt.Errorf("invalid Lexical document: root.type is %q, expected \"root\"", doc.Root.Type)
	}
	return nil
}
