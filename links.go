package knowledge

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Link is a [[wikilink]] in a body: [[Target]], [[Target#Heading]],
// [[Target|shown as]], or an embed, ![[Target]].
type Link struct {
	// Target is the note it points to, as written.
	Target  string
	Heading string
	Alias   string
	Embed   bool
	// Line is where it is, from 1.
	Line int
}

// Key is what a link resolves by: its target's slug, so [[Funding Drag]]
// and [[funding drag]] reach the same note.
func (l Link) Key() string { return Slug(l.Target) }

// Links finds the wikilinks in a markdown body, in order. Links inside code
// (fenced blocks and inline `spans`) are text, not links.
func Links(body string) []Link {
	var out []Link
	fence := ""
	for n, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if f := fenceOf(trimmed); f != "" {
			switch {
			case fence == "":
				fence = f
			case strings.HasPrefix(trimmed, fence):
				fence = ""
			}
			continue
		}
		if fence != "" {
			continue
		}
		out = append(out, lineLinks(line, n+1)...)
	}
	return out
}

// fenceOf is the fence a line opens or closes (``` or ~~~, three or more).
func fenceOf(line string) string {
	for _, c := range []string{"```", "~~~"} {
		if strings.HasPrefix(line, c) {
			n := len(line) - len(strings.TrimLeft(line, c[:1]))
			return strings.Repeat(c[:1], n)
		}
	}
	return ""
}

func lineLinks(line string, n int) []Link {
	var out []Link
	inCode := false
	for i := 0; i < len(line); i++ {
		switch {
		case line[i] == '`':
			inCode = !inCode
		case inCode:
		case strings.HasPrefix(line[i:], "[["):
			end := strings.Index(line[i+2:], "]]")
			if end < 0 {
				return out
			}
			inner := line[i+2 : i+2+end]
			if l, ok := parseLink(inner); ok {
				l.Embed = i > 0 && line[i-1] == '!'
				l.Line = n
				out = append(out, l)
			}
			i += 2 + end + 1
		}
	}
	return out
}

func parseLink(inner string) (Link, bool) {
	var l Link
	target, alias, hasAlias := strings.Cut(inner, "|")
	if hasAlias {
		l.Alias = strings.TrimSpace(alias)
	}
	target, heading, _ := strings.Cut(target, "#")
	l.Target, l.Heading = strings.TrimSpace(target), strings.TrimSpace(heading)
	if l.Target == "" || strings.ContainsAny(l.Target, "[]") {
		return Link{}, false
	}
	return l, true
}

// maxSlug bounds a slug.
const maxSlug = 80

// Slug is a title as the platform's id and a link's key (a file's name is
// FileName): lower case, letters and
// digits of any script kept, everything else one dash, "untitled" when
// nothing is left. "Funding drag on longs (BTC)" is
// "funding-drag-on-longs-btc".
func Slug(title string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(title) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if dash && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			dash = false
			continue
		}
		dash = true
	}
	s := b.String()
	if len(s) > maxSlug {
		s = strings.TrimRight(truncate(s, maxSlug), "-")
	}
	if s == "" {
		return "untitled"
	}
	return s
}

// truncate cuts s to at most n bytes without splitting a character.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}
