package barnacle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func sampleProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "main.md", "# Application Entry Point\n\nStarts the app.\n\nSee [Database](database.md) and [Parser](parser/parser.md).\n\n```mermaid\nflowchart TD\n  A --> B\n```\n")
	writeFile(t, root, "database.md", "# Database\n\nStorage layer.\n")
	writeFile(t, root, "parser/parser.md", "# Parser\n\nParses things. Back to [main](../main.md).\n")
	writeFile(t, root, "parser/grammar.mmd", "flowchart LR\n  Tok --> AST\n")
	writeFile(t, root, ".hidden/secret.md", "# Secret\n")
	writeFile(t, root, "node_modules/pkg/dep.md", "# Dep\n")
	return root
}

func TestScanFindsDocsAndSkipsExcluded(t *testing.T) {
	root := sampleProject(t)
	cfg := Config{Title: "Sample"}
	docs, err := Scan(root, filepath.Join(root, "site"), cfg)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, d := range docs {
		got = append(got, d.RelPath)
	}
	want := []string{"database.md", "main.md", "parser/grammar.mmd", "parser/parser.md"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("scanned %v, want %v", got, want)
	}
}

func TestScanExcludeConfig(t *testing.T) {
	root := sampleProject(t)
	cfg := Config{Title: "Sample", Exclude: []string{"parser"}}
	docs, err := Scan(root, filepath.Join(root, "site"), cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range docs {
		if strings.HasPrefix(d.RelPath, "parser/") {
			t.Errorf("excluded directory still scanned: %s", d.RelPath)
		}
	}
}

func TestGenerateSite(t *testing.T) {
	root := sampleProject(t)
	cfg := Config{Title: "Sample"}
	site := filepath.Join(root, "site")
	pages, missing, err := Generate(root, site, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if pages == 0 {
		t.Fatal("no pages generated")
	}
	for _, rel := range missing {
		if rel == "main.md" {
			t.Errorf("main.md contains a Mermaid diagram but was reported missing")
		}
	}
	got := strings.Join(missing, ",")
	wantMissing := "database.md,parser/parser.md"
	if got != wantMissing {
		t.Errorf("missing mermaid pages = %q, want %q", got, wantMissing)
	}

	expected := []string{
		"index.html",
		"main.html",
		"database.html",
		"parser/parser.html",
		"parser/grammar.html",
		"assets/barnacle.css",
		"assets/mermaid.min.js",
		markerFile,
	}
	for _, rel := range expected {
		if _, err := os.Stat(filepath.Join(site, filepath.FromSlash(rel))); err != nil {
			t.Errorf("expected output %s: %v", rel, err)
		}
	}

	read := func(rel string) string {
		data, err := os.ReadFile(filepath.Join(site, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}

	mainHTML := read("main.html")
	if !strings.Contains(mainHTML, `href="database.html"`) {
		t.Errorf("main.html missing rewritten database link")
	}
	if !strings.Contains(mainHTML, `href="parser/parser.html"`) {
		t.Errorf("main.html missing rewritten parser link")
	}
	if !strings.Contains(mainHTML, `<pre class="mermaid">`) {
		t.Errorf("main.html missing mermaid diagram block")
	}
	if !strings.Contains(mainHTML, `mermaid.initialize`) {
		t.Errorf("main.html missing mermaid bootstrap script")
	}

	nested := read("parser/parser.html")
	if !strings.Contains(nested, `href="../main.html"`) {
		t.Errorf("nested page link to main.md not rewritten relatively")
	}
	if !strings.Contains(nested, `href="../assets/barnacle.css"`) {
		t.Errorf("nested page stylesheet path wrong")
	}

	grammar := read("parser/grammar.html")
	if !strings.Contains(grammar, `<pre class="mermaid">flowchart LR`) {
		t.Errorf("grammar.mmd not rendered as mermaid block:\n%s", grammar)
	}

	// The landing page links to every page.
	index := read("index.html")
	for _, rel := range []string{"main.html", "database.html", "parser/parser.html", "parser/grammar.html"} {
		if !strings.Contains(index, fmt.Sprintf(`href="%s"`, rel)) {
			t.Errorf("index.html missing link to %s", rel)
		}
	}
}

func TestGenerateCollision(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "dup.md", "# A\n")
	writeFile(t, root, "dup.mmd", "flowchart TD\n  A --> B\n")
	_, _, err := Generate(root, filepath.Join(root, "site"), Config{Title: "x"})
	if err == nil {
		t.Fatal("expected collision error, got nil")
	}
	if !strings.Contains(err.Error(), "dup.html") {
		t.Errorf("collision error should name the conflicting output: %v", err)
	}
}

func TestGenerateRegeneratesMarkerSite(t *testing.T) {
	root := sampleProject(t)
	cfg := Config{Title: "Sample"}
	site := filepath.Join(root, "site")
	if _, _, err := Generate(root, site, cfg); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(site, "stale.html")
	if err := os.WriteFile(stale, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Generate(root, site, cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("stale generated file should be removed on regeneration")
	}
}
