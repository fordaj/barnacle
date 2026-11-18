package barnacle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/fordaj/barnacle/tools/lucide"
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
	out, err := MarkdownToHTML(RenderContext{}, src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`href="../parser/parser.html#impl"`,
		`href="https://example.com/x.md"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, `parser.md#impl`) {
		t.Errorf("output still contains .md link\n%s", out)
	}
}

func TestMarkdownToHTMLMermaidBlock(t *testing.T) {
	src := []byte("# Diagram\n\n```mermaid\nflowchart TD\n  A[main] --> B[start]\n```\n")
	out, err := MarkdownToHTML(RenderContext{}, src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `<pre class="mermaid">`) {
		t.Errorf("mermaid block not converted\n%s", out)
	}
	if strings.Contains(out, `language-mermaid`) {
		t.Errorf("unconverted mermaid code block remains\n%s", out)
	}
	if !HasMermaid(out) {
		t.Errorf("HasMermaid reported false")
	}
}

func TestMarkdownToHTMLMermaidClickLinks(t *testing.T) {
	src := []byte("# Diagram\n\n```mermaid\nflowchart LR\n  A[main] --> B[parser]\n  click A href \"parser/parser.md\"\n  click B href '../main.md#start' _blank\n  click A \"docs/deep dive.md\"\n```\n")
	out, err := MarkdownToHTML(RenderContext{}, src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`click A href &quot;parser/parser.html&quot;`,
		`click B href '../main.html#start'`,
		// Tooltip-style URLs containing spaces are never rewritten.
		`click A &quot;docs/deep dive.md&quot;`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestMarkdownToHTMLMermaidClickCallbacksUntouched(t *testing.T) {
	src := []byte("# Diagram\n\n```mermaid\nflowchart LR\n  A[main] --> B\n  click A callback \"see scan.md\"\n  click B call cb() \"see nav.md\"\n  click D href \"site.go.md\" \"The site generator, plain text.md\"\n```\n")
	out, err := MarkdownToHTML(RenderContext{}, src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`click A callback &quot;see scan.md&quot;`,
		`click B call cb() &quot;see nav.md&quot;`,
		`click D href &quot;site.go.html&quot;`,
		`The site generator, plain text.md`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestMermaidPageHTMLRewritesClickLinks(t *testing.T) {
	src := []byte("flowchart LR\n  A --> B\n  click A href \"scan.go.md\"\n  click B \"render.go.md\"\n  click C callback \"template.go.md\"\n")
	out, err := MermaidPageHTML(RenderContext{}, "Diagram", src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`click A href &quot;scan.go.html&quot;`,
		`click B &quot;render.go.html&quot;`,
		`click C callback &quot;template.go.md&quot;`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestMarkdownToHTMLRawMermaidBlock(t *testing.T) {
	src := []byte("# Diagram\n\n<pre class=\"mermaid\">flowchart LR\n  A --> B\n  click A href \"scan.go.md\"\n</pre>\n")
	out, err := MarkdownToHTML(RenderContext{}, src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `click A href &quot;scan.go.html&quot;`) {
		t.Errorf("raw mermaid block click target not rewritten\n%s", out)
	}
}

func TestMarkdownToHTMLMermaidClickExternalUntouched(t *testing.T) {
	src := []byte("# Diagram\n\n```mermaid\nflowchart LR\n  A[main] --> B\n  click A href \"https://example.com/docs/scan.md\"\n  click B href \"#anchor\"\n  click C href \"mailto:a@b.c\"\n```\n")
	out, err := MarkdownToHTML(RenderContext{}, src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`https://example.com/docs/scan.md`,
		`href &quot;#anchor&quot;`,
		`href &quot;mailto:a@b.c&quot;`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestMarkdownGFMTable(t *testing.T) {
	src := []byte("# T\n\n| a | b |\n|---|---|\n| 1 | 2 |\n")
	out, err := MarkdownToHTML(RenderContext{}, src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<table>") {
		t.Errorf("GFM table not rendered\n%s", out)
	}
}

func TestInlineLocalSVGsRewritesRelativeRefs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "icons/wave.svg", `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M12 3v12"/></svg>`)
	src := []byte("# Title\n\n<img src=\"icons/wave.svg\"/>\n")
	out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `<svg`) {
		t.Errorf("svg was not inlined\n%s", out)
	}
	if !strings.Contains(out, `<path d="M12 3v12" style="fill:none;stroke:currentColor"/>`) {
		t.Errorf("svg body missing or malformed\n%s", out)
	}
	if strings.Contains(out, `src="icons/wave.svg"`) {
		t.Errorf("relative svg reference not rewritten\n%s", out)
	}
}

func TestInlineLocalSVGsRewritesMermaidLabelRefs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "wave.svg", `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M12 3v12"/></svg>`)
	src := []byte("# Diagram\n\n```mermaid\nflowchart LR\n  A[\"<img src='wave.svg' width='15'/>main\"] --> B\n```\n")
	out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `&lt;svg`) || !strings.Contains(out, `aria-hidden=&quot;true&quot;`) {
		t.Errorf("svg in mermaid label was not inlined\n%s", out)
	}
	start := strings.Index(out, "&lt;svg")
	end := strings.Index(out, "&lt;/svg&gt;")
	if start < 0 || end < start {
		t.Fatalf("svg tags not found in balanced order\n%s", out)
	}
	if bytes := out[start:end]; strings.ContainsAny(bytes, "\n\r") {
		t.Errorf("inline SVG contains a raw newline; mermaid would render <br/>s inside the label\n%q", bytes)
	}
}

func TestInlineLocalSVGsMissingFileFails(t *testing.T) {
	root := t.TempDir()
	src := []byte("# Title\n\n<img src=\"icons/missing.svg\"/>\n")
	_, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, src)
	if err == nil {
		t.Fatal("expected error for missing svg")
	}
}

func TestInlineLocalSVGsEscapesRootFails(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "secret.svg", `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M12 3v12"/></svg>`)
	src := []byte("# Title\n\n<img src=\"secret.svg\"/>\n")
	// The file is inside SourceDir but outside Root — escaping is an error.
	_, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: filepath.Join(root, "inner")}, src)
	if err == nil {
		t.Fatal("expected error for path escaping the repository root")
	}
}

