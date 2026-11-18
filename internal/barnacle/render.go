package barnacle

import (
	"bytes"
	"fmt"
	"html"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldhtml "github.com/yuin/goldmark/renderer/html"

	"github.com/fordaj/barnacle/tools/lucide"
)

var markdown = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(goldhtml.WithUnsafe()),
)

// RenderContext carries the per-document context needed by rewrites that
// read the filesystem, such as local SVG inlining. Its zero value disables
// those rewrites (context is empty), which keeps all label rewriting inert.
type RenderContext struct {
	// SourceDir is the absolute path of the directory holding the source
	// document. Local resource references are resolved against it.
	SourceDir string
	// Root is the absolute repository root. Local SVG references must
	// resolve beneath it; anything escaping is an error.
	Root string
}

// MarkdownToHTML renders a Markdown document to an HTML fragment. Links to
// other documentation files (.md/.mmd) are rewritten to point at the
// generated .html pages — in ordinary links and inside Mermaid `click`
// directives — and Mermaid code fences are turned into <pre class="mermaid">
// blocks so they render in the browser. Local SVG references in <img> tags
// are inlined as inline <svg> elements — both relative filesystem paths and
// barnacle:<name>.svg references to the embedded Lucide set (unless the
// context is the zero value). Captioned figures and tables are numbered
// and labelled pandoc-crossref style, tables are wrapped in a scroll
// container, and @fig:label / @tbl:label cross-references are replaced
// with links to the numbered captions (see crossref.go). The caption pass
// runs before SVG inlining, because figures are an <img>-shaped pattern.
func MarkdownToHTML(ctx RenderContext, source []byte) (string, error) {
	var buf bytes.Buffer
	if err := markdown.Convert(source, &buf); err != nil {
		return "", fmt.Errorf("rendering markdown: %w", err)
	}
	content := buf.String()
	content, err := rewriteMermaidBlocks(content, ctx)
	if err != nil {
		return "", err
	}
	content = rewriteDocLinks(content)
	content, err = applyCrossrefs(content)
	if err != nil {
		return "", err
	}
	content, err = inlineLocalSVGs(content, ctx)
	if err != nil {
		return "", err
	}
	return content, nil
}

var hrefRe = regexp.MustCompile(`href="([^"]*)"`)

// rewriteDocLinks rewrites the href attributes of links that point at
// Markdown/Mermaid documentation files so they target the generated HTML
// pages instead. External links, anchors and fragments are left untouched.
func rewriteDocLinks(rendered string) string {
	return hrefRe.ReplaceAllStringFunc(rendered, func(match string) string {
		dest := match[len(`href="`) : len(match)-1]
		rewritten := RewriteDocLink(dest)
		if rewritten == dest {
			return match
		}
		return `href="` + rewritten + `"`
	})
}

var schemeRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*:`)

// RewriteDocLink rewrites a single link destination. Relative links whose
// path ends in .md, .markdown or .mmd have their extension replaced with
// .html. Everything else (absolute URLs, protocol-relative URLs, fragments,
// mailto links, …) is returned unchanged.
func RewriteDocLink(dest string) string {
	if dest == "" || strings.HasPrefix(dest, "#") || strings.HasPrefix(dest, "//") || schemeRe.MatchString(dest) {
		return dest
	}

	fragment := ""
	if i := strings.IndexByte(dest, '#'); i >= 0 {
		fragment = dest[i:]
		dest = dest[:i]
	}
	query := ""
	if i := strings.IndexByte(dest, '?'); i >= 0 {
		query = dest[i:]
		dest = dest[:i]
	}

	lower := strings.ToLower(dest)
	for _, ext := range []string{".md", ".markdown", ".mmd"} {
		if strings.HasSuffix(lower, ext) {
			dest = dest[:len(dest)-len(ext)] + ".html"
			break
		}
	}
	return dest + query + fragment
}

var mermaidBlockRe = regexp.MustCompile(`(?s)<pre><code class="language-mermaid">(.*?)</code></pre>`)
var preMermaidRe = regexp.MustCompile(`(?s)<pre class="mermaid">(.*?)</pre>`)

// rewriteMermaidBlocks converts fenced ```mermaid code blocks into
// <pre class="mermaid"> elements, which mermaid.js renders in the browser.
// The block contents stay HTML-escaped; mermaid reads them via textContent,
// which unescapes the entities. Documentation targets in <a href> inside
// diagram labels, and in `click` directives, are rewritten to the generated
// .html pages (see rewriteMermaidLinks). Local SVG references in <img>
// tags inside diagram labels are inlined per the context.
func rewriteMermaidBlocks(rendered string, ctx RenderContext) (string, error) {
	rendered = mermaidBlockRe.ReplaceAllString(rendered, `<pre class="mermaid">$1</pre>`)
	var b strings.Builder
	b.Grow(len(rendered))
	last := 0
	for _, loc := range preMermaidRe.FindAllStringSubmatchIndex(rendered, -1) {
		body := html.UnescapeString(rendered[loc[2]:loc[3]])
		body, err := inlineLocalSVGs(body, ctx)
		if err != nil {
			return "", err
		}
		body = rewriteMermaidLinks(body)
		b.WriteString(rendered[last:loc[0]])
		b.WriteString(`<pre class="mermaid">` + escapeMermaidText(body) + `</pre>`)
		last = loc[1]
	}
	b.WriteString(rendered[last:])
	return b.String(), nil
}

// MermaidPageHTML renders a standalone .mmd file as a page body with a
// heading and the diagram source in a <pre class="mermaid"> block. Local SVG
// references in <img> tags are inlined as data URIs (unless the context is
// the zero value).
func MermaidPageHTML(ctx RenderContext, title string, source []byte) (string, error) {
	src, err := inlineLocalSVGs(string(source), ctx)
	if err != nil {
		return "", fmt.Errorf("rendering mermaid page %q: %w", title, err)
	}
	src = rewriteMermaidLinks(src)
	return `<h1>` + html.EscapeString(title) + `</h1>` + "\n" +
		`<pre class="mermaid">` + escapeMermaidText(src) + `</pre>` + "\n", nil
}

// escapeMermaidText escapes the characters goldmark escapes inside code
// blocks, so the re-escaped Mermaid source reads exactly like the original
// block did: & < > and double quotes. Mermaid reads the block via
// textContent, which decodes all of these entities back out.
func escapeMermaidText(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return r.Replace(s)
}

