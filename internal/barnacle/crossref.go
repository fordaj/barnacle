package barnacle

import (
	"fmt"
	"html"
	"regexp"
	"sort"
	"strings"
)

// Cross-references and captions in the pandoc-crossref style.
//
// Two block-level constructs are recognized in rendered Markdown output:
//
//   - a paragraph that contains nothing but an image whose definition is
//     followed immediately by an attribute {#fig:label} —
//     ![Caption](image.png){#fig:label}
//     — becomes a numbered, labelled <figure> with a <figcaption>;
//   - a table immediately followed by a one-line paragraph that starts
//     with ":" — `: Caption {#tbl:label}` — gains a numbered, labelled
//     native <caption> element (the {#tbl:label} attribute is optional).
//
// Figures and tables are numbered independently per page, in document
// order. Every captioned figure and every captioned table is numbered.
// Labels are needed only for cross-references; a caption without a label
// is still numbered and still emits the caption text.
//
// Cross-references use the pandoc citation syntax: @fig:label and
// @tbl:label, optionally bracketed ([@fig:label]). Both are replaced
// with a same-page link to the labelled element carrying the element's
// number ("Figure 2" / "Table 1"). References to undefined labels are
// left alone; a duplicate label within one page is an error.
//
// The pass also wraps every <table> in a <div class="table-scroll"> so
// wide tables stay readable on narrow screens (horizontal overflow
// scrolls inside the wrapper instead of widening the page).
//
// Cross-references are not rewritten inside <pre> or <code> elements,
// and the figure/table patterns cannot match inside them either (their
// contents are HTML-escaped by the Markdown renderer).

// captionKind distinguishes the two numbered object kinds.
type captionKind int

const (
	captionFigure captionKind = iota
	captionTable
)

var (
	// A figure is an image alone in its paragraph, immediately followed
	// by the {#fig:label} attribute (image block and label cannot be
	// separated by spaces, exactly like pandoc-crossref).
	figureRe = regexp.MustCompile(`(?s)<p>\s*(<img\b[^>]*?>)\s*\{#(fig:[A-Za-z0-9_-]+)\}\s*</p>`)

	// A table caption is the one-line paragraph right after the table:
	// ": caption text {#tbl:label}". The label attribute is optional.
	// The caption body is inline HTML (goldmark may have emitted <code>
	// etc. in it) and is carried into the <caption> element verbatim.
	tableCaptionRe = regexp.MustCompile(`(?s)(<table\b[^>]*>)(.*?)(</table>)\s*<p>:\s*(.*?)\s*(?:\{#(tbl:[A-Za-z0-9_-]+)\})?\s*</p>`)

	tableRe = regexp.MustCompile(`(?s)<table\b.*?</table>`)

	// A cross-reference: @fig:label / @tbl:label, optionally wrapped in
	// balanced square brackets ([@fig:label]). The label charset excludes
	// periods so a sentence-final period is not swallowed into the
	// reference. The leading boundary keeps e-mail addresses (user@fig:x)
	// from matching; the boundary character is re-emitted verbatim.
	xrefRe = regexp.MustCompile(`(^|[^A-Za-z0-9_])(\[@(?:fig|tbl):[A-Za-z0-9_-]+\]|@(?:fig|tbl):[A-Za-z0-9_-]+)`)

	// Segments that must not be touched by cross-reference rewriting:
	// raw blocks and inline code (their contents may legitimately
	// contain @fig:label as documentation), and HTML tags (an @fig:…
	// inside an attribute value is not a reference).
	xrefProtectRe = regexp.MustCompile(`(?s)<pre\b.*?</pre>|<code\b.*?</code>|<[^>]+>`)
)

// captionItem is one numbered caption found in the document.
type captionItem struct {
	kind  captionKind
	pos   int // byte offset of the match in the rendered HTML
	end   int
	label string
	num   int // assigned per kind, starting at 1

	// figure fields
	imgTag  string
	caption string // unescaped alt-text caption

	// table fields
	tableOpen   string
	tableBody   string
	tableClose  string
	captionHTML string
}

// numberedCaption maps a label to its assigned caption.
type numberedCaption struct {
	kind captionKind
	num  int
}

// applyCrossrefs numbers captioned figures and tables, wraps tables in a
// scroll container and rewrites cross-references. It reports an error on
// duplicate labels within the page.
func applyCrossrefs(rendered string) (string, error) {
	items, err := collectCaptionItems(rendered)
	if err != nil {
		return "", err
	}
	numbered := assignNumbers(items)

	// Rebuild the document in one pass, substituting captions.
	var b strings.Builder
	b.Grow(len(rendered) + 256)
	last := 0
	for _, item := range items {
		b.WriteString(rendered[last:item.pos])
		switch item.kind {
		case captionFigure:
			b.WriteString(renderFigureCaption(item))
		case captionTable:
			b.WriteString(renderTableCaption(item))
		}
		last = item.end
	}
	b.WriteString(rendered[last:])

	result := b.String()
	result = tableRe.ReplaceAllString(result, `<div class="table-scroll">$0</div>`)
	result = rewriteXRefs(result, numbered)
	return result, nil
}

