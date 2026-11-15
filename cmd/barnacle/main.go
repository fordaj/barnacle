// Command barnacle turns Markdown and Mermaid documentation files that
// live alongside source code into a static, interconnected HTML website.
//
// Usage:
//
//	barnacle [--config <path>] [--site <path>]
//
// With no flags, barnacle reads ./.barnacle.json (if it exists) and writes
// the generated site to ./site.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fordaj/barnacle/internal/barnacle"
)

// version is set at build time via -ldflags "-X main.version=<v>".
var version = "dev"

const usageText = `barnacle — build a navigable, diagram-centric documentation website

Usage:
  barnacle [options]

Options:
  --config <path>  Path to a Barnacle configuration file (default: ./.barnacle.json)
  --site <path>    Output directory for the generated website (default: ./site)
  --version        Print the version and exit
  --help           Show this help and exit
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("barnacle", flag.ContinueOnError)
	fs.SetOutput(stderr)

	configPath := fs.String("config", "", "path to a Barnacle configuration file")
	sitePath := fs.String("site", "", "output directory for the generated website")
	showVersion := fs.Bool("version", false, "print the version and exit")
	fs.Usage = func() { fmt.Fprint(stderr, usageText) }

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "barnacle: unexpected argument %q\n\n%s", fs.Arg(0), usageText)
		return 2
	}

	if *showVersion {
		fmt.Fprintf(stdout, "barnacle %s\n", version)
		return 0
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "barnacle: %v\n", err)
		return 1
	}

	explicitConfig := *configPath != ""
	config := filepath.Join(cwd, ".barnacle.json")
	if explicitConfig {
		config = *configPath
		if !filepath.IsAbs(config) {
			config = filepath.Join(cwd, config)
		}
	}

	// Documentation is always scanned from the current directory, which is
	// expected to be the repository root; --config only changes where the
	// configuration is read from.
	root := cwd

	site := filepath.Join(cwd, "site")
	if *sitePath != "" {
		site = *sitePath
	}

	cfg, err := barnacle.LoadConfig(config, explicitConfig, root)
	if err != nil {
		fmt.Fprintf(stderr, "barnacle: %v\n", err)
		return 1
	}

	pages, err := barnacle.Generate(root, site, cfg)
	if err != nil {
		fmt.Fprintf(stderr, "barnacle: %v\n", err)
		return 1
	}

	absSite, err := filepath.Abs(site)
	if err != nil {
		absSite = site
	}
	fmt.Fprintf(stdout, "barnacle: generated %d pages in %s\n", pages, absSite)
	return 0
}
