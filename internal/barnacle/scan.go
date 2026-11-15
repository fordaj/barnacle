package barnacle

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Document is a single documentation source file discovered in the
// repository.
type Document struct {
	// RelPath is the slash-separated path relative to the root, for
	// example "parser/parser.md".
	RelPath string
	// OutputRel is the slash-separated path of the generated HTML page
	// relative to the site root, for example "parser/parser.html".
	OutputRel string
	// SourcePath is the absolute path of the documentation file.
	SourcePath string
	// Title is the human-readable title of the page.
	Title string
	// MermaidOnly is true for .mmd files, which contain nothing but a
	// diagram.
	MermaidOnly bool
}

// Scan walks root and returns every Markdown (.md) and Mermaid (.mmd)
// documentation file, sorted by relative path. siteDir is the absolute
// path of the output directory and is skipped during the walk.
func Scan(root, siteDir string, cfg Config) ([]Document, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	siteAbs, err := filepath.Abs(siteDir)
	if err != nil {
		return nil, err
	}

	var docs []Document
	err = filepath.WalkDir(rootAbs, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)

		// Never scan the output directory.
		if siteAbs != "" {
			if pathAbs, err := filepath.Abs(path); err == nil && pathAbs == siteAbs {
				return filepath.SkipDir
			}
		}

		if entry.IsDir() {
			name := entry.Name()
			if rel == "." {
				return nil
			}
			if shouldExclude(name, relSlash, cfg.Exclude) {
				return filepath.SkipDir
			}
			if !cfg.IncludeHidden && strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if !entry.Type().IsRegular() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".md" && ext != ".mmd" {
			return nil
		}
		if shouldExclude(entry.Name(), relSlash, cfg.Exclude) {
			return nil
		}
		if !cfg.IncludeHidden && hasHiddenSegment(relSlash) {
			return nil
		}

		doc := Document{
			RelPath:    relSlash,
			SourcePath: path,
			MermaidOnly: ext == ".mmd",
		}
		doc.OutputRel = strings.TrimSuffix(relSlash, filepath.Ext(relSlash)) + ".html"

		title, err := extractTitle(path, ext, entry.Name())
		if err != nil {
			return err
		}
		doc.Title = title
		docs = append(docs, doc)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(docs, func(i, j int) bool { return docs[i].RelPath < docs[j].RelPath })
	return docs, nil
}

func shouldExclude(name, relSlash string, excludes []string) bool {
	for _, entry := range defaultExcludes {
		if name == entry {
			return true
		}
	}
	for _, entry := range excludes {
		entry = filepath.ToSlash(strings.TrimSpace(entry))
		entry = strings.TrimPrefix(entry, "./")
		entry = strings.TrimSuffix(entry, "/")
		if entry == "" {
			continue
		}
		if !strings.Contains(entry, "/") {
			if name == entry {
				return true
			}
			continue
		}
		// Glob against the relative path; also skip the directory itself
		// and everything below it.
		if ok, _ := filepath.Match(entry, relSlash); ok {
			return true
		}
		if strings.HasPrefix(relSlash, entry+"/") {
			return true
		}
	}
	return false
}

func hasHiddenSegment(relSlash string) bool {
	for _, segment := range strings.Split(relSlash, "/") {
		if strings.HasPrefix(segment, ".") {
			return true
		}
	}
	return false
}

// extractTitle reads the first ATX heading from a Markdown file, or falls
// back to the file name without extension.
func extractTitle(path, ext, name string) (string, error) {
	if ext == ".mmd" {
		return humanizeName(name), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		level := 0
		for level < len(trimmed) && trimmed[level] == '#' {
			level++
		}
		if level != 1 {
			continue
		}
		title := strings.TrimSpace(strings.TrimRight(strings.TrimSpace(trimmed[level:]), "#"))
		if title != "" {
			return title, nil
		}
	}
	return humanizeName(name), nil
}

func humanizeName(name string) string {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.Join(strings.Fields(base), " ")
	if base == "" {
		return name
	}
	return strings.ToUpper(base[:1]) + base[1:]
}
