package plist

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportBundleRoundTrip(t *testing.T) {
	pl := &LaunchAgentPlist{
		Label:            "com.test.roundtrip",
		Program:          "/usr/bin/true",
		ProgramArguments: []string{"/usr/bin/true", "--flag"},
		RunAtLoad:        true,
		StartInterval:    300,
		StandardOutPath:  "/tmp/test.stdout.log",
		StandardErrorPath: "/tmp/test.stderr.log",
	}

	bundle := NewExportBundle(pl, "/Users/test/Library/LaunchAgents/com.test.roundtrip.plist", "user", "agent")

	if bundle.Version != 1 {
		t.Errorf("expected version 1, got %d", bundle.Version)
	}
	if bundle.Label != "com.test.roundtrip" {
		t.Errorf("expected label %q, got %q", "com.test.roundtrip", bundle.Label)
	}
	if bundle.Domain != "user" {
		t.Errorf("expected domain %q, got %q", "user", bundle.Domain)
	}
	if bundle.Type != "agent" {
		t.Errorf("expected type %q, got %q", "agent", bundle.Type)
	}
	if bundle.ExportedAt == "" {
		t.Error("expected non-empty ExportedAt")
	}

	// Write to buffer.
	var buf bytes.Buffer
	if err := WriteBundle(&buf, bundle); err != nil {
		t.Fatalf("failed to write bundle: %v", err)
	}

	// Read back.
	restored, err := ReadBundle(&buf)
	if err != nil {
		t.Fatalf("failed to read bundle: %v", err)
	}

	if restored.Version != bundle.Version {
		t.Errorf("version mismatch: %d != %d", restored.Version, bundle.Version)
	}
	if restored.Label != bundle.Label {
		t.Errorf("label mismatch: %q != %q", restored.Label, bundle.Label)
	}
	if restored.Domain != bundle.Domain {
		t.Errorf("domain mismatch: %q != %q", restored.Domain, bundle.Domain)
	}
	if restored.Type != bundle.Type {
		t.Errorf("type mismatch: %q != %q", restored.Type, bundle.Type)
	}
	if restored.Plist.Label != pl.Label {
		t.Errorf("plist label mismatch: %q != %q", restored.Plist.Label, pl.Label)
	}
	if restored.Plist.Program != pl.Program {
		t.Errorf("plist program mismatch: %q != %q", restored.Plist.Program, pl.Program)
	}
	if restored.Plist.StartInterval != pl.StartInterval {
		t.Errorf("plist start interval mismatch: %d != %d", restored.Plist.StartInterval, pl.StartInterval)
	}
	if restored.Plist.RunAtLoad != pl.RunAtLoad {
		t.Errorf("plist run at load mismatch: %v != %v", restored.Plist.RunAtLoad, pl.RunAtLoad)
	}
}

func TestWriteBundleToFileAndReadBack(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test-export.json")

	pl := &LaunchAgentPlist{
		Label:   "com.test.file",
		Program: "/usr/bin/true",
	}

	bundle := NewExportBundle(pl, "/some/path.plist", "global", "daemon")

	if err := WriteBundleToFile(path, bundle); err != nil {
		t.Fatalf("failed to write bundle to file: %v", err)
	}

	// Verify file exists.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("bundle file does not exist: %v", err)
	}

	// Read back.
	restored, err := ReadBundleFromFile(path)
	if err != nil {
		t.Fatalf("failed to read bundle from file: %v", err)
	}

	if restored.Label != "com.test.file" {
		t.Errorf("expected label %q, got %q", "com.test.file", restored.Label)
	}
	if restored.Domain != "global" {
		t.Errorf("expected domain %q, got %q", "global", restored.Domain)
	}
	if restored.Type != "daemon" {
		t.Errorf("expected type %q, got %q", "daemon", restored.Type)
	}
}

func TestReadBundleInvalid(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "empty json",
			input:   "{}",
			wantErr: "missing version",
		},
		{
			name:    "missing label",
			input:   `{"version": 1}`,
			wantErr: "missing label",
		},
		{
			name:    "invalid json",
			input:   "not json at all",
			wantErr: "failed to decode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ReadBundle(strings.NewReader(tt.input))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestReadBundleFromFileNotFound(t *testing.T) {
	_, err := ReadBundleFromFile("/nonexistent/path/bundle.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}