// rewriteMermaidLinks rewrites documentation targets (.md/.markdown/.mmd) in
// a Mermaid block, mirroring what rewriteDocLinks does for Markdown. The
// generated site mirrors the source tree, so replacing the extension is all
// that is needed; fragments and query strings are preserved; absolute URLs,
// fragments, mailto: and friends pass through (RewriteDocLink rules).
//
// Two forms are rewritten:
//
//   - any href="..." token in the block — an <a href> inside a diagram
//     label (mermaid renders label HTML), or the target of a `click`
//     directive — is rewritten to the generated page;
//   - the shorthand node-click form `click A "target.md"` (no href keyword),
//     which Mermaid also supports.
//
// JavaScript callbacks (click A callback / click A call cb()) are left
// alone, exactly like Markdown links, and the HTML inside diagram labels
// keeps whatever attributes the author used (title, target, …).
func rewriteMermaidLinks(src string) string {
	lines := strings.Split(src, "\n")
	for i, line := range lines {
		line = rewriteMermaidHrefs(line)
		line = rewriteMermaidClickLine(line)
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

var hrefDestRe = regexp.MustCompile(`(?i)href\s*=?\s*(?:"([^"]*)"|'([^']*)')`)

// rewriteMermaidHrefs rewrites every href="..." token whose destination is
// a documentation file. It is line-local and preserves quote style.
func rewriteMermaidHrefs(line string) string {
	return hrefDestRe.ReplaceAllStringFunc(line, func(match string) string {
		m := hrefDestRe.FindStringSubmatch(match)
		dest := m[1]
		if dest == "" {
			dest = m[2]
		}
		if dest == "" {
			return match
		}
		return strings.Replace(match, dest, RewriteDocLink(dest), 1)
	})
}

// clickShorthandRe matches the shorthand node-click form of a mermaid
// click directive, in three parts:
//
//	^(\s*(?i:click)\s+)                 group 1 — leading keyword and whitespace
//	([^\s"']+(?:\s+[^\s"']+)*?\s+)      group 2 — node name and optional tooltip words,
//	                                            up to the destination (lazy, keeps it
//	                                            left of the final quoted string)
//	(?:"(...)"|'(...)'|(...))           groups 3/4/5 — quoted or bareword destination
var clickShorthandRe = regexp.MustCompile(`^(\s*(?i:click)\s+)([^\s"']+(?:\s+[^\s"']+)*?\s+)(?:"([^"\s]*)"|'([^'\s]*)'|([^\s"']+))`)

// rewriteMermaidClickLine handles the shorthand node-click form
// (`click A "target.md"`), which carries its URL as a bareword or quoted
// destination with no href keyword. Callback forms (click A callback /
// click A call fn()) are never rewritten; href targets are already handled
// by rewriteMermaidHrefs.
func rewriteMermaidClickLine(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	if len(trimmed) < 5 || !strings.EqualFold(trimmed[:5], "click") {
		return line
	}
	if len(trimmed) > 5 && trimmed[5] != ' ' && trimmed[5] != '\t' {
		return line
	}
	m := clickShorthandRe.FindStringSubmatch(line)
	if m != nil {
		targets := m[2]
		if !containsWord(targets, "callback") && !containsWord(targets, "call") && !containsWord(targets, "href") {
			quote := ""
			url := m[5]
			switch {
			case m[3] != "":
				quote, url = `"`, m[3]
			case m[4] != "":
				quote, url = `'`, m[4]
			}
			rewritten := m[1] + targets + quote + RewriteDocLink(url) + quote
			return rewritten + line[len(m[0]):]
		}
	}
	return line
}

func containsWord(s, word string) bool {
	for _, field := range strings.Fields(strings.ToLower(s)) {
		if field == word {
			return true
		}
	}
	return false
}

// HasMermaid reports whether rendered HTML contains a Mermaid diagram.
func HasMermaid(rendered string) bool {
	return strings.Contains(rendered, `<pre class="mermaid">`)
}

var imgTagRe = regexp.MustCompile(`<img\b[^>]*>`)

// inlineLocalSVGs replaces every <img src="..."> whose destination is a
// relative .svg file with the file's contents as an inline <svg> element,
// carrying over the destination's display attributes (width, height, aria-*,
// role, style) and escaping nothing a Mermaid label would not escape itself.
// The resulting pages never reference the icon files externally, and the
// icons compose with text inside a label exactly as Mermaid's own HTML
// labels expect (an inline <svg> is a perfectly ordinary element there).
// External URLs, data: URIs, fragments and non-SVG files are left untouched;
// a reference to a missing or unreadable SVG fails the render, and a path
// escaping the repository root is an error, never a read.
//
// In addition to filesystem paths, the destination "barnacle:<name>.svg"
// selects an icon from the Lucide set embedded in the binary (see
// inlineBarnacleIcon), so diagrams can use the bundled icons without any
// repository-relative path that could clash with the documented project's
// own layout. The zero context disables all of these rewrites, exactly as
// for filesystem references.
func inlineLocalSVGs(rendered string, ctx RenderContext) (string, error) {
	if ctx.Root == "" || ctx.SourceDir == "" {
		return rendered, nil
	}
	root := filepath.Clean(ctx.Root)
	srcDir := filepath.Clean(ctx.SourceDir)
	var b strings.Builder
	b.Grow(len(rendered))
	last := 0
	for _, loc := range imgTagRe.FindAllStringIndex(rendered, -1) {
		replacement, err := inlineImgTag(rendered[loc[0]:loc[1]], srcDir, root)
		if err != nil {
			return "", err
		}
		b.WriteString(rendered[last:loc[0]])
		b.WriteString(replacement)
		last = loc[1]
	}
	b.WriteString(rendered[last:])
	return b.String(), nil
}

// inlineImgTag inlines one <img> tag whose destination is a relative .svg
// file, or a barnacle:<name>.svg reference to the embedded Lucide set. Tags
// that reference anything else (external URLs, data: URIs, fragments,
// non-SVG files) are returned unchanged. A reference to a missing or
// unreadable SVG, or a path escaping the repository root, is an error —
// never a silent pass-through.
func inlineImgTag(tag, srcDir, root string) (string, error) {
	attrs := parseImgAttrs(tag)
	dest := attrs["src"]
	if strings.HasPrefix(strings.ToLower(dest), barnacleScheme) {
		return inlineBarnacleIcon(dest, attrs)
	}
	if !isLocalSVG(dest) {
		return tag, nil
	}
	abs := filepath.Clean(filepath.Join(srcDir, filepath.FromSlash(dest)))
	rel, err := filepath.Rel(root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("svg reference %q escapes the repository root", dest)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return "", fmt.Errorf("svg reference %q: %w", dest, err)
	}
	inline, err := svgToInline(data, attrs)
	if err != nil {
		return "", fmt.Errorf("svg reference %q: %w", dest, err)
	}
	return inline, nil
}

// barnacleScheme is the destination prefix that selects an icon from the
// Lucide set embedded in the binary: barnacle:<name>.svg. It is deliberately
// not a filesystem path, so a repository that happens to contain a
// "barnacle:" directory (or a file literally named "barnacle:folder.svg")
// is never misread.
const barnacleScheme = "barnacle:"

// inlineBarnacleIcon resolves a barnacle:<name>.svg destination against the
// embedded Lucide icons and inlines it like any local SVG. The name must be
// exactly the vendored icon name (lowercase, hyphenated) plus the .svg
// suffix — anything else is an error naming 'barnacle list-icons', never a
// pass-through. The .svg suffix keeps the shorthand parallel to the local
// file convention and leaves room for other embedded kinds later.
func inlineBarnacleIcon(dest string, attrs map[string]string) (string, error) {
	rest := dest[len(barnacleScheme):]
	if !strings.HasSuffix(strings.ToLower(rest), ".svg") {
		return "", fmt.Errorf("barnacle icon reference %q: expected the form \"barnacle:<name>.svg\" (run 'barnacle list-icons' to see the embedded set)", dest)
	}
	name := rest[:len(rest)-len(".svg")]
	data, err := lucide.Icon(name)
	if err != nil {
		return "", fmt.Errorf("barnacle icon reference %q: %w", dest, err)
	}
	inline, err := svgToInline(data, attrs)
	if err != nil {
		return "", fmt.Errorf("barnacle icon reference %q: %w", dest, err)
	}
	return inline, nil
}

var attrRe = regexp.MustCompile(`([A-Za-z-]+)\s*=\s*(?:"([^"]*)"|'([^']*)')`)

// parseImgAttrs extracts the attribute map of an <img> tag (name → value,
// single- or double-quoted). Boolean attributes without a value are absent.
func parseImgAttrs(tag string) map[string]string {
	out := map[string]string{}
	for _, m := range attrRe.FindAllStringSubmatch(tag, -1) {
		v := m[2]
		if v == "" {
			v = m[3]
		}
		out[strings.ToLower(m[1])] = v
	}
	return out
}

var svgOpenRe = regexp.MustCompile(`(?s)^\s*(?:<\?xml[^>]*\?>)?\s*(?:<!--.*?-->\s*)*<svg\b([^>]*)>\s*(.*)\s*</svg>\s*$`)

var shapeRe = regexp.MustCompile(`<(?P<tag>path|circle|rect|line|polyline|polygon|ellipse)(\s[^/>]*)?(/)?>`)

// visualProps are the properties mermaid's default theme paints over with
// #ECECFF fill and #9370DB stroke: fill, stroke, stroke-width, stroke-linecap,
// stroke-linejoin. When the file declares any of them on the root <svg>,
// they are projected onto every shape so the file's intent renders under
// mermaid's paint. When the file declares none, the lucide defaults apply.
var visualProps = []string{"fill", "stroke", "stroke-width", "stroke-linecap", "stroke-linejoin"}

// displayDimensions picks the rendered size of an inlined icon or diagram.
// The <img>'s own width and height win (either one alone squares the
// icon). Otherwise a file declaring both width and height keeps them
// (CSS max-width keeps them inside the content column). Otherwise a
// square viewBox keeps the 20px house default (icon-shaped files, the
// common case for the embedded Lucide set), while a non-square viewBox
// becomes a fluid width:100% / height:auto, so wide diagrams —
// architecture drawings, timing charts — fill the column instead of
// being squeezed into 20px squares. A missing viewBox also falls back
// to the house default.
func displayDimensions(imgAttrs, fileAttrs map[string]string) (string, string) {
	width, hasW := imgAttrs["width"]
	height, hasH := imgAttrs["height"]
	switch {
	case hasW && hasH:
		return width, height
	case hasW:
		return width, width
	case hasH:
		return height, height
	}
	fw, hasFW := fileAttrs["width"]
	fh, hasFH := fileAttrs["height"]
	if hasFW && hasFH && fw != "" && fh != "" {
		return fw, fh
	}
	if viewBox, ok := fileAttrs["viewbox"]; ok {
		if vbW, vbH, ok := parseViewBoxSize(viewBox); ok {
			if math.Abs(vbW-vbH) < 1e-6 {
				return "20px", "20px"
			}
			return "100%", "auto"
		}
	}
	return "20px", "20px"
}

// parseViewBoxSize reads the width and height of a viewBox attribute
// ("minx miny width height").
func parseViewBoxSize(viewBox string) (float64, float64, bool) {
	parts := strings.Fields(strings.ReplaceAll(viewBox, ",", " "))
	if len(parts) != 4 {
		return 0, 0, false
	}
	w, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, 0, false
	}
	h, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return 0, 0, false
	}
	if w <= 0 || h <= 0 {
		return 0, 0, false
	}
	return w, h, true
}