func TestInlineLocalSVGsExternalUntouched(t *testing.T) {
	root := t.TempDir()
	for _, dest := range []string{
		`https://example.com/icon.svg`,
		`data:image/svg+xml;utf8,<svg/>`,
		`#icon`,
		`//cdn.example.com/icon.svg`,
	} {
		src := []byte("# Title\n\n<img src=\"" + dest + "\"/>\n")
		out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, src)
		if err != nil {
			t.Fatalf("dest %q: unexpected error: %v", dest, err)
		}
		if !strings.Contains(out, dest) {
			t.Errorf("dest %q was rewritten\n%s", dest, out)
		}
	}
}

func TestInlineLocalSVGsZeroContextSkips(t *testing.T) {
	src := []byte("# Title\n\n<img src=\"icons/wave.svg\"/>\n")
	out, err := MarkdownToHTML(RenderContext{}, src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `src="icons/wave.svg"`) {
		t.Errorf("zero context must skip inlining, got\n%s", out)
	}
}

func TestInlineLocalSVGsRespectsShapeFillOverride(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "custom.svg", `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="contextColor"><path d="M0 0h24" fill="#ff0000"/><path d="M0 12h24"/></svg>`)
	src := []byte("# Title\n\n<img src=\"custom.svg\"/>\n")
	out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, src)
	if err != nil {
		t.Fatal(err)
	}
	// shape with explicit fill= wins over the projection
	if !strings.Contains(out, `fill="#ff0000"`) {
		t.Errorf("shape-declared fill was not preserved\n%s", out)
	}
	// stroke was projected inline because shape had none
	if !strings.Contains(out, `stroke="contextColor"`) {
		t.Errorf("file-declared stroke was not projected inline\n%s", out)
	}
}

func TestInlineLocalSVGsDefaultsLikeLucide(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "mini.svg", `<svg viewBox="0 0 24 24"><path d="M0 0h24"/></svg>`)
	src := []byte("# Title\n\n<img src=\"mini.svg\"/>\n")
	out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `fill:none`) {
		t.Errorf("default fill:none not applied\n%s", out)
	}
	if !strings.Contains(out, `stroke:currentColor`) {
		t.Errorf("default stroke:currentColor not applied\n%s", out)
	}
}

func TestInlineLocalSVGSetAltBecomesAriaLabel(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "icons/wave.svg", `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M12 3v12"/></svg>`)
	src := []byte("# Title\n\n<img src=\"icons/wave.svg\" alt=\"a wave\"/>\n")
	out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `aria-label="a wave"`) {
		t.Errorf("author alt became aria-label\n%s", out)
	}
	if strings.Contains(out, `aria-hidden="true"`) {
		t.Errorf("aria-label alt must suppress aria-hidden\n%s", out)
	}
}

