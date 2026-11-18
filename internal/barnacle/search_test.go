package barnacle

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildSearchIndexExtractsSections(t *testing.T) {
	docs := []Document{
		{RelPath: "guide/usage.md", OutputRel: "guide/usage.html", Title: "Usage"},
	}
	pages := []string{`<h1 id="usage">Usage</h1>
<p>Install the thing first.</p>
<h2 id="configuration">Configuration</h2>
<p>Set <code>option = true</code> in the file.</p>
<h2 id="troubleshooting">Troubleshooting</h2>
<p>Restart &amp; retry.</p>
`}
	index := buildSearchIndex(docs, pages)
	if len(index) != 1 {
		t.Fatalf("got %d pages, want 1", len(index))
	}
	page := index[0]
	if page.URL != "guide/usage.html" {
		t.Errorf("URL = %q, want guide/usage.html", page.URL)
	}
	if page.Title != "Usage" {
		t.Errorf("Title = %q, want Usage", page.Title)
	}
	byAnchor := map[string]searchSection{}
	for _, s := range page.Sections {
		byAnchor[s.Anchor] = s
	}
	top, ok := byAnchor["usage"]
	if !ok {
		t.Fatalf("missing h1 section; have %v", page.Sections)
	}
	if top.Title != "Usage" || !strings.Contains(top.Text, "Install the thing first.") {
		t.Errorf("h1 section wrong: %+v", top)
	}
	cfg, ok := byAnchor["configuration"]
	if !ok {
		t.Fatalf("missing configuration section; have %v", page.Sections)
	}
	if !strings.Contains(cfg.Text, "option = true") {
		t.Errorf("code text not searchable: %q", cfg.Text)
	}
	if strings.Contains(cfg.Text, "<") {
		t.Errorf("tags left in section text: %q", cfg.Text)
	}
	tr, ok := byAnchor["troubleshooting"]
	if !ok {
		t.Fatalf("missing troubleshooting section; have %v", page.Sections)
	}
	if !strings.Contains(tr.Text, "Restart & retry.") {
		t.Errorf("entities not unescaped: %q", tr.Text)
	}
}

func TestBuildSearchIndexLeadingSectionWithoutHeadings(t *testing.T) {
	docs := []Document{
		{RelPath: "notes.md", OutputRel: "notes.html", Title: "Notes"},
	}
	pages := []string{`<p>Body without any heading markup.</p>`}
	index := buildSearchIndex(docs, pages)
	if len(index[0].Sections) != 1 {
		t.Fatalf("want 1 implicit section, got %v", index[0].Sections)
	}
	s := index[0].Sections[0]
	if s.Anchor != "" || s.Title != "Notes" || !strings.Contains(s.Text, "Body without any heading") {
		t.Errorf("implicit section wrong: %+v", s)
	}
}

func TestBuildSearchIndexMermaidNotSearchable(t *testing.T) {
	docs := []Document{
		{RelPath: "diagram.md", OutputRel: "diagram.html", Title: "Diagram"},
	}
	pages := []string{`<h1 id="diagram">Diagram</h1>
<pre class="mermaid">flowchart TD
  A[quokka] --&gt; B[wallaby]
</pre>
<p>Keep it meridian.</p>
`}
	index := buildSearchIndex(docs, pages)
	joined := ""
	for _, s := range index[0].Sections {
		joined += s.Text
	}
	if strings.Contains(joined, "quokka") {
		t.Errorf("mermaid source leaked into search index: %q", joined)
	}
	if !strings.Contains(joined, "meridian") {
		t.Errorf("prose missing from search index: %q", joined)
	}
}

func TestBuildSearchIndexSectionTextTruncated(t *testing.T) {
	docs := []Document{
		{RelPath: "big.md", OutputRel: "big.html", Title: "Big"},
	}
	body := "<p>" + strings.Repeat("word ", 2000) + "</p>"
	pages := []string{`<h1 id="big">Big</h1>` + body}
	index := buildSearchIndex(docs, pages)
	for _, s := range index[0].Sections {
		if len(s.Text) > searchSectionLimit {
			t.Errorf("section text not truncated: %d chars", len(s.Text))
		}
	}
}

func TestEscapeOutputRel(t *testing.T) {
	cases := map[string]string{
		"main.html":               "main.html",
		"parser/parser.html":      "parser/parser.html",
		"deep/nested path/x.html": "deep/nested%20path/x.html",
		`windows\style\y.html`:    "windows/style/y.html",
	}
	for in, want := range cases {
		if got := escapeOutputRel(in); got != want {
			t.Errorf("escapeOutputRel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderSearchIndexJS(t *testing.T) {
	index := []searchPage{
		{URL: "a.html", Title: "A", Sections: []searchSection{{Anchor: "x", Title: "X", Text: "body"}}},
	}
	data, err := renderSearchIndexJS(index)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	const prefix = "window.BARNACLE_SEARCH_INDEX = "
	start := strings.Index(text, prefix)
	if start < 0 {
		t.Fatalf("index assignment missing in %q", text[:80])
	}
	payload := strings.TrimSuffix(strings.TrimSuffix(text[start+len(prefix):], "\n"), ";")
	var parsed []map[string]any
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if len(parsed) != 1 || parsed[0]["u"] != "a.html" {
		t.Errorf("payload round-trip wrong: %v", parsed)
	}
}

func TestEmptyIndexStillEmitsValidJS(t *testing.T) {
	data, err := renderSearchIndexJS(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "window.BARNACLE_SEARCH_INDEX = null") && !strings.Contains(string(data), "window.BARNACLE_SEARCH_INDEX = []") {
		t.Errorf("empty index should serialize to null or []: %s", data)
	}
}