// collectCaptionItems finds every figure and table caption in the rendered
// HTML, checks labels for duplicates and returns them in document order.
func collectCaptionItems(rendered string) ([]captionItem, error) {
	var items []captionItem
	for _, loc := range figureRe.FindAllStringSubmatchIndex(rendered, -1) {
		imgTag := rendered[loc[2]:loc[3]]
		caption := ""
		if alt, ok := parseImgAttrs(imgTag)["alt"]; ok {
			caption = html.UnescapeString(alt)
		}
		items = append(items, captionItem{
			kind:    captionFigure,
			pos:     loc[0],
			end:     loc[1],
			label:   rendered[loc[4]:loc[5]],
			imgTag:  imgTag,
			caption: caption,
		})
	}
	for _, loc := range tableCaptionRe.FindAllStringSubmatchIndex(rendered, -1) {
		label := ""
		if loc[10] >= 0 {
			label = rendered[loc[10]:loc[11]]
		}
		items = append(items, captionItem{
			kind:        captionTable,
			pos:         loc[0],
			end:         loc[1],
			label:       label,
			tableOpen:   rendered[loc[2]:loc[3]],
			tableBody:   rendered[loc[4]:loc[5]],
			tableClose:  rendered[loc[6]:loc[7]],
			captionHTML: strings.TrimSpace(rendered[loc[8]:loc[9]]),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].pos < items[j].pos })

	seen := map[string]captionKind{}
	for _, item := range items {
		if item.label == "" {
			continue
		}
		if prev, dup := seen[item.label]; dup {
			return nil, fmt.Errorf("duplicate caption label %q (%s and %s)", item.label, kindName(prev), kindName(item.kind))
		}
		seen[item.label] = item.kind
	}
	return items, nil
}

// assignNumbers walks the items in document order and numbers each one
// per kind, starting at 1. It also returns the label → caption map used
// by cross-reference rewriting.
func assignNumbers(items []captionItem) map[string]numberedCaption {
	var figures, tables int
	numbered := map[string]numberedCaption{}
	for i := range items {
		switch items[i].kind {
		case captionFigure:
			figures++
			items[i].num = figures
		default:
			tables++
			items[i].num = tables
		}
		if items[i].label != "" {
			numbered[items[i].label] = numberedCaption{kind: items[i].kind, num: items[i].num}
		}
	}
	return numbered
}

func kindName(kind captionKind) string {
	if kind == captionTable {
		return "table"
	}
	return "figure"
}

func renderFigureCaption(item captionItem) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<figure class="figure" id="%s">%s`, html.EscapeString(item.label), item.imgTag)
	b.WriteString(`<figcaption><span class="caption-label">`)
	fmt.Fprintf(&b, `Figure %d:`, item.num)
	b.WriteString(`</span>`)
	if caption := strings.TrimSpace(item.caption); caption != "" {
		b.WriteString(" ")
		b.WriteString(html.EscapeString(caption))
	}
	b.WriteString(`</figcaption></figure>`)
	return b.String()
}

func renderTableCaption(item captionItem) string {
	var b strings.Builder
	if item.label != "" {
		open := item.tableOpen
		if idx := strings.Index(open, " id="); idx < 0 {
			open = strings.Replace(open, "<table", `<table id="`+html.EscapeString(item.label)+`"`, 1)
		}
		b.WriteString(open)
	} else {
		b.WriteString(item.tableOpen)
	}
	b.WriteString(`<caption><span class="caption-label">`)
	fmt.Fprintf(&b, `Table %d:`, item.num)
	b.WriteString(`</span>`)
	if item.captionHTML != "" {
		b.WriteString(" ")
		b.WriteString(item.captionHTML)
	}
	b.WriteString(`</caption>`)
	b.WriteString(item.tableBody)
	b.WriteString(item.tableClose)
	return b.String()
}

// rewriteXRefs replaces @fig:label / @tbl:label (optionally bracketed)
// with same-page links to the numbered caption. References to unknown
// labels are left untouched. Matches inside <pre>, <code> and inside
// tags are protected.
func rewriteXRefs(rendered string, numbered map[string]numberedCaption) string {
	matches := xrefProtectRe.FindAllStringIndex(rendered, -1)
	var b strings.Builder
	b.Grow(len(rendered) + 128)
	last := 0
	for _, loc := range matches {
		b.WriteString(replaceXRefs(rendered[last:loc[0]], numbered))
		b.WriteString(rendered[loc[0]:loc[1]])
		last = loc[1]
	}
	b.WriteString(replaceXRefs(rendered[last:], numbered))
	return b.String()
}

func replaceXRefs(text string, numbered map[string]numberedCaption) string {
	return xrefRe.ReplaceAllStringFunc(text, func(match string) string {
		m := xrefRe.FindStringSubmatch(match)
		boundary, token := m[1], m[2]
		inner := strings.TrimPrefix(token, "[")
		inner = strings.TrimSuffix(inner, "]")
		inner = strings.TrimPrefix(inner, "@")
		kind, rest, _ := strings.Cut(inner, ":")
		label := kind + ":" + rest
		number, ok := numbered[label]
		if !ok {
			return match
		}
		word := "Table"
		if number.kind == captionFigure {
			word = "Figure"
		}
		return fmt.Sprintf(`%s<a class="xref" href="#%s">%s %d</a>`, boundary, html.EscapeString(label), word, number.num)
	})
}
