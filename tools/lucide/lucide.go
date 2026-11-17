// Package lucide embeds the vendored Lucide icon set into the barnacle
// binary, so the icons, their version and their license travel with every
// build of barnacle — every binary is self-contained.
//
// The contents of this directory are refreshed by update.sh (see
// update.sh.md); this file only re-exports them to Go. The embedded set is
// used by barnacle's local-SVG inlining through the `barnacle:<name>.svg`
// shorthand in <img> tags, and is printed by `barnacle list-icons`.
//
// Attribution: the icons are from [Lucide](https://lucide.dev) — see
// https://github.com/lucide-icons/lucide. Lucide is ISC-licensed; the icons
// derived from Feather retain the MIT License. Both texts are embedded here
// (LICENSE) and shipped next to every release binary as LUCIDE-LICENSE.
package lucide

import (
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"
)

//go:embed icons LICENSE VERSION
var FS embed.FS

const (
	// Name is the upstream project name, used in attribution output.
	Name = "Lucide"
	// Homepage is the canonical project URL.
	Homepage = "https://lucide.dev"
	// Source is the upstream repository.
	Source = "https://github.com/lucide-icons/lucide"
	// LicenseName is the license the set ships under (with a Feather-derived
	// MIT-licensed section; see LICENSE).
	LicenseName = "ISC"
)

var nameRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Version returns the vendored release tag (for example "1.73.0").
func Version() (string, error) {
	data, err := FS.ReadFile("VERSION")
	if err != nil {
		return "", fmt.Errorf("reading embedded lucide VERSION: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// License returns the upstream LICENSE text (the ISC text plus the
// Feather-derived MIT section), exactly as vendored by update.sh.
func License() ([]byte, error) {
	data, err := FS.ReadFile("LICENSE")
	if err != nil {
		return nil, fmt.Errorf("reading embedded lucide LICENSE: %w", err)
	}
	return data, nil
}

// IconNames returns the sorted names (without the .svg suffix) of every
// embedded icon. The result is never empty for a correctly vendored bundle.
func IconNames() ([]string, error) {
	entries, err := fs.ReadDir(FS, "icons")
	if err != nil {
		return nil, fmt.Errorf("reading embedded lucide icons: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".svg") {
			continue
		}
		names = append(names, strings.TrimSuffix(name, ".svg"))
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("embedded lucide icon set is empty")
	}
	sort.Strings(names)
	return names, nil
}

// Icon returns the bytes of the embedded icon <name> (without the .svg
// suffix). The name must match the vendored naming scheme exactly (lowercase
// letters, digits and hyphens, for example "arrow-right").
func Icon(name string) ([]byte, error) {
	if !nameRe.MatchString(name) {
		return nil, fmt.Errorf("invalid lucide icon name %q (want lowercase letters, digits and hyphens, e.g. %q)", name, "arrow-right")
	}
	data, err := FS.ReadFile("icons/" + name + ".svg")
	if err != nil {
		return nil, fmt.Errorf("unknown lucide icon %q (run 'barnacle list-icons' to see the embedded set)", name)
	}
	return data, nil
}

// Attribution returns the credit lines `barnacle list-icons` prints in
// addition to the list itself: the project, a link to it, and its license.
func Attribution() string {
	version, err := Version()
	if err != nil {
		version = "unknown"
	}
	return fmt.Sprintf(`The icons above are %s v%s (%s), © Lucide Icons and contributors.
%s is %s-licensed; the Feather-derived icons retain their MIT License.
The full license text is embedded in this binary and is also shipped next
to every release binary (LUCIDE-LICENSE) and in every release archive.
%s · source: %s`, Name, version, Homepage, Name, LicenseName, Homepage, Source)
}