// svgToInline renders a file's bytes as an inline <svg> element. Display
// attributes carried by the <img> win (width/height), otherwise the SVG's
// own width/height/viewBox settings are scaled to the house default of 20.
// viewBox is always retained or synthesized; content passes through verbatim.
func svgToInline(data []byte, imgAttrs map[string]string) (string, error) {
	m := svgOpenRe.FindSubmatch(data)
	if m == nil {
		return "", fmt.Errorf("not a well-formed svg document")
	}
	fileAttrs := parseImgAttrs("<svg " + string(m[1]) + ">")
	inner := string(m[2])

	// Project the file's visual attributes (or the stroke-only defaults) onto
	// every shape as inline styles, so mermaid's .node path theme paint can't
	// repaint stroke-only icons into solid lavender blobs. A shape that
	// already declares the property keeps its own value.
	projection := make(map[string]string, len(visualProps))
	for _, k := range visualProps {
		if v, ok := fileAttrs[k]; ok && v != "" {
			projection[k] = v
		}
	}
	if _, ok := projection["fill"]; !ok {
		projection["fill"] = "none"
	}
	if _, ok := projection["stroke"]; !ok {
		projection["stroke"] = "currentColor"
	}
	if len(projection) > 0 {
		inner = projectVisual(inner, projection)
	}

	attrs := fileAttrs
	attrs["width"], attrs["height"] = displayDimensions(imgAttrs, fileAttrs)
	if attrs["viewbox"] == "" {
		attrs["viewbox"] = "0 0 24 24" // sensible default for icon-shaped files
	}
	// Display overrides from the img tag itself; everything else passes through.
	for k, v := range imgAttrs {
		if k == "src" || k == "width" || k == "height" {
			continue // handled above
		}
		attrs[k] = v
	}
	// Author-provided alt wins as an accessible label; otherwise mark the icon
	// as decorative. The file's own aria-* attributes are never overwritten.
	if alt, ok := imgAttrs["alt"]; ok && strings.TrimSpace(alt) != "" {
		if _, ok := attrs["aria-label"]; !ok {
			attrs["aria-label"] = alt
		}
	} else if _, ok := attrs["aria-label"]; !ok {
		attrs["aria-hidden"] = "true"
	}
	delete(attrs, "alt")

	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("<svg")
	for _, k := range keys {
		fmt.Fprintf(&b, " %s=\"%s\"", k, html.EscapeString(attrs[k]))
	}
	b.WriteString(">")
	fmt.Fprint(&b, strings.Join(strings.Fields(inner), " "))
	b.WriteString("</svg>")
	return b.String(), nil
}

