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

func TestMarkdownToHTMLMermaidClickLinks(t *testing.T) {
	src := []byte("# Diagram\n\n```mermaid\nflowchart LR\n  A[main] --> B[parser]\n  click A href \"parser/parser.md\"\n  click B href '../main.md#start' _blank\n  click A \"docs/deep dive.md\"\n```\n")
	out, err := MarkdownToHTML(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`click A href &quot;parser/parser.html&quot;`,
		`click B href '../main.html#start'`,
		// Tooltip-style URLs containing spaces are never rewritten.
		`click A &quot;docs/deep dive.md&quot;`,
	} {
		if !contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestMarkdownToHTMLMermaidClickCallbacksUntouched(t *testing.T) {
	src := []byte("# Diagram\n\n```mermaid\nflowchart LR\n  A[main] --> B\n  click A callback \"see scan.md\"\n  click B call cb() \"see nav.md\"\n  click D href \"site.go.md\" \"The site generator, plain text.md\"\n```\n")
	out, err := MarkdownToHTML(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`click A callback &quot;see scan.md&quot;`,
		`click B call cb() &quot;see nav.md&quot;`,
		`click D href &quot;site.go.html&quot;`,
		`The site generator, plain text.md`,
	} {
		if !contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestMermaidPageHTMLRewritesClickLinks(t *testing.T) {
	src := []byte("flowchart LR\n  A --> B\n  click A href \"scan.go.md\"\n  click B \"render.go.md\"\n  click C callback \"template.go.md\"\n")
	out := MermaidPageHTML("Diagram", src)
	for _, want := range []string{
		`click A href &quot;scan.go.html&quot;`,
		`click B &quot;render.go.html&quot;`,
		`click C callback &quot;template.go.md&quot;`,
	} {
		if !contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestMarkdownToHTMLRawMermaidBlock(t *testing.T) {
	src := []byte("# Diagram\n\n<pre class=\"mermaid\">flowchart LR\n  A --> B\n  click A href \"scan.go.md\"\n</pre>\n")
	out, err := MarkdownToHTML(src)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out, `click A href &quot;scan.go.html&quot;`) {
		t.Errorf("raw mermaid block click target not rewritten\n%s", out)
	}
}

func TestMarkdownToHTMLMermaidClickExternalUntouched(t *testing.T) {
	src := []byte("# Diagram\n\n```mermaid\nflowchart LR\n  A[main] --> B\n  click A href \"https://example.com/docs/scan.md\"\n  click B href \"#anchor\"\n  click C href \"mailto:a@b.c\"\n```\n")
	out, err := MarkdownToHTML(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`https://example.com/docs/scan.md`,
		`href &quot;#anchor&quot;`,
		`href &quot;mailto:a@b.c&quot;`,
	} {
		if !contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
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
