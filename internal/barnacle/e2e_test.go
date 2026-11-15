package barnacle

// End-to-end tests: barnacle generates a small but representative site, and
// the suite then verifies it the way a browser would in two layers.
//
// Layer 1 (always runs, no browser required): every relative href/src in the
// generated site — including the anchors Mermaid renders for `click`
// directives, rewritten to the mirrored .html pages — resolves to a file that
// exists, and no local .md/.markdown/.mmd target remains. This layer alone
// catches the class of bug where a clickable diagram node points at a
// non-existent .md page.
//
// Layer 2 (Chrome/Chromium present): every generated page is loaded in a
// headless browser, and every unique anchor target found on those pages is
// navigated to as well. The suite requires each page to load cleanly
// (exit 0, no net::ERR on stderr, barnacle's generator marker present) and
// every Mermaid diagram to actually render (data-processed="true", no
// Mermaid error glyphs). To force a specific browser binary, set
// BARNACLE_CHROME to the executable's path. Without a browser, layer 2 is
// skipped — layer 1 is not.

import (
	"bytes"
	"context"
	"html"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

// fixtureRepo builds a repository that mirrors the patterns that matter for
// the generated site: nested pages, relative links in every legal form,
// retained-suffix page names, Mermaid click directives (including callbacks,
// which must not be rewritten), and a standalone .mmd page.
func fixtureRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	writeFile(t, root, "README.md", `# Fixture Site

The landing page of the fixture.

`+"```mermaid\nflowchart LR\n  A[usage] --> B[api]\n  click A href \"guide/usage.md\"\n  click B href \"guide/api.go.md\" \"API docs\"\n```"+`

`+"```mermaid\nstateDiagram-v2\n  [*] --> Idle\n  Idle --> Waiting : send(cmd)\n  Waiting --> Done : ack\n  Waiting --> Error : timeout\n  %% barnacle: link \"send(cmd)\" href \"guide/usage.md\" \"Usage docs\"\n  %% barnacle: link \"ack\" href \"guide/api.go.md\"\n```"+`

Browse [usage](guide/usage.md), [the API](guide/api.go.md), and
[the reference](reference.mmd). External links like
[external](https://example.com/docs/scan.md) and fragment links like
[here](#fixture) stay untouched, as do [mailto links](mailto:a@b.c).
`)

	writeFile(t, root, "guide/usage.md", `# Usage

`+"```mermaid\nflowchart TD\n  R[read] --> W[write]\n  R e1@--> Done\n  click R href \"usage.cpp.md\"\n  click W \"api.go.md\" \"tooltip text\"\n  click R callback \"callback with scan.md\"\n  %% barnacle: link e1 href \"api.go.md\"\n```"+`

See <a href="api.go.md">the API</a> and <a href="../README.md">home</a>.

## Configuration

Fragment targets work: [usage](usage.md#configuration) and [api](api.go.md#config).
`)

	writeFile(t, root, "guide/api.go.md", `# API

`+"```mermaid\nflowchart LR\n  G[usage] --> H[home]\n  click G href \"usage.md\"\n  click H href \"../README.md\"\n```"+`

Back to [usage](usage.md) or [home](../README.md).
`)

	writeFile(t, root, "guide/usage.cpp.md", `# Usage (C++)

`+"```mermaid\nflowchart LR\n  C[plain] --> D[plain]\n```"+`

The C++ variant of [usage](usage.md).
`)

	writeFile(t, root, "reference.mmd", "flowchart LR\n  R1[api] --> R2[usage]\n  click R1 href \"guide/api.go.md\"\n  click R2 href \"guide/usage.md\" _blank\n")

	return root
}

