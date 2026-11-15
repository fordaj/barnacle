package barnacle

import (
	"bytes"
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

// rewriteMermaidBlocks converts fenced ```mermaid code blocks into
// <pre class="mermaid"> elements, which mermaid.js renders in the browser.
// The block contents stay HTML-escaped; mermaid reads them via textContent,
// which unescapes the entities. Documentation targets in `click` directives
// are rewritten to the generated .html pages first (see rewriteMermaidLinks).
func rewriteMermaidBlocks(rendered string) string {
	return mermaidBlockRe.ReplaceAllStringFunc(rendered, func(block string) string {
		sub := mermaidBlockRe.FindStringSubmatch(block)
		body := html.UnescapeString(sub[1])
		body = rewriteMermaidLinks(body)
		return `<pre class="mermaid">` + escapeMermaidText(body) + `</pre>`
	})
}

// MermaidPageHTML renders a standalone .mmd file as a page body with a
// heading and the diagram source in a <pre class="mermaid"> block.
func MermaidPageHTML(title string, source []byte) string {
	src := rewriteMermaidLinks(string(source))
	return `<h1>` + html.EscapeString(title) + `</h1>` + "\n" +
		`<pre class="mermaid">` + escapeMermaidText(src) + `</pre>` + "\n"
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
// Mermaid `click` directives so clickable diagram nodes navigate to the
// generated .html pages, mirroring what rewriteDocLinks does for Markdown.
// Clicking a node is a relative navigation resolved against the page URL, and
// the generated site mirrors the source tree, so swapping the extension is
// enough — the same invariant Markdown links rely on.
//
// Two forms are supported, matching Mermaid's flowchart syntax:
//
//	click NodeId href "target.md" ["tooltip"] [_blank]
//	click NodeId "target.md" ["tooltip"] [_blank]
//
// JavaScript callback forms (click A callback / click A call cb()) are left
// alone; the URLs inside an href are always rewritten, and the URL in the
// shorthand form is rewritten only when it contains no whitespace. Targets
// are rewritten only for documentation extensions, so tooltips and external
// links are never touched.
func rewriteMermaidLinks(src string) string {
	lines := strings.Split(src, "\n")
	for i, line := range lines {
		lines[i] = rewriteMermaidClickLine(line)
	}
	return strings.Join(lines, "\n")
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
