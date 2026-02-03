package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfigInputJSONInline(t *testing.T) {
	parsed, err := parseConfigInput(`{"image":"nginx:latest","env":{"FOO":"bar"}}`)
	if err != nil {
		t.Fatalf("expected JSON to parse: %v", err)
	}
	if parsed["image"] != "nginx:latest" {
		t.Fatalf("expected image to be parsed, got %v", parsed["image"])
	}
	env, ok := parsed["env"].(map[string]interface{})
	if !ok || env["FOO"] != "bar" {
		t.Fatalf("expected env to be parsed, got %v", parsed["env"])
	}
}

func TestParseConfigInputYAMLInline(t *testing.T) {
	parsed, err := parseConfigInput("image: nginx:latest\nenv:\n  FOO: bar\nports:\n  - container_port: 8080\n")
	if err != nil {
		t.Fatalf("expected YAML to parse: %v", err)
	}
	if parsed["image"] != "nginx:latest" {
		t.Fatalf("expected image to be parsed, got %v", parsed["image"])
	}
}

func TestParseConfigInputFileURI(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "config.yaml")
	content := []byte("image: nginx:latest\nenv:\n  FOO: bar\n")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	parsed, err := parseConfigInput("file://" + filePath)
	if err != nil {
		t.Fatalf("expected file URI to parse: %v", err)
	}
	if parsed["image"] != "nginx:latest" {
		t.Fatalf("expected image to be parsed, got %v", parsed["image"])
	}
}

func TestParseConfigInputFilePathJSON(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "config.json")
	content := []byte(`{"image":"nginx:latest","env":{"FOO":"bar"}}`)
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	parsed, err := parseConfigInput(filePath)
	if err != nil {
		t.Fatalf("expected file path to parse: %v", err)
	}
	if parsed["image"] != "nginx:latest" {
		t.Fatalf("expected image to be parsed, got %v", parsed["image"])
	}
}
