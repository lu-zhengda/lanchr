package plist

import (
	"strings"
	"testing"
)

func TestBuiltinTemplates(t *testing.T) {
	templates := BuiltinTemplates()

	// Verify we have the expected number of templates.
	if len(templates) != 9 {
		t.Fatalf("expected 9 built-in templates, got %d", len(templates))
	}

	// Verify all templates have a name and description.
	for _, tmpl := range templates {
		if tmpl.Name == "" {
			t.Error("template has empty name")
		}
		if tmpl.Description == "" {
			t.Errorf("template %q has empty description", tmpl.Name)
		}
	}
}

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"simple", false},
		{"interval", false},
		{"calendar", false},
		{"keepalive", false},
		{"watcher", false},
		{"monitor-cpu", false},
		{"monitor-ports", false},
		{"monitor-security", false},
		{"monitor-disk", false},
		{"nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := GetTemplate(tt.name)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for template %q, got nil", tt.name)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for template %q: %v", tt.name, err)
			}
			if tmpl.Name != tt.name {
				t.Errorf("expected template name %q, got %q", tt.name, tmpl.Name)
			}
		})
	}
}

func TestGetTemplateErrorMessage(t *testing.T) {
	_, err := GetTemplate("bogus")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	msg := err.Error()
	for _, name := range []string{"simple", "monitor-cpu", "monitor-disk"} {
		if !strings.Contains(msg, name) {
			t.Errorf("error message should list %q; got: %s", name, msg)
		}
	}
}

func TestMonitorTemplateDefaults(t *testing.T) {
	tests := []struct {
		name        string
		wantProgram string
		wantArgs    []string
	}{
		{
			name:        "monitor-cpu",
			wantProgram: "/opt/homebrew/bin/pstop",
			wantArgs:    []string{"/opt/homebrew/bin/pstop", "watch", "--alert", "--cpu", "80", "--json"},
		},
		{
			name:        "monitor-ports",
			wantProgram: "/opt/homebrew/bin/whport",
			wantArgs:    []string{"/opt/homebrew/bin/whport", "watch", "--alert", "--json"},
		},
		{
			name:        "monitor-security",
			wantProgram: "/opt/homebrew/bin/macdog",
			wantArgs:    []string{"/opt/homebrew/bin/macdog", "audit", "--watch", "--json"},
		},
		{
			name:        "monitor-disk",
			wantProgram: "/opt/homebrew/bin/macbroom",
			wantArgs:    []string{"/opt/homebrew/bin/macbroom", "watch", "--free", "10G", "--json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := GetTemplate(tt.name)
			if err != nil {
				t.Fatalf("failed to get template %q: %v", tt.name, err)
			}

			if tmpl.Plist.Program != tt.wantProgram {
				t.Errorf("expected Program %q, got %q", tt.wantProgram, tmpl.Plist.Program)
			}

			if len(tmpl.Plist.ProgramArguments) != len(tt.wantArgs) {
				t.Fatalf("expected %d ProgramArguments, got %d", len(tt.wantArgs), len(tmpl.Plist.ProgramArguments))
			}
			for i, arg := range tt.wantArgs {
				if tmpl.Plist.ProgramArguments[i] != arg {
					t.Errorf("ProgramArguments[%d]: expected %q, got %q", i, arg, tmpl.Plist.ProgramArguments[i])
				}
			}

			if tmpl.Plist.StartInterval <= 0 {
				t.Error("monitor template should have a positive StartInterval")
			}

			if !tmpl.Plist.RunAtLoad {
				t.Error("monitor template should have RunAtLoad=true")
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	pl := &LaunchAgentPlist{Label: "com.test.example"}
	ApplyDefaults(pl)

	if pl.StandardOutPath != "/tmp/com.test.example.stdout.log" {
		t.Errorf("unexpected stdout path: %s", pl.StandardOutPath)
	}
	if pl.StandardErrorPath != "/tmp/com.test.example.stderr.log" {
		t.Errorf("unexpected stderr path: %s", pl.StandardErrorPath)
	}

	// Should not overwrite existing paths.
	pl2 := &LaunchAgentPlist{
		Label:             "com.test.example",
		StandardOutPath:   "/custom/out.log",
		StandardErrorPath: "/custom/err.log",
	}
	ApplyDefaults(pl2)

	if pl2.StandardOutPath != "/custom/out.log" {
		t.Errorf("should not overwrite existing stdout path: %s", pl2.StandardOutPath)
	}
	if pl2.StandardErrorPath != "/custom/err.log" {
		t.Errorf("should not overwrite existing stderr path: %s", pl2.StandardErrorPath)
	}
}
