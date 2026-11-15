package barnacle

import (
	"strings"
	"testing"
)

func TestRewriteDocLink(t *testing.T) {
	cases := []struct {
		dest string
		want string
	}{
		{"parser/parser.md", "parser/parser.html"},
		{"parser/parser.mmd", "parser/parser.html"},
		{"../parser/Parser.MD", "../parser/Parser.html"},
		{"guide.markdown", "guide.html"},
		{"doc.md#section", "doc.html#section"},
		{"doc.md?x=1#section", "doc.html?x=1#section"},
		{"https://example.com/doc.md", "https://example.com/doc.md"},
		{"http://example.com/x.mmd#y", "http://example.com/x.mmd#y"},
		{"mailto:a@b.c", "mailto:a@b.c"},
		{"#fragment", "#fragment"},
		{"//cdn.example.com/x.md", "//cdn.example.com/x.md"},
		{"/abs/doc.md", "/abs/doc.html"},
		{"image.png", "image.png"},
		{"doc.pdf", "doc.pdf"},
		{"", ""},
	}
	for _, tc := range cases {
		t.Run(tc.dest, func(t *testing.T) {
			if got := RewriteDocLink(tc.dest); got != tc.want {
				t.Errorf("RewriteDocLink(%q) = %q, want %q", tc.dest, got, tc.want)
			}
		})
	}
}

func TestMarkdownToHTMLRewritesDocLinks(t *testing.T) {
	src := []byte("# Title\n\nSee [the parser](../parser/parser.md#impl) and [external](https://example.com/x.md).\n")
	out, err := MarkdownToHTML(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`href="../parser/parser.html#impl"`,
		`href="https://example.com/x.md"`,
	} {
		if !contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
	if contains(out, `parser.md#impl`) {
		t.Errorf("output still contains .md link\n%s", out)
	}
}

func TestMarkdownToHTMLMermaidBlock(t *testing.T) {
	src := []byte("# Diagram\n\n```mermaid\nflowchart TD\n  A[main] --> B[start]\n```\n")
	out, err := MarkdownToHTML(src)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out, `<pre class="mermaid">`) {
		t.Errorf("mermaid block not converted\n%s", out)
	}
	if contains(out, `language-mermaid`) {
		t.Errorf("unconverted mermaid code block remains\n%s", out)
	}
	if !HasMermaid(out) {
		t.Errorf("HasMermaid reported false")
	}
}

func TestMarkdownGFMTable(t *testing.T) {
	src := []byte("# T\n\n| a | b |\n|---|---|\n| 1 | 2 |\n")
	out, err := MarkdownToHTML(src)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out, "<table>") {
		t.Errorf("GFM table not rendered\n%s", out)
	}
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}
