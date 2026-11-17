package barnacle

import (
	"bytes"
	"sort"
	"strings"
	"testing"

	"github.com/fordaj/barnacle/tools/lucide"
)

func TestListIconsPrintsSortedSetWithAttribution(t *testing.T) {
	var buf bytes.Buffer
	if err := ListIcons(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"icons from Lucide v",
		"https://lucide.dev",
		"ISC-licensed",
		"MIT License",
		"LUCIDE-LICENSE",
		"folder",
		"file-code",
		"zoom-out",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("list-icons output missing %q\n%.400s", want, out)
		}
	}
}

func TestEmbeddedLucideSetIsConsistent(t *testing.T) {
	names, err := lucide.IconNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) < 1000 {
		t.Errorf("embedded set suspiciously small: %d icons", len(names))
	}
	if !sort.StringsAreSorted(names) {
		t.Error("IconNames() is not sorted")
	}
	for _, name := range []string{"folder", "file-code", "hexagon"} {
		found := false
		for _, n := range names {
			if n == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("embedded set is missing well-known icon %q", name)
		}
	}
	if _, err := lucide.Version(); err != nil {
		t.Errorf("embedded VERSION unreadable: %v", err)
	}
	license, err := lucide.License()
	if err != nil {
		t.Fatalf("embedded LICENSE unreadable: %v", err)
	}
	for _, want := range []string{"ISC License", "MIT", "Lucide Icons and Contributors"} {
		if !bytes.Contains(license, []byte(want)) {
			t.Errorf("embedded LICENSE missing %q", want)
		}
	}
}
