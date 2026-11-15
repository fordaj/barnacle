package barnacle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldhtml "github.com/yuin/goldmark/renderer/html"
)

var markdown = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(goldhtml.WithUnsafe()),
)

// MarkdownToHTML renders a Markdown document to an HTML fragment. Links to
// other documentation files (.md/.mmd) are rewritten to point at the
// generated .html pages — in ordinary links and inside Mermaid `click`
// directives — and Mermaid code fences are turned into <pre class="mermaid">
// blocks so they render in the browser.
func MarkdownToHTML(source []byte) (string, error) {
	var buf bytes.Buffer
	if err := markdown.Convert(source, &buf); err != nil {
		return "", fmt.Errorf("rendering markdown: %w", err)
	}
	content := buf.String()
	content = rewriteDocLinks(content)
	content = rewriteMermaidBlocks(content)
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
// which unescapes the entities. Documentation targets in `click` directives
// and `%% barnacle: link` directives are rewritten to the generated .html
// pages (see rewriteMermaidLinks), exactly once per block — fenced or
// hand-authored alike. Blocks with `%% barnacle: link` directives are
// followed by a JSON manifest the page template uses to wire the clicks.
func rewriteMermaidBlocks(rendered string) string {
	rendered = mermaidBlockRe.ReplaceAllString(rendered, `<pre class="mermaid">$1</pre>`)
	return preMermaidRe.ReplaceAllStringFunc(rendered, func(block string) string {
		body := html.UnescapeString(preMermaidRe.FindStringSubmatch(block)[1])
		body, directives := rewriteMermaidLinks(body)
		return `<pre class="mermaid">` + escapeMermaidText(body) + `</pre>` + edgeLinksScript(directives)
	})
}

// MermaidPageHTML renders a standalone .mmd file as a page body with a
// heading and the diagram source in a <pre class="mermaid"> block.
func MermaidPageHTML(title string, source []byte) string {
	src, directives := rewriteMermaidLinks(string(source))
	return `<h1>` + html.EscapeString(title) + `</h1>` + "\n" +
		`<pre class="mermaid">` + escapeMermaidText(src) + `</pre>` + "\n" +
		edgeLinksScript(directives)
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
// Mermaid `click` directives and `%% barnacle: link` directives so clickable
// diagram shapes navigate to the generated .html pages, mirroring what
// rewriteDocLinks does for Markdown. Clicking a node is a relative navigation
// resolved against the page URL, and the generated site mirrors the source
// tree, so swapping the extension is enough — the same invariant Markdown
// links rely on.
//
// Two node-click forms are supported, matching Mermaid's flowchart syntax:
//
//	click NodeId href "target.md" ["tooltip"] [_blank]
//	click NodeId "target.md" ["tooltip"] [_blank]
//
// JavaScript callback forms (click A callback / click A call cb()) are left
// alone; the URLs inside an href are always rewritten, and the URL in the
// shorthand form is rewritten only when it contains no whitespace. Targets
// are rewritten only for documentation extensions, so tooltips and external
// links are never touched.
//
// The .md/.mmd targets of `%% barnacle: link` directives are rewritten the
// same way, and the directives are returned as edgeLinkDirective values so
// the caller can emit the JSON manifest barnacle-edges.js reads. The
// directive lines stay in the diagram source (mermaid renders them as
// comments).
func rewriteMermaidLinks(src string) (string, []edgeLinkDirective) {
	lines := strings.Split(src, "\n")
	var directives []edgeLinkDirective
	for i, line := range lines {
		line = rewriteMermaidClickLine(line)
		if d, ok := parseEdgeLinkDirective(line); ok {
			d.Href = RewriteDocLink(d.Href)
			line = formatEdgeLinkDirective(d)
			directives = append(directives, d)
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n"), directives
}

// edgeLinkDirective is a `%% barnacle: link` directive from a Mermaid
// diagram. The directive is a mermaid comment (mermaid renders nothing for
// it); barnacle parses it at build time, rewrites the target, and emits a
// JSON manifest the page template places next to the diagram for
// barnacle-edges.js to wire up in the browser.
//
// Syntax:
//
//	%% barnacle: link e1 href "target.md" ["tooltip"] [_blank]
//	%% barnacle: link "send(cmd)" href "target.md" ["tooltip"] [_blank]
//
// A bare key matches the edge's data-id (the explicit `e1` from
// `A e1@--> B`, or the auto-assigned id like `edge0`/`L_A_B_0`); a quoted
// key matches the edge's label text. Targets ending in .md/.markdown/.mmd
// are rewritten to .html like everything else.
type edgeLinkDirective struct {
	Key     string `json:"key"`
	By      string `json:"by"`
	Href    string `json:"href"`
	Tooltip string `json:"tooltip,omitempty"`
	Blank   bool   `json:"blank,omitempty"`
}

var edgeLinkRe = regexp.MustCompile(`^(\s*%%\s*barnacle:\s*link\s+)(?:"([^"]*)"|'([^']*)'|([^\s"']+))\s+href\s+(?:"([^"]*)"|'([^']*)'|([^\s"']+))(?:\s+(?:"([^"]*)"|'([^']*)'))?(?:\s+(_blank|_self))?\s*$`)

func parseEdgeLinkDirective(line string) (edgeLinkDirective, bool) {
	m := edgeLinkRe.FindStringSubmatch(line)
	if m == nil {
		return edgeLinkDirective{}, false
	}
	var d edgeLinkDirective
	switch {
	case m[2] != "":
		d.By, d.Key = "label", m[2]
	case m[3] != "":
		d.By, d.Key = "label", m[3]
	default:
		d.By, d.Key = "id", m[4]
	}
	switch {
	case m[5] != "":
		d.Href = m[5]
	case m[6] != "":
		d.Href = m[6]
	default:
		d.Href = m[7]
	}
	switch {
	case m[8] != "":
		d.Tooltip = m[8]
	case m[9] != "":
		d.Tooltip = m[9]
	}
	if strings.EqualFold(m[10], "_blank") {
		d.Blank = true
	}
	if d.Key == "" || d.Href == "" {
		return edgeLinkDirective{}, false
	}
	return d, true
}

// formatEdgeLinkDirective renders a directive back into the mermaid source.
// The directive is a comment to mermaid; keeping it in the output keeps the
// diagram self-describing when the page source is viewed.
func formatEdgeLinkDirective(d edgeLinkDirective) string {
	norm := func(s string) string {
		s = strings.ReplaceAll(s, "\r", " ")
		s = strings.ReplaceAll(s, "\n", " ")
		s = strings.ReplaceAll(s, `"`, `'`)
		return s
	}
	key := norm(d.Key)
	if d.By == "label" {
		key = `"` + key + `"`
	}
	var b strings.Builder
	b.WriteString("%% barnacle: link ")
	b.WriteString(key)
	fmt.Fprintf(&b, ` href "%s"`, norm(d.Href))
	if d.Tooltip != "" {
		fmt.Fprintf(&b, ` "%s"`, norm(d.Tooltip))
	}
	if d.Blank {
		b.WriteString(" _blank")
	}
	return b.String()
}

// edgeLinksScript renders the JSON manifest barnacle-edges.js consumes into
// a <script> element placed right after the diagram. "<" is escaped so the
// manifest can never break out of the script element.
func edgeLinksScript(directives []edgeLinkDirective) string {
	if len(directives) == 0 {
		return ""
	}
	data, err := json.Marshal(edgeLinksManifest{Edges: directives})
	if err != nil {
		return ""
	}
	data = bytes.ReplaceAll(data, []byte("<"), []byte(`\u003c`))
	return `<script type="application/json" class="barnacle-edge-links">` + string(data) + "</script>\n"
}

type edgeLinksManifest struct {
	Edges []edgeLinkDirective `json:"edges"`
}

// HasEdgeLinks reports whether rendered HTML contains an edge-links manifest.
func HasEdgeLinks(rendered string) bool {
	return strings.Contains(rendered, `class="barnacle-edge-links"`)
}

var (
	clickHrefRe      = regexp.MustCompile(`(?i)(href\s+)(?:"([^"\s]*)"|'([^'\s]*)'|([^\s"']+))`)
	clickShorthandRe = regexp.MustCompile(`^(\s*(?i:click)\s+)([^\s"']+(?:\s+[^\s"']+)*?\s+)(?:"([^"\s]*)"|'([^'\s]*)'|([^\s"']+))`)
)

func rewriteMermaidClickLine(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	if len(trimmed) < 5 || !strings.EqualFold(trimmed[:5], "click") {
		return line
	}
	if len(trimmed) > 5 && trimmed[5] != ' ' && trimmed[5] != '\t' {
		return line
	}
	// Form 1: click <targets> href "url" ["tooltip"] [_blank]
	line = clickHrefRe.ReplaceAllStringFunc(line, func(match string) string {
		m := clickHrefRe.FindStringSubmatch(match)
		quote := ""
		url := m[4]
		switch {
		case m[2] != "":
			quote, url = `"`, m[2]
		case m[3] != "":
			quote, url = `'`, m[3]
		}
		// Preserve the quote style (or lack of it) of the original token.
		return m[1] + quote + RewriteDocLink(url) + quote
	})
	// Form 2: click <targets> "url" ["tooltip"] [_blank]
	// Only when the targets are bare words; skip callback forms.
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
			line = rewritten + line[len(m[0]):]
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