func TestInlineBarnacleIconsRewritesEmbeddedRefs(t *testing.T) {
	root := t.TempDir()
	src := []byte("# Title\n\n<img src=\"barnacle:folder.svg\" width=\"20\" alt=\"a folder\"/>\n")
	out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `<svg`) || !strings.Contains(out, `aria-label="a folder"`) {
		t.Errorf("embedded icon was not inlined with the img attributes\n%s", out)
	}
	if strings.Contains(out, `src="barnacle:folder.svg"`) {
		t.Errorf("barnacle: reference not rewritten\n%s", out)
	}
}

func TestInlineBarnacleIconsInMermaidLabelRefs(t *testing.T) {
	root := t.TempDir()
	src := []byte("# Diagram\n\n```mermaid\nflowchart LR\n  A[\"<img src='barnacle:file-code.svg' width='15'/>main\"] --> B\n```\n")
	out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `&lt;svg`) {
		t.Errorf("embedded icon in mermaid label was not inlined\n%s", out)
	}
}

func TestInlineBarnacleIconsUnknownFails(t *testing.T) {
	root := t.TempDir()
	for _, dest := range []string{
		`barnacle:not-a-real-icon.svg`,
		`barnacle:Folder.svg`, // names are lowercase
		`barnacle:folder`,     // missing .svg suffix
		`barnacle:../escape.svg`,
		`barnacle:a/b.svg`,
	} {
		src := []byte("# Title\n\n<img src=\"" + dest + "\"/>\n")
		if _, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, src); err == nil {
			t.Errorf("dest %q: expected error", dest)
		}
	}
}

func TestInlineBarnacleIconsZeroContextSkips(t *testing.T) {
	src := []byte("# Title\n\n<img src=\"barnacle:folder.svg\"/>\n")
	out, err := MarkdownToHTML(RenderContext{}, src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `src="barnacle:folder.svg"`) {
		t.Errorf("zero context must skip inlining, got\n%s", out)
	}
}

func TestInlineBarnacleIconsMatchRelativeFileInlining(t *testing.T) {
	root := t.TempDir()
	// Well-known icons embedded in the binary; the same bytes written to
	// disk must produce byte-identical output.
	for _, name := range []string{"folder", "file-code", "tag", "hexagon", "check", "terminal"} {
		data, err := lucide.Icon(name)
		if err != nil {
			t.Fatalf("embedded icon %q unreadable: %v", name, err)
		}
		writeFile(t, root, "icons/"+name+".svg", string(data))

		render := func(dest string) string {
			src := []byte("# Title\n\n<img src=\"" + dest + "\" width=\"20\" alt=\"icon\"/>\n")
			out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, src)
			if err != nil {
				t.Fatalf("dest %q: %v", dest, err)
			}
			return out
		}
		viaFile := render("icons/" + name + ".svg")
		viaEmbed := render("barnacle:" + name + ".svg")
		if viaFile != viaEmbed {
			t.Errorf("icon %q: barnacle: and relative-file inlining differ\nfile:  %s\nembed: %s", name, viaFile, viaEmbed)
		}
		if !strings.Contains(viaEmbed, "<svg") {
			t.Errorf("icon %q: barnacle: reference was not inlined\n%s", name, viaEmbed)
		}
	}
}

func TestInlineLocalSVGWideViewBoxBecomesFluid(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "wide.svg", `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1400 200"><rect width="1400" height="200"/></svg>`)
	out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, []byte("<img src=\"wide.svg\">\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `width="100%"`) {
		t.Errorf("wide viewBox svg should be fluid width\n%s", out)
	}
	if !strings.Contains(out, `height="auto"`) {
		t.Errorf("wide viewBox svg should have auto height\n%s", out)
	}
}

func TestInlineLocalSVGFileDimensionsKept(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "fixed.svg", `<svg xmlns="http://www.w3.org/2000/svg" width="640" height="360" viewBox="0 0 640 360"><rect width="640" height="360"/></svg>`)
	out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, []byte("<img src=\"fixed.svg\">\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `width="640"`) || !strings.Contains(out, `height="360"`) {
		t.Errorf("file-declared dimensions should win over defaults\n%s", out)
	}
}

func TestInlineLocalSVGSquareViewBoxKeepsHouseDefault(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "icon.svg", `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48"><rect width="48" height="48"/></svg>`)
	out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, []byte("<img src=\"icon.svg\">\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `width="20px"`) || !strings.Contains(out, `height="20px"`) {
		t.Errorf("square viewBox should keep the 20px house default\n%s", out)
	}
}

func TestInlineLocalSVGImgDimensionsWin(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "wide.svg", `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1400 200"><rect width="1400" height="200"/></svg>`)
	out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, []byte("<img src=\"wide.svg\" width=\"320\">\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `width="320"`) || !strings.Contains(out, `height="320"`) {
		t.Errorf("img width alone should square the image\n%s", out)
	}
}
