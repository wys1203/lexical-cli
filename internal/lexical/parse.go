package lexical

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Parse decodes a Lexical EditorState document from raw JSON bytes. It returns
// a descriptive error for malformed JSON or a missing/invalid root.
func Parse(data []byte) (*Document, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("empty input: expected Lexical EditorState JSON")
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	var doc Document
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if err := Validate(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}
