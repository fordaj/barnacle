package barnacle

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

// treeNode is a directory or file in the navigation tree.
type treeNode struct {
	name     string
	isDir    bool
	doc      *Document
	children []*treeNode
}

// buildTree builds a directory tree from the (sorted) document list.
func buildTree(docs []Document) *treeNode {
	root := &treeNode{isDir: true}
	for i := range docs {
		doc := docs[i]
		parts := strings.Split(doc.RelPath, "/")
		current := root
		for j, part := range parts {
			last := j == len(parts)-1
			var child *treeNode
			for _, c := range current.children {
				if c.name == part && c.isDir == !last {
					child = c
					break
				}
			}
			if child == nil {
				child = &treeNode{name: part, isDir: !last}
				current.children = append(current.children, child)
			}
			if last {
				d := doc
				child.doc = &d
			}
			current = child
		}
	}
	return root
}

var inlineCodeRe = regexp.MustCompile("`([^`\n]+)`")

// renderInlineCode escapes s for HTML and turns single-backtick spans into
// <code> elements. It is used for raw Markdown fragments that are shown as
// HTML — page titles in the sidebar and the landing page — where backticks
// would otherwise render literally. Only inline code is honored; anything
// else (links, emphasis, …) stays escaped text.
func renderInlineCode(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 24)
	last := 0
	for _, loc := range inlineCodeRe.FindAllStringSubmatchIndex(s, -1) {
		b.WriteString(html.EscapeString(s[last:loc[0]]))
		b.WriteString("<code>")
		b.WriteString(html.EscapeString(s[loc[2]:loc[3]]))
		b.WriteString("</code>")
		last = loc[1]
	}
	b.WriteString(html.EscapeString(s[last:]))
	return b.String()
}

// renderNav renders the sidebar navigation as nested lists. Links are
// relative to the page at currentOutRel (slash-separated path relative to
// the site root). Page titles render Markdown inline code as <code> —
// they carry raw Markdown (see scan.go's extractTitle), and the sidebar
// is an HTML context.
func renderNav(root *treeNode, currentOutRel string) string {
	var b strings.Builder
	b.WriteString(`<ul class="nav-tree">`)
	renderNavNodes(&b, root, currentOutRel)
	b.WriteString(`</ul>`)
	return b.String()
}

func renderNavNodes(b *strings.Builder, node *treeNode, currentOutRel string) {
	for _, child := range node.children {
		if child.isDir {
			fmt.Fprintf(b, `<li class="nav-dir"><span class="nav-dir-name">%s</span><ul>`, html.EscapeString(child.name))
			renderNavNodes(b, child, currentOutRel)
			b.WriteString(`</ul></li>`)
			continue
		}
		doc := child.doc
		href := relativeURL(currentOutRel, doc.OutputRel)
		current := ""
		class := "nav-link"
		if doc.OutputRel == currentOutRel {
			current = ` aria-current="page"`
			class += " current"
		}
		fmt.Fprintf(b, `<li><a class="%s"%s href="%s">%s</a></li>`,
			class, current, html.EscapeString(href), renderInlineCode(doc.Title))
	}
}

// renderIndex renders the body of the landing page: an overview of every
// documented page. The site title and page titles render inline code like
// in the sidebar (HTML contexts); the browser tab keeps raw text.
func renderIndex(title string, root *treeNode) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<h1>%s</h1>`, renderInlineCode(title))
	b.WriteString("<p>Documentation lives alongside the code it describes. Pick a page to explore, or follow the links between documents.</p>")
	if len(root.children) == 0 {
		b.WriteString(`<p class="empty-state">No documentation files found. Add <code>.md</code> or <code>.mmd</code> files next to your source code and run barnacle again.</p>`)
		return b.String()
	}
	b.WriteString(`<ul class="doc-index">`)
	renderIndexNodes(&b, root, "")
	b.WriteString(`</ul>`)
	return b.String()
}

func renderIndexNodes(b *strings.Builder, node *treeNode, prefix string) {
	for _, child := range node.children {
		if child.isDir {
			fmt.Fprintf(b, `<li class="doc-index-dir"><span class="nav-dir-name">%s</span><ul>`, html.EscapeString(child.name))
			renderIndexNodes(b, child, prefix+child.name+"/")
			b.WriteString(`</ul></li>`)
			continue
		}
		doc := child.doc
		fmt.Fprintf(b, `<li><a href="%s">%s</a> <span class="doc-path">%s</span></li>`,
			html.EscapeString(relativeURL("index.html", doc.OutputRel)),
			renderInlineCode(doc.Title),
			html.EscapeString(doc.RelPath))
	}
}
