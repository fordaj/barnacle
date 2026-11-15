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

func TestMermaidEdgeLinkDirective(t *testing.T) {
	src := []byte("# Diagram\n\n```mermaid\nflowchart LR\n  Idle e1@--> Running\n  %% barnacle: link e1 href \"usage.md\" \"Run it\" _blank\n  Idle --> Done : send(cmd)\n  %% barnacle: link \"send(cmd)\" href \"api.go.md\"\n```\n")
	out, err := MarkdownToHTML(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`%% barnacle: link e1 href &quot;usage.html&quot; &quot;Run it&quot; _blank`,
		`%% barnacle: link &quot;send(cmd)&quot; href &quot;api.go.html&quot;`,
		`class="barnacle-edge-links"`,
		`{"edges":[{"key":"e1","by":"id","href":"usage.html","tooltip":"Run it","blank":true},{"key":"send(cmd)","by":"label","href":"api.go.html"}]}`,
	} {
		if !contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
	if !HasEdgeLinks(out) {
		t.Errorf("HasEdgeLinks reported false")
	}
}

func TestMermaidEdgeLinkFragmentAndExternalUntouched(t *testing.T) {
	src := []byte("flowchart LR\n  A e1@--> B\n  %% barnacle: link e1 href \"#local\"\n  %% barnacle: link e1 href \"https://example.com/docs/x.md\"\n  %% barnacle: link e1 href \"mailto:a@b.c\"\n")
	out := MermaidPageHTML("Diagram", src)
	for _, unwanted := range []string{`#local.html`, `x.html`, `mailto:a@b.html`} {
		if contains(out, unwanted) {
			t.Errorf("output unexpectedly rewrote target: %q\n%s", unwanted, out)
		}
	}
}

func TestEdgeLinkManifestEscapesScriptBreakout(t *testing.T) {
	d := edgeLinkDirective{Key: "e1", By: "id", Href: "x.html</script><script>alert(1)</script>"}
	script := edgeLinksScript([]edgeLinkDirective{d})
	if contains(script, `</script><script>alert(1)</script>`) {
		t.Errorf("manifest contains unescaped script breakout\n%s", script)
	}
	if !contains(script, `\u003c/script\u003e`) {
		t.Errorf("manifest did not escape < characters\n%s", script)
	}
}

func TestEdgeLinkDirectiveParse(t *testing.T) {
	d, ok := parseEdgeLinkDirective(`%% barnacle: link e1 href "usage.md"`)
	if !ok || d.Key != "e1" || d.By != "id" || d.Href != "usage.md" || d.Blank || d.Tooltip != "" {
		t.Errorf("parse bare: %+v ok=%v", d, ok)
	}
	d, ok = parseEdgeLinkDirective(`  %% barnacle: link "send(cmd)" href 'api.go.md' "The API" _blank`)
	if !ok || d.Key != "send(cmd)" || d.By != "label" || d.Href != "api.go.md" || d.Tooltip != "The API" || !d.Blank {
		t.Errorf("parse label: %+v ok=%v", d, ok)
	}
	d, ok = parseEdgeLinkDirective(`%% barnacle: link e1 href`)
	if ok {
		t.Errorf("directive without target must not parse: %+v", d)
	}
	if _, ok := parseEdgeLinkDirective(`%% just a comment`); ok {
		t.Errorf("ordinary comment must not parse as a directive")
	}
}

func TestEdgeLinkDirectiveRoundTrip(t *testing.T) {
	d, ok := parseEdgeLinkDirective(`%% barnacle: link "send(cmd)" href "usage.md" "tip" _blank`)
	if !ok {
		t.Fatal("parse failed")
	}
	d.Href = RewriteDocLink(d.Href)
	got := formatEdgeLinkDirective(d)
	want := `%% barnacle: link "send(cmd)" href "usage.html" "tip" _blank`
	if got != want {
		t.Errorf("formatEdgeLinkDirective() = %q, want %q", got, want)
	}
}

func TestMermaidEdgeLinkRewrittenIdempotent(t *testing.T) {
	raw := "flowchart LR\n  A e1@--> B\n  %% barnacle: link e1 href \"usage.md\"\n"
	once, d1 := rewriteMermaidLinks(raw)
	twice, d2 := rewriteMermaidLinks(once)
	if once != twice {
		t.Errorf("rewriting is not idempotent:\nfirst:\n%s\nsecond:\n%s", once, twice)
	}
	if len(d1) != 1 || len(d2) != 1 || d2[0].Href != "usage.html" {
		t.Errorf("directive targets wrong: %+v then %+v", d1, d2)
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
