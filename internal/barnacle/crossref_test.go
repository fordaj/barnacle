package barnacle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func render(t *testing.T, src string) string {
	t.Helper()
	out, err := MarkdownToHTML(RenderContext{}, []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestFigureCaptionWithLabel(t *testing.T) {
	out := render(t, "![System architecture](arch.png){#fig:arch}\n")
	for _, want := range []string{
		`<figure class="figure" id="fig:arch">`,
		`<img src="arch.png" alt="System architecture">`,
		`<figcaption><span class="caption-label">Figure 1:</span> System architecture</figcaption>`,
		`</figure>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestFigureCaptionEscapesAlt(t *testing.T) {
	out := render(t, `![a & "b"](x.png){#fig:escape}`+"\n")
	if !strings.Contains(out, `<figcaption><span class="caption-label">Figure 1:</span> a &amp; &#34;b&#34;</figcaption>`) {
		t.Errorf("caption text not escaped or wrong\n%s", out)
	}
}

func TestFigureMidParagraphIsNotAFigure(t *testing.T) {
	out := render(t, "text ![cap](x.png){#fig:inline} more\n")
	if strings.Contains(out, `<figure`) {
		t.Errorf("mid-paragraph image became a figure\n%s", out)
	}
	if !strings.Contains(out, `{#fig:inline}`) {
		t.Errorf("mid-paragraph attribute unexpectedly consumed\n%s", out)
	}
}

func TestTableCaptionWithLabel(t *testing.T) {
	src := "| a | b |\n|---|---|\n| 1 | 2 |\n\n: Timings on a laptop {#tbl:timings}\n"
	out := render(t, src)
	for _, want := range []string{
		`<table id="tbl:timings"><caption><span class="caption-label">Table 1:</span> Timings on a laptop</caption>`,
		`<div class="table-scroll">`,
		`</table></div>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, `: Timings on a laptop`) {
		t.Errorf("caption paragraph left in output\n%s", out)
	}
}

func TestTableCaptionWithoutLabel(t *testing.T) {
	src := "| a |\n|---|\n| 1 |\n\n: Just a caption\n"
	out := render(t, src)
	if !strings.Contains(out, `Table 1:</span> Just a caption`) {
		t.Errorf("table caption missing\n%s", out)
	}
	if strings.Contains(out, `id="tbl:`) {
		t.Errorf("unlabelled table gained an id\n%s", out)
	}
}

func TestTableInlineCodeInCaption(t *testing.T) {
	src := "| a |\n|---|\n| 1 |\n\n: Output of `render()` for @tbl:missing ignored {#tbl:code}\n"
	out := render(t, src)
	if !strings.Contains(out, `<code>render()</code>`) {
		t.Errorf("inline code in table caption lost\n%s", out)
	}
	if !strings.Contains(out, `<table id="tbl:code">`) {
		t.Errorf("table label missing\n%s", out)
	}
	if strings.Contains(out, `<a class="xref" href="#tbl:missing"`) {
		t.Errorf("reference to an undefined label should be left alone\n%s", out)
	}
}

func TestPlainTableIsWrappedNotCaptioned(t *testing.T) {
	src := "| a | b |\n|---|---|\n| 1 | 2 |\n"
	out := render(t, src)
	if !strings.Contains(out, `<div class="table-scroll"><table>`) {
		t.Errorf("plain table not wrapped in scroll container\n%s", out)
	}
	if strings.Contains(out, `<caption>`) {
		t.Errorf("plain table unexpectedly gained a caption\n%s", out)
	}
}

func TestFiguresAndTablesNumberIndependently(t *testing.T) {
	src := `![one](1.png){#fig:one}

| a |
|---|
| 1 |

: first table {#tbl:t}

![two](2.png){#fig:two}

| b |
|---|
| 2 |

: second table {#tbl:u}

![three](3.png){#fig:three}
`
	out := render(t, src)
	for _, want := range []string{
		`id="fig:one"><img`, `id="fig:two"><img`, `id="fig:three"><img`,
		`Figure 1:</span> one`, `Figure 2:</span> two`, `Figure 3:</span> three`,
		`Table 1:</span> first table`, `Table 2:</span> second table`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestCrossReferences(t *testing.T) {
	src := "![arch](a.png){#fig:arch}\n\n| a |\n|---|\n| 1 |\n\n: cap {#tbl:cap}\n\nSee @fig:arch, [@tbl:cap], @tbl:unknown and @fig:missing.\n"
	out := render(t, src)
	for _, want := range []string{
		`<a class="xref" href="#fig:arch">Figure 1</a>`,
		`<a class="xref" href="#tbl:cap">Table 1</a>`,
		`@tbl:unknown`, `@fig:missing`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestCrossReferenceSentencePeriod(t *testing.T) {
	src := "![arch](a.png){#fig:arch}\n\nSee @fig:arch. Next sentence.\n"
	out := render(t, src)
	if !strings.Contains(out, `<a class="xref" href="#fig:arch">Figure 1</a>. Next sentence.`) {
		t.Errorf("sentence-final period was swallowed\n%s", out)
	}
}

func TestCrossReferenceNotInCode(t *testing.T) {
	src := "![arch](a.png){#fig:arch}\n\nInline `@fig:arch` code.\n\n```\n@fig:arch\n```\n"
	out := render(t, src)
	if !strings.Contains(out, `<code>@fig:arch</code>`) {
		t.Errorf("inline code reference was rewritten\n%s", out)
	}
	if !strings.Contains(out, "<pre") {
		t.Fatalf("code block missing\n%s", out)
	}
	if strings.Count(out, `<a class="xref"`) != 0 {
		t.Errorf("code reference was rewritten\n%s", out)
	}
}

func TestCrossReferenceEmailUntouched(t *testing.T) {
	src := "![arch](a.png){#fig:arch}\n\nMail user@fig:arch for the diagram.\n"
	out := render(t, src)
	if strings.Contains(out, `<a class="xref"`) {
		t.Errorf("email address rewritten as cross-reference\n%s", out)
	}
	if !strings.Contains(out, "user@fig:arch") {
		t.Errorf("email address mangled\n%s", out)
	}
}

func TestFigureWithLocalSVGImage(t *testing.T) {
	// Regression: the caption pass must run before local SVGs are inlined
	// as <svg> elements; otherwise the figure pattern (an <img>-shaped
	// construct) never matches and the {#fig:} attribute leaks through.
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "diagram.svg"), []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect width="100" height="100"/></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := MarkdownToHTML(RenderContext{SourceDir: root, Root: root}, []byte("![Architecture](diagram.svg){#fig:arch}\n\nSee @fig:arch.\n"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<figure class="figure" id="fig:arch">`,
		`<svg`,
		`Figure 1:</span> Architecture`,
		`<a class="xref" href="#fig:arch">Figure 1</a>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, `{#fig:arch}`) {
		t.Errorf("figure attribute leaked into output\n%s", out)
	}
}

func TestDuplicateFigureLabelFails(t *testing.T) {
	src := "![one](1.png){#fig:dup}\n\n![two](2.png){#fig:dup}\n"
	_, err := MarkdownToHTML(RenderContext{}, []byte(src))
	if err == nil {
		t.Fatal("expected duplicate label error, got nil")
	}
	if !strings.Contains(err.Error(), "fig:dup") {
		t.Errorf("error should name the duplicate label: %v", err)
	}
}

func TestDuplicateTableLabelFails(t *testing.T) {
	src := "| a |\n|---|\n| 1 |\n\n: one {#tbl:dup}\n\n| b |\n|---|\n| 2 |\n\n: two {#tbl:dup}\n"
	_, err := MarkdownToHTML(RenderContext{}, []byte(src))
	if err == nil {
		t.Fatal("expected duplicate label error, got nil")
	}
	if !strings.Contains(err.Error(), "tbl:dup") {
		t.Errorf("error should name the duplicate label: %v", err)
	}
}

func TestFigureAndTableCrossReferenceTypes(t *testing.T) {
	src := "![fig](f.png){#fig:same}\n\n| a |\n|---|\n| 1 |\n\n: tbl {#tbl:same}\n\nRef @fig:same then @tbl:same.\n"
	out := render(t, src)
	if !strings.Contains(out, `<a class="xref" href="#fig:same">Figure 1</a>`) {
		t.Errorf("figure reference missing\n%s", out)
	}
	if !strings.Contains(out, `<a class="xref" href="#tbl:same">Table 1</a>`) {
		t.Errorf("table reference missing\n%s", out)
	}
}