func generateFixture(t *testing.T) (site string, pages int) {
	t.Helper()
	root := fixtureRepo(t)
	site = filepath.Join(root, "site")
	pages, _, err := Generate(root, site, Config{Title: "Fixture Site"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return site, pages
}

var (
	hrefSrcRe           = regexp.MustCompile(`(?:href|src|xlink:href)="([^"]*)"`)
	mermaidRe           = regexp.MustCompile(`(?s)<pre class="mermaid">(.*?)</pre>`)
	mermaidHrefRe       = regexp.MustCompile(`(?i)href\s+(?:"([^"\s]*)"|'([^'\s]*)'|([^\s"']+))`)
	mermaidShortRe      = regexp.MustCompile(`(?i)^\s*click\s+([^"\s]+(?:\s+[^"\s]+)*?)\s+"([^"\s]*)"`)
	edgeLinksManifestRe = regexp.MustCompile(`(?s)<script type="application/json" class="barnacle-edge-links">.*?</script>`)
	edgeLinksKeyRe      = regexp.MustCompile(`"key"\s*:`)
	schemeRe2           = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*:`)
	externalRel         = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9+.-]*:|//|#)`)
	docExtRe            = regexp.MustCompile(`(?i)\.(md|markdown|mmd)$`)
)

// hasDocExt reports whether dest's path (before any query/fragment)
// ends in a documentation extension.
func hasDocExt(dest string) bool {
	path := dest
	if i := strings.IndexByte(path, '#'); i >= 0 {
		path = path[:i]
	}
	if i := strings.IndexByte(path, '?'); i >= 0 {
		path = path[:i]
	}
	return docExtRe.MatchString(path)
}

// htmlFiles walks the site directory and returns every .html file.
func htmlFiles(t *testing.T, site string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(site, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("fixture site contains no .html files")
	}
	return files
}

// resolveRef resolves a relative reference from the page it appeared in and
// returns the absolute target path. Reports false for external or same-page
// references.
func resolveRef(pageDir, ref string) (string, bool) {
	if ref == "" || strings.HasPrefix(ref, "#") || strings.HasPrefix(ref, "//") || schemeRe2.MatchString(ref) {
		return "", false
	}
	path := ref
	if i := strings.IndexByte(path, '#'); i >= 0 {
		path = path[:i]
	}
	if i := strings.IndexByte(path, '?'); i >= 0 {
		path = path[:i]
	}
	if path == "" {
		return "", false
	}
	unescaped, err := url.PathUnescape(path)
	if err != nil {
		unescaped = path
	}
	return filepath.Join(pageDir, filepath.FromSlash(unescaped)), true
}

func TestE2EGeneratedSiteAnchorsResolve(t *testing.T) {
	site, pages := generateFixture(t)
	if pages < 6 {
		t.Errorf("expected at least 6 pages (index + 5 docs), got %d", pages)
	}

	files := htmlFiles(t, site)
	for _, page := range files {
		data, err := os.ReadFile(page)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		pageDir := filepath.Dir(page)

		for _, m := range hrefSrcRe.FindAllStringSubmatch(text, -1) {
			dest := m[1]
			if externalRel.MatchString(dest) {
				continue
			}
			// A local .md/.mmd target in the generated output is the
			// bug class this suite exists to prevent: the file does not
			// exist, so clicking fails in a real browser.
			if hasDocExt(dest) {
				t.Errorf("%s: unrewritten documentation target %q", page, dest)
				continue
			}
			target, ok := resolveRef(pageDir, dest)
			if !ok {
				continue
			}
			if _, err := os.Stat(target); err != nil {
				t.Errorf("%s: link to %q resolves to missing file %s", page, dest, target)
			}
		}

		// Mermaid blocks: no .md/.mmd click or barnacle-directive targets
		// may remain either. Block bodies are HTML-escaped in the output,
		// so unescape first.
		for _, m := range mermaidRe.FindAllStringSubmatch(text, -1) {
			body := html.UnescapeString(m[1])
			for _, line := range strings.Split(body, "\n") {
				trimmed := strings.TrimLeft(line, " \t")
				lower := strings.ToLower(trimmed)
				isClick := strings.HasPrefix(lower, "click") && (len(lower) == 5 || lower[5] == ' ' || lower[5] == '\t')
				isDirective := strings.HasPrefix(lower, "%%") && strings.Contains(lower, "barnacle:")
				if !isClick && !isDirective {
					continue
				}
				for _, mm := range mermaidHrefRe.FindAllStringSubmatch(line, -1) {
					for _, g := range mm[1:] {
						if g != "" && hasDocExt(g) {
							t.Errorf("%s: mermaid click keeps an unrewritten documentation target %q\n%s", page, g, line)
						}
					}
				}
				if mm := mermaidShortRe.FindStringSubmatch(line); mm != nil {
					targets, dest := mm[1], mm[2]
					if !containsWord(targets, "callback") && !containsWord(targets, "call") && !containsWord(targets, "href") && hasDocExt(dest) {
						t.Errorf("%s: mermaid click keeps an unrewritten documentation target %q\n%s", page, dest, line)
					}
				}
			}
		}
	}
}

// findChrome locates a Chrome or Chromium executable. BARNACLE_CHROME wins
// over everything; otherwise the platform's usual locations are tried.
func findChrome(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("BARNACLE_CHROME"); p != "" {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
		t.Fatalf("BARNACLE_CHROME is set but not an executable file: %s", p)
	}
	var candidates []string
	switch runtime.GOOS {
	case "darwin":
		candidates = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		}
	case "windows":
		for _, env := range []string{"PROGRAMFILES", "PROGRAMFILES(X86)", "LOCALAPPDATA"} {
			base := os.Getenv(env)
			if base == "" {
				continue
			}
			candidates = append(candidates,
				filepath.Join(base, `Google\Chrome\Application\chrome.exe`),
				filepath.Join(base, `Chromium\Application\chrome.exe`),
				filepath.Join(base, `Microsoft\Edge\Application\msedge.exe`),
			)
		}
	default:
		for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "microsoft-edge"} {
			if p, err := exec.LookPath(name); err == nil {
				return p
			}
		}
	}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c
		}
	}
	if p, err := exec.LookPath("chrome"); err == nil {
		return p
	}
	return ""
}

