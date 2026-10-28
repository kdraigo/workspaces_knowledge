package knowledge

import (
	"errors"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		name  string
		in    string
		front map[string]any
		body  string
	}{
		{"no frontmatter", "# Title\n\ntext\n", nil, "# Title\n\ntext\n"},
		{"frontmatter", "---\nkind: note\ntags: [a, b]\n---\n# Title\n", map[string]any{"kind": "note", "tags": []any{"a", "b"}}, "# Title\n"},
		{"closed with dots", "---\nkind: brief\n...\nbody", map[string]any{"kind": "brief"}, "body"},
		{"empty frontmatter", "---\n---\nbody\n", nil, "body\n"},
		{"windows line endings", "---\r\nkind: note\r\n---\r\nbody\r\n", map[string]any{"kind": "note"}, "body\n"},
		{"a later rule is body", "text\n---\nmore\n", nil, "text\n---\nmore\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := Parse([]byte(tc.in))
			if err != nil {
				t.Fatal(err)
			}
			if d.Body != tc.body {
				t.Errorf("body %q, want %q", d.Body, tc.body)
			}
			if len(d.Front) != len(tc.front) {
				t.Fatalf("front %v, want %v", d.Front, tc.front)
			}
			for k, v := range tc.front {
				if got := d.Front[k]; !equal(got, v) {
					t.Errorf("front[%s] = %v, want %v", k, got, v)
				}
			}
		})
	}
	if _, err := Parse([]byte("---\nkind: note\nbody without a close\n")); !errors.Is(err, ErrUnclosedFront) {
		t.Errorf("unclosed: %v", err)
	}
	if _, err := Parse([]byte("---\nkind: [unclosed\n---\n")); err == nil {
		t.Error("bad YAML must be an error")
	}
}

func equal(a, b any) bool {
	as, aok := a.([]any)
	bs, bok := b.([]any)
	if aok && bok {
		if len(as) != len(bs) {
			return false
		}
		for i := range as {
			if as[i] != bs[i] {
				return false
			}
		}
		return true
	}
	return a == b
}

// A rendered file reads back the same: keys sorted, body untouched.
func TestRenderRoundTrip(t *testing.T) {
	d := Doc{Front: map[string]any{"title": "Funding drag", "kind": "note", "tags": []string{"futures", "funding"}}, Body: "Longs pay. See [[Carry]].\n"}
	b, err := d.Render()
	if err != nil {
		t.Fatal(err)
	}
	want := "---\nkind: note\ntags:\n  - futures\n  - funding\ntitle: Funding drag\n---\nLongs pay. See [[Carry]].\n"
	if string(b) != want {
		t.Errorf("rendered\n%s\nwant\n%s", b, want)
	}
	back, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	if back.Body != d.Body || back.String("title") != "Funding drag" || strings.Join(back.Strings("tags"), ",") != "futures,funding" {
		t.Errorf("round trip: %+v", back)
	}
	if b, _ := (Doc{Body: "only body"}).Render(); string(b) != "only body" {
		t.Errorf("no frontmatter rendered %q", b)
	}
}

func TestLinks(t *testing.T) {
	body := strings.Join([]string{
		"See [[Funding drag]] and [[Carry#Basis|the basis]].",
		"An embed: ![[Equity curve]].",
		"Inline `[[not a link]]` stays text, and so does [[]] and [[a [b]]].",
		"```",
		"[[inside a fence]]",
		"```",
		"~~~~go",
		"[[inside a tilde fence]]",
		"~~~~",
		"After the fences: [[Lessons]]",
	}, "\n")
	got := Links(body)
	want := []Link{
		{Target: "Funding drag", Line: 1},
		{Target: "Carry", Heading: "Basis", Alias: "the basis", Line: 1},
		{Target: "Equity curve", Embed: true, Line: 2},
		{Target: "Lessons", Line: 10},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("link %d: %+v, want %+v", i, got[i], want[i])
		}
	}
	if got[0].Key() != "funding-drag" {
		t.Errorf("key %q", got[0].Key())
	}
}

func TestSlug(t *testing.T) {
	for in, want := range map[string]string{
		"Funding drag on longs (BTC)": "funding-drag-on-longs-btc",
		"  RSI 14 / 30↑  ":            "rsi-14-30",
		"Ставка финансирования":       "ставка-финансирования",
		"???":                     "untitled",
		strings.Repeat("ab ", 40): strings.TrimRight(strings.Repeat("ab-", 27), "-"),
	} {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
	// A long title in a two-byte script is cut on a character, not inside one.
	long := Slug(strings.Repeat("я", 60))
	if len(long) > maxSlug || !strings.HasPrefix(strings.Repeat("я", 60), long) {
		t.Errorf("cut inside a character: %q (%d bytes)", long, len(long))
	}
}

func TestValidKind(t *testing.T) {
	for _, k := range Kinds {
		if !ValidKind(k) {
			t.Errorf("%s", k)
		}
	}
	if ValidKind("memo") {
		t.Error("memo is not a kind")
	}
}

func TestFileName(t *testing.T) {
	for in, want := range map[string]string{
		"Funding drag on longs (BTC)": "Funding drag on longs (BTC)",
		"RSI 14 / 30: exits?":         "RSI 14 30 exits",
		"  [[x]] #tag ^block  ":       "x tag block",
		"Ставка финансирования":       "Ставка финансирования",
		".hidden. ":        "hidden",
		"line\nbreak\ttab": "line break tab",
		"???":              "Untitled",
		"con":              "con_",
		"LPT3.txt":         "LPT3.txt_",
		"COM0":             "COM0",
	} {
		if got := FileName(in); got != want {
			t.Errorf("FileName(%q) = %q, want %q", in, got, want)
		}
	}
	// A long title in a two-byte script is cut on a character.
	long := FileName(strings.Repeat("я", 100))
	if len(long) > maxFileName || !strings.HasPrefix(strings.Repeat("я", 100), long) {
		t.Errorf("cut inside a character: %q (%d bytes)", long, len(long))
	}
}
