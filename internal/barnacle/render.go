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
// generated .html pages, and Mermaid code fences are turned into
// <pre class="mermaid"> blocks so they render in the browser.
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
// The source inside the block stays HTML-escaped; mermaid reads it via
// textContent, which unescapes the entities.
func rewriteMermaidBlocks(rendered string) string {
	return mermaidBlockRe.ReplaceAllString(rendered, `<pre class="mermaid">$1</pre>`)
}

// MermaidPageHTML renders a standalone .mmd file as a page body containing
// the diagram source in a <pre class="mermaid"> block.
func MermaidPageHTML(source []byte) string {
	return `<pre class="mermaid">` + html.EscapeString(string(source)) + `</pre>` + "\n"
}

// HasMermaid reports whether rendered HTML contains a Mermaid diagram.
func HasMermaid(rendered string) bool {
	return strings.Contains(rendered, `<pre class="mermaid">`)
}
