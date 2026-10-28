package knowledge

import (
	"strings"
	"unicode"
)

// maxFileName bounds a file name in bytes, before ".md"; file systems allow
// 255, and Obsidian shows the name as the note's title.
const maxFileName = 120

// unsafeInName are the characters a file name cannot hold on some system
// (\/:*?"<>|), or that Obsidian reads as link syntax (#^[]).
const unsafeInName = `\/:*?"<>|#^[]`

// FileName is a title as a markdown file's name, without ".md". Obsidian
// resolves [[Funding drag]] to the file named "Funding drag.md", so the name
// is the title itself, as far as file systems and links allow: unsafe
// characters become spaces, runs of spaces one, no leading dot (a hidden
// file) or trailing dot or space (Windows), Windows' reserved device names
// suffixed, "Untitled" when nothing is left.
func FileName(title string) string {
	var b strings.Builder
	space := false
	for _, r := range title {
		if strings.ContainsRune(unsafeInName, r) || unicode.IsControl(r) || unicode.IsSpace(r) {
			space = true
			continue
		}
		if space && b.Len() > 0 {
			b.WriteByte(' ')
		}
		space = false
		b.WriteRune(r)
	}
	s := strings.TrimLeft(b.String(), ".")
	s = strings.TrimRight(truncate(s, maxFileName), ". ")
	if s == "" {
		return "Untitled"
	}
	if reserved(s) {
		s += "_"
	}
	return s
}

// reserved is a name Windows keeps for a device, with or without extension.
func reserved(name string) bool {
	base, _, _ := strings.Cut(strings.ToUpper(name), ".")
	switch base {
	case "CON", "PRN", "AUX", "NUL":
		return true
	}
	return len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9'
}
