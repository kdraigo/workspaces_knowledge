package knowledge

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Doc is one file: its frontmatter and its markdown body.
type Doc struct {
	// Front is the YAML frontmatter as a map; nil when the file has none.
	Front map[string]any
	// Body is everything after the frontmatter.
	Body string
}

// ErrUnclosedFront is a file that opens frontmatter and never closes it.
var ErrUnclosedFront = errors.New("frontmatter opened with --- and never closed")

// Parse reads a file. Frontmatter is a "---" line at the very start, YAML,
// and a closing "---" (or "...") line; without the opening line the whole
// file is the body. Windows line endings are read as Unix ones.
func Parse(data []byte) (Doc, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return Doc{Body: text}, nil
	}
	lines := strings.SplitAfter(text[len("---\n"):], "\n")
	var front strings.Builder
	for i, line := range lines {
		if l := strings.TrimSuffix(line, "\n"); l == "---" || l == "..." {
			doc := Doc{Body: strings.Join(lines[i+1:], "")}
			if strings.TrimSpace(front.String()) == "" {
				return doc, nil
			}
			if err := yaml.Unmarshal([]byte(front.String()), &doc.Front); err != nil {
				return Doc{}, fmt.Errorf("frontmatter: %w", err)
			}
			return doc, nil
		}
		front.WriteString(line)
	}
	return Doc{}, ErrUnclosedFront
}

// Render writes the file: frontmatter with its keys sorted, then the body.
// A Doc without frontmatter renders as its body alone.
func (d Doc) Render() ([]byte, error) {
	if len(d.Front) == 0 {
		return []byte(d.Body), nil
	}
	var b bytes.Buffer
	b.WriteString("---\n")
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(d.Front); err != nil {
		return nil, fmt.Errorf("frontmatter: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	b.WriteString("---\n")
	b.WriteString(d.Body)
	return b.Bytes(), nil
}

// String reads a frontmatter key as a string; "" when it is missing or not
// a string.
func (d Doc) String(key string) string {
	s, _ := d.Front[key].(string)
	return s
}

// Strings reads a frontmatter key as a list of strings ("tags: [a, b]", or a
// single "tags: a"); nil when it is missing.
func (d Doc) Strings(key string) []string {
	switch v := d.Front[key].(type) {
	case string:
		return []string{v}
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
