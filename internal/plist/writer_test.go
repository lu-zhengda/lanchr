package plist

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteWithoutValidation(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.plist")

	w := NewWriter()

	// Should fail without a label.
	err := w.WriteWithoutValidation(&LaunchAgentPlist{}, path)
	if err == nil {
		t.Fatal("expected error for empty label, got nil")
	}

	// Should succeed with a label and a non-existent binary path.
	pl := &LaunchAgentPlist{
		Label:   "com.test.writer",
		Program: "/nonexistent/binary/path",
	}
	if err := w.WriteWithoutValidation(pl, path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("plist file was not created: %v", err)
	}

	// Parse it back to verify round-trip.
	parser := NewParser()
	parsed, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("failed to parse written plist: %v", err)
	}

	if parsed.Label != "com.test.writer" {
		t.Errorf("expected label %q, got %q", "com.test.writer", parsed.Label)
	}
	if parsed.Program != "/nonexistent/binary/path" {
		t.Errorf("expected program %q, got %q", "/nonexistent/binary/path", parsed.Program)
	}
}

func TestValidate(t *testing.T) {
	w := NewWriter()

	// No label, no program.
	errs := w.Validate(&LaunchAgentPlist{})
	if len(errs) < 2 {
		t.Errorf("expected at least 2 validation errors, got %d", len(errs))
	}

	// Valid plist with real binary.
	errs = w.Validate(&LaunchAgentPlist{
		Label:   "com.test.valid",
		Program: "/usr/bin/true",
	})
	if len(errs) != 0 {
		t.Errorf("expected 0 validation errors, got %d: %v", len(errs), errs)
	}

	// Valid label but nonexistent binary.
	errs = w.Validate(&LaunchAgentPlist{
		Label:   "com.test.missing",
		Program: "/nonexistent/binary",
	})
	found := false
	for _, e := range errs {
		if e.Field == "Program" {
			found = true
		}
	}
	if !found {
		t.Error("expected a Program validation error for nonexistent binary")
	}
}