// projectVisual appends the projection styles to every shape, preserving
// whatever the shape itself already declares (fill, stroke, stroke-width
// and their inline-style forms). This is the minimal interjection meant
// to stop mermaid's theme paint from retinting stroke-only glyphs.
func projectVisual(inner string, proj map[string]string) string {
	return shapeRe.ReplaceAllStringFunc(inner, func(tag string) string {
		match := shapeRe.FindStringSubmatch(tag)
		name, head, tail := match[1], match[2], match[3]
		close := "/"
		if tail == "" {
			close = ""
		}
		shapeAttrs := parseImgAttrs("<" + name + head + close + ">")
		keep := make([]string, 0, len(proj))
		for _, k := range visualProps {
			v, ok := proj[k]
			if !ok {
				continue
			}
			if _, own := shapeAttrs[k]; own {
				continue // shape declares its own
			}
			if styleContent, own := shapeAttrs["style"]; own && styleHasProp(styleContent, k) {
				continue
			}
			keep = append(keep, k+":"+v)
		}
		if len(keep) == 0 {
			return tag
		}
		if styleContent, ok := shapeAttrs["style"]; ok {
			merged := mergeStyle(styleContent, keep)
			head = styleAttrRe.ReplaceAllString(head, `style="`+merged+`"`)
			return "<" + name + head + close + ">"
		}
		return "<" + name + head + ` style="` + strings.Join(keep, ";") + `"` + close + ">"
	})
}

var styleAttrRe = regexp.MustCompile(`style\s*=\s*(?:"[^"]*"|'[^']*')`)

// mergeStyle appends the projection declarations to an inline style
// property list, preserving the declarations the shape itself had.
func mergeStyle(style string, extra []string) string {
	return strings.Join(append(splitStyle(style), extra...), ";")
}

// splitStyle splits an inline style property list on ';'.
func splitStyle(style string) []string {
	return strings.FieldsFunc(style, func(r rune) bool { return r == ';' })
}

// styleHasProp reports whether a style property list explicitly declares key.
func styleHasProp(style, key string) bool {
	for _, chunk := range splitStyle(style) {
		if i := strings.Index(chunk, ":"); i >= 0 && strings.TrimSpace(chunk[:i]) == key {
			return true
		}
	}
	return false
}

func isLocalSVG(dest string) bool {
	if dest == "" || strings.HasPrefix(dest, "#") || strings.HasPrefix(dest, "//") || schemeRe.MatchString(dest) {
		return false
	}
	// Fragments and queries make no sense for inlining; treat the whole
	// destination as external to the inlining pass.
	if strings.ContainsAny(dest, "?#") {
		return false
	}
	return strings.HasSuffix(strings.ToLower(dest), ".svg")
}