// dumpDOM loads page in headless Chrome and returns the resulting DOM and
// any stderr output. net::ERR_* failures appear only on stderr; headless
// Chrome exits 0 even for error pages, so stderr is what must be checked.
func dumpDOM(t *testing.T, chrome, pageURL string) (dom, stderr string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, chrome,
		"--headless", "--disable-gpu", "--no-sandbox", "--disable-dev-shm-usage",
		"--virtual-time-budget=3000", "--dump-dom", pageURL)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("chrome timed out loading %s", pageURL)
	}
	if err != nil {
		t.Fatalf("chrome failed loading %s: %v", pageURL, err)
	}
	return outBuf.String(), errBuf.String()
}

func fileURL(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.ToSlash(abs)
	if runtime.GOOS == "windows" {
		return "file:///" + p
	}
	return "file://" + p
}

func TestE2EBrowserNavigation(t *testing.T) {
	chrome := findChrome(t)
	if chrome == "" {
		t.Skip("no Chrome/Chromium found on this machine; browser layer skipped (set BARNACLE_CHROME to the executable)")
	}

	site, _ := generateFixture(t)
	files := htmlFiles(t, site)

	// Load every page; collect unique anchor targets to navigate to.
	targets := map[string]string{}
	for _, page := range files {
		dom, stderr := dumpDOM(t, chrome, fileURL(t, page))
		assertPageOK(t, page, dom, stderr)
		for _, m := range hrefSrcRe.FindAllStringSubmatch(dom, -1) {
			dest := m[1]
			if externalRel.MatchString(dest) {
				continue
			}
			target, ok := resolveRef(filepath.Dir(page), dest)
			if !ok {
				continue
			}
			// The rendered SVG anchors carry the click targets the
			// browser will actually navigate to; a documentation
			// extension here means the pinned binary (or a bug) left the
			// target unrewritten.
			if hasDocExt(dest) {
				t.Errorf("%s: unrewritten documentation target %q", page, dest)
				continue
			}
			if strings.HasSuffix(target, ".html") {
				targets[target] = page
			}
		}
		// After mermaid renders, every edge in a page's manifests must be
		// wired into a clickable SVG anchor. A regression in the injection
		// would leave this count short.
		manifests := edgeLinksManifestRe.FindAllString(dom, -1)
		expectedEdges := 0
		for _, m := range manifests {
			expectedEdges += len(edgeLinksKeyRe.FindAllStringIndex(m, -1))
		}
		if expectedEdges > 0 {
			anchors := strings.Count(dom, `class="barnacle-edge-link"`)
			if anchors != expectedEdges {
				t.Errorf("%s: %d edge anchors injected for %d manifest edges", page, anchors, expectedEdges)
			}
		}

		if strings.Contains(dom, `<pre class="mermaid"`) {
			if !strings.Contains(dom, `data-processed="true"`) {
				t.Errorf("%s: mermaid diagram did not render (data-processed missing)", page)
			}
		}
	}

	// Navigating every anchor target must land on a rendering barnacle page —
	// this is the click-through path the user-facing regression came from.
	for target, from := range targets {
		_ = from
		abs, err := filepath.Abs(target)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(abs, site) {
			continue
		}
		dom, stderr := dumpDOM(t, chrome, fileURL(t, target))
		assertPageOK(t, target, dom, stderr)
	}
	t.Logf("navigated %d pages and %d anchor targets via %s", len(files), len(targets), chrome)
}

// assertPageOK fails the test when the page did not load cleanly in the
// browser. net::ERR_* failures appear only on stderr.
func assertPageOK(t *testing.T, page, dom, stderr string) {
	t.Helper()
	if i := strings.Index(stderr, "net::ERR"); i >= 0 {
		end := strings.IndexByte(stderr[i:], '\n')
		if end < 0 {
			end = len(stderr[i:])
		}
		t.Errorf("%s: page failed to load in browser: %s", page, strings.TrimSpace(stderr[i:i+end]))
		return
	}
	if !strings.Contains(dom, `<meta name="generator" content="barnacle">`) {
		t.Errorf("%s: page rendered but is not a barnacle page (generator marker missing; bytes=%d)", page, len(dom))
	}
	if strings.Contains(dom, `<path class="error-icon"`) || strings.Contains(dom, `<text class="error-text"`) || strings.Contains(dom, "Syntax error in text") {
		t.Errorf("%s: mermaid rendered an error graphic", page)
	}
}
