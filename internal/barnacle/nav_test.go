package barnacle

import (
	"strings"
	"testing"
)

func TestRenderInlineCode(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "Configuration", "Configuration"},
		{"escapes html", "a <b> & c", "a &lt;b&gt; &amp; c"},
		{"single span", "The page template (`template.go`)", "The page template (<code>template.go</code>)"},
		{"multiple spans", "`Main` and `Generate` work", "<code>Main</code> and <code>Generate</code> work"},
		{"span escaped", "a `<b>` c", "a <code>&lt;b&gt;</code> c"},
		{"unmatched backtick stays literal", "use `x to do things", "use `x to do things"},
		{"empty span stays literal", "a `` b", "a `` b"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := renderInlineCode(tc.in); got != tc.want {
				t.Errorf("renderInlineCode(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestRenderNavInlineCode(t *testing.T) {
	docs := []Document{
		{RelPath: "main.go.md", OutputRel: "main.go.html", Title: "The command (`barnacle`) runs `generate` <here>"},
	}
	html := renderNav(buildTree(docs), "other.html")
	for _, want := range []string{
		"<code>barnacle</code>",
		"<code>generate</code>",
		"&lt;here&gt;", // outer text still escaped
	} {
		if !strings.Contains(html, want) {
			t.Errorf("nav output missing %q\n%s", want, html)
		}
	}
}

func TestRenderIndexInlineCode(t *testing.T) {
	docs := []Document{
		{RelPath: "parser/parser.go.md", OutputRel: "parser/parser.go.html", Title: "The `Parser`"},
	}
	html := renderIndex("Site `title`", buildTree(docs))
	for _, want := range []string{
		"<h1>Site <code>title</code></h1>",
		"<code>Parser</code>",
		`href="parser/parser.go.html"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("index output missing %q\n%s", want, html)
		}
	}
}
