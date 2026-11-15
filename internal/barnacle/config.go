// Package barnacle turns Markdown and Mermaid documentation files that
// live alongside source code into a static, interconnected HTML website.
package barnacle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config is the repository-level configuration read from .barnacle.json.
type Config struct {
	// Title is the name of the site shown in headers and the page titles.
	// Defaults to the base name of the repository root.
	Title string `json:"title"`

	// Exclude lists directories or files to skip while scanning. An entry
	// without a slash is matched against the base name of every path on the
	// walk; an entry with a slash is matched as a glob against the
	// slash-separated path relative to the root (for example "docs/archive/**"
	// is not supported, use "docs/archive" to skip a directory).
	Exclude []string `json:"exclude"`

	// IncludeHidden allows scanning dot-directories and dot-files other
	// than .git, which is always excluded.
	IncludeHidden bool `json:"includeHidden"`
}

// defaultExcludes are directory and file base names always skipped.
var defaultExcludes = []string{
	".git",
	"node_modules",
	"vendor",
	"dist",
}

// LoadConfig reads and validates the configuration file at path. A missing
// file is not an error when using the default location: defaults are
// returned instead. An explicitly requested config that is missing or
// invalid is an error.
func LoadConfig(path string, explicit bool, root string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && !explicit {
			return defaultConfig(root), nil
		}
		return Config{}, fmt.Errorf("reading config %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config %s: %w", path, err)
	}
	if strings.TrimSpace(cfg.Title) == "" {
		cfg.Title = defaultTitle(root)
	}
	for _, entry := range cfg.Exclude {
		if strings.TrimSpace(entry) == "" {
			return Config{}, fmt.Errorf("config %s: exclude entries must not be empty", path)
		}
	}
	return cfg, nil
}

func defaultConfig(root string) Config {
	return Config{Title: defaultTitle(root)}
}

func defaultTitle(root string) string {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	base := filepath.Base(abs)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "Documentation"
	}
	return base
}
