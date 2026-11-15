package barnacle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMermaidWarningEnabled(t *testing.T) {
	def := Config{}
	if !def.MermaidWarningEnabled() {
		t.Errorf("default config should enable the Mermaid guidance note")
	}
	on := true
	off := false
	if !(Config{WarnMissingMermaid: &on}).MermaidWarningEnabled() {
		t.Errorf("explicit true should enable the note")
	}
	if (Config{WarnMissingMermaid: &off}).MermaidWarningEnabled() {
		t.Errorf("explicit false should silence the note")
	}
}

func TestLoadConfigWarnMissingMermaid(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".barnacle.json")

	write := func(content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write(`{"warnMissingMermaid": false}`)
	cfg, err := LoadConfig(path, true, root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MermaidWarningEnabled() {
		t.Errorf("\"warnMissingMermaid\": false should silence the note")
	}

	write(`{"warnMissingMermaid": true}`)
	cfg, err = LoadConfig(path, true, root)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.MermaidWarningEnabled() {
		t.Errorf("\"warnMissingMermaid\": true should enable the note")
	}

	// Absent field: default stays enabled.
	write(`{}`)
	cfg, err = LoadConfig(path, true, root)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.MermaidWarningEnabled() {
		t.Errorf("missing warnMissingMermaid should default to enabled")
	}

	// Wrong type: invalid config, like every other field.
	write(`{"warnMissingMermaid": "false"}`)
	if _, err := LoadConfig(path, true, root); err == nil {
		t.Errorf("non-boolean warnMissingMermaid should be a config error")
	}
}
