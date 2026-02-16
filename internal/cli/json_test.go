package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/lu-zhengda/lanchr/internal/agent"
	"github.com/lu-zhengda/lanchr/internal/platform"
)

func TestToJSONServices(t *testing.T) {
	services := []agent.Service{
		{
			Label:          "com.example.test",
			Domain:         platform.DomainUser,
			Type:           platform.TypeAgent,
			Status:         agent.StatusRunning,
			PID:            1234,
			LastExitStatus: 0,
			PlistPath:      "/Users/test/Library/LaunchAgents/com.example.test.plist",
			Program:        "/usr/local/bin/test",
		},
		{
			Label:          "com.example.stopped",
			Domain:         platform.DomainGlobal,
			Type:           platform.TypeDaemon,
			Status:         agent.StatusStopped,
			PID:            -1,
			LastExitStatus: 1,
			PlistPath:      "",
			ProgramArgs:    []string{"/usr/bin/daemon", "--flag"},
		},
	}

	got := toJSONServices(services)

	if len(got) != 2 {
		t.Fatalf("got %d services, want 2", len(got))
	}

	// First service.
	if got[0].Label != "com.example.test" {
		t.Errorf("got label %q, want %q", got[0].Label, "com.example.test")
	}
	if got[0].Domain != "user" {
		t.Errorf("got domain %q, want %q", got[0].Domain, "user")
	}
	if got[0].Type != "agent" {
		t.Errorf("got type %q, want %q", got[0].Type, "agent")
	}
	if got[0].Status != "running" {
		t.Errorf("got status %q, want %q", got[0].Status, "running")
	}
	if got[0].PID != 1234 {
		t.Errorf("got PID %d, want %d", got[0].PID, 1234)
	}
	if got[0].Program != "/usr/local/bin/test" {
		t.Errorf("got program %q, want %q", got[0].Program, "/usr/local/bin/test")
	}

	// Second service: program from ProgramArgs.
	if got[1].Program != "/usr/bin/daemon" {
		t.Errorf("got program %q, want %q", got[1].Program, "/usr/bin/daemon")
	}
	if got[1].PlistPath != "" {
		t.Errorf("got plist_path %q, want empty", got[1].PlistPath)
	}

	// Verify JSON round-trip.
	var buf bytes.Buffer
	if err := fprintJSON(&buf, got); err != nil {
		t.Fatalf("fprintJSON() error = %v", err)
	}
	var parsed []jsonService
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(parsed) != 2 {
		t.Fatalf("round-trip: got %d services, want 2", len(parsed))
	}
	if parsed[0].Label != "com.example.test" {
		t.Errorf("round-trip: got label %q, want %q", parsed[0].Label, "com.example.test")
	}
}

func TestToJSONServices_Empty(t *testing.T) {
	got := toJSONServices(nil)
	if len(got) != 0 {
		t.Errorf("got %d services for nil input, want 0", len(got))
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, got); err != nil {
		t.Fatalf("fprintJSON() error = %v", err)
	}
	if got := buf.String(); got != "[]\n" {
		t.Errorf("got %q, want %q", got, "[]\n")
	}
}

func TestToJSONServiceDetail(t *testing.T) {
	svc := &agent.Service{
		Label:             "com.example.detailed",
		Domain:            platform.DomainUser,
		Type:              platform.TypeAgent,
		Status:            agent.StatusRunning,
		PID:               5678,
		LastExitStatus:    0,
		PlistPath:         "/path/to/plist",
		Program:           "/usr/bin/prog",
		ProgramArgs:       []string{"/usr/bin/prog", "--verbose"},
		RunAtLoad:         true,
		KeepAlive:         true,
		StartInterval:     300,
		WatchPaths:        []string{"/tmp/watch"},
		WorkingDirectory:  "/var/tmp",
		StandardOutPath:   "/var/log/stdout.log",
		StandardErrorPath: "/var/log/stderr.log",
		EnvironmentVars:   map[string]string{"HOME": "/root"},
		ExitTimeout:       30,
		Disabled:          false,
		BlameLine:         "blame-info",
	}

	got := toJSONServiceDetail(svc)

	if got.Label != "com.example.detailed" {
		t.Errorf("got label %q, want %q", got.Label, "com.example.detailed")
	}
	if got.PID != 5678 {
		t.Errorf("got PID %d, want %d", got.PID, 5678)
	}
	if !got.RunAtLoad {
		t.Error("got RunAtLoad=false, want true")
	}
	if got.StartInterval != 300 {
		t.Errorf("got StartInterval %d, want %d", got.StartInterval, 300)
	}
	if len(got.ProgramArgs) != 2 {
		t.Errorf("got %d program args, want 2", len(got.ProgramArgs))
	}
	if len(got.WatchPaths) != 1 || got.WatchPaths[0] != "/tmp/watch" {
		t.Errorf("got watch paths %v, want [/tmp/watch]", got.WatchPaths)
	}
	if got.EnvironmentVars["HOME"] != "/root" {
		t.Errorf("got env HOME=%q, want %q", got.EnvironmentVars["HOME"], "/root")
	}
	if got.BlameLine != "blame-info" {
		t.Errorf("got blame %q, want %q", got.BlameLine, "blame-info")
	}

	// Verify JSON round-trip.
	var buf bytes.Buffer
	if err := fprintJSON(&buf, got); err != nil {
		t.Fatalf("fprintJSON() error = %v", err)
	}
	var parsed jsonServiceDetail
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Label != "com.example.detailed" {
		t.Errorf("round-trip: got label %q, want %q", parsed.Label, "com.example.detailed")
	}
	if parsed.ExitTimeout != 30 {
		t.Errorf("round-trip: got exit_timeout %d, want %d", parsed.ExitTimeout, 30)
	}
}

func TestToJSONDoctor(t *testing.T) {
	t.Run("with findings", func(t *testing.T) {
		findings := []agent.Finding{
			{
				Severity:   agent.SeverityCritical,
				Label:      "com.example.broken",
				PlistPath:  "/path/to/broken.plist",
				Message:    "binary not found",
				Suggestion: "fix the path",
			},
			{
				Severity:   agent.SeverityWarning,
				Label:      "com.example.warn",
				PlistPath:  "/path/to/warn.plist",
				Message:    "world-writable plist",
				Suggestion: "chmod 644",
			},
		}

		got := toJSONDoctor(findings)

		if len(got.Findings) != 2 {
			t.Fatalf("got %d findings, want 2", len(got.Findings))
		}
		if got.Findings[0].Severity != "CRITICAL" {
			t.Errorf("got severity %q, want %q", got.Findings[0].Severity, "CRITICAL")
		}
		if got.Findings[1].Severity != "WARNING" {
			t.Errorf("got severity %q, want %q", got.Findings[1].Severity, "WARNING")
		}
		if got.Summary.Critical != 1 {
			t.Errorf("got critical %d, want 1", got.Summary.Critical)
		}
		if got.Summary.Warning != 1 {
			t.Errorf("got warning %d, want 1", got.Summary.Warning)
		}

		// Verify JSON round-trip.
		var buf bytes.Buffer
		if err := fprintJSON(&buf, got); err != nil {
			t.Fatalf("fprintJSON() error = %v", err)
		}
		var parsed jsonDoctor
		if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if parsed.Findings[0].Label != "com.example.broken" {
			t.Errorf("round-trip: got label %q, want %q", parsed.Findings[0].Label, "com.example.broken")
		}
	})

	t.Run("no findings", func(t *testing.T) {
		got := toJSONDoctor(nil)

		if len(got.Findings) != 0 {
			t.Errorf("got %d findings, want 0", len(got.Findings))
		}
		if got.Summary.Critical != 0 || got.Summary.Warning != 0 {
			t.Errorf("got summary %+v, want all zeros", got.Summary)
		}

		// Verify JSON output contains empty array, not null.
		var buf bytes.Buffer
		if err := fprintJSON(&buf, got); err != nil {
			t.Fatalf("fprintJSON() error = %v", err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if string(raw["findings"]) != "[]" {
			t.Errorf("got findings %s, want []", string(raw["findings"]))
		}
	})
}

func TestJSONAction_RoundTrip(t *testing.T) {
	actions := []string{"enable", "disable", "restart", "load", "unload"}
	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			input := jsonAction{OK: true, Action: action, Label: "com.example.test"}

			var buf bytes.Buffer
			if err := fprintJSON(&buf, input); err != nil {
				t.Fatalf("fprintJSON() error = %v", err)
			}

			var got jsonAction
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
			if !got.OK {
				t.Error("got ok=false, want true")
			}
			if got.Action != action {
				t.Errorf("got action %q, want %q", got.Action, action)
			}
			if got.Label != "com.example.test" {
				t.Errorf("got label %q, want %q", got.Label, "com.example.test")
			}
		})
	}
}

func TestJSONCreate_RoundTrip(t *testing.T) {
	input := jsonCreate{
		OK:        true,
		Action:    "create",
		Label:     "com.example.new",
		PlistPath: "/path/to/new.plist",
		Loaded:    true,
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, input); err != nil {
		t.Fatalf("fprintJSON() error = %v", err)
	}

	var got jsonCreate
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if !got.OK || got.Action != "create" || got.Label != "com.example.new" {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if got.PlistPath != "/path/to/new.plist" {
		t.Errorf("got plist_path %q, want %q", got.PlistPath, "/path/to/new.plist")
	}
	if !got.Loaded {
		t.Error("got loaded=false, want true")
	}
}

func TestJSONImport_RoundTrip(t *testing.T) {
	input := jsonImport{
		OK:        true,
		Action:    "import",
		Label:     "com.example.imported",
		PlistPath: "/path/to/imported.plist",
		Loaded:    false,
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, input); err != nil {
		t.Fatalf("fprintJSON() error = %v", err)
	}

	var got jsonImport
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if got.Action != "import" || got.Label != "com.example.imported" {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if got.Loaded {
		t.Error("got loaded=true, want false")
	}
}

func TestJSONEdit_RoundTrip(t *testing.T) {
	input := jsonEdit{
		OK:           true,
		Action:       "edit",
		Label:        "com.example.edited",
		PlistPath:    "/path/to/edited.plist",
		ValidationOK: true,
		Reloaded:     false,
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, input); err != nil {
		t.Fatalf("fprintJSON() error = %v", err)
	}

	var got jsonEdit
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if got.Action != "edit" || !got.ValidationOK {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if got.Reloaded {
		t.Error("got reloaded=true, want false")
	}
}

func TestJSONLogs_RoundTrip(t *testing.T) {
	input := jsonLogs{
		Label:  "com.example.test",
		Source: "stdout",
		Lines:  []string{"line 1", "line 2", "line 3"},
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, input); err != nil {
		t.Fatalf("fprintJSON() error = %v", err)
	}

	var got jsonLogs
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if got.Label != "com.example.test" {
		t.Errorf("got label %q, want %q", got.Label, "com.example.test")
	}
	if got.Source != "stdout" {
		t.Errorf("got source %q, want %q", got.Source, "stdout")
	}
	if len(got.Lines) != 3 {
		t.Errorf("got %d lines, want 3", len(got.Lines))
	}
}

func TestJSONServiceDetail_OmitsEmpty(t *testing.T) {
	svc := &agent.Service{
		Label:  "com.example.minimal",
		Domain: platform.DomainUser,
		Type:   platform.TypeAgent,
		Status: agent.StatusStopped,
	}

	detail := toJSONServiceDetail(svc)

	var buf bytes.Buffer
	if err := fprintJSON(&buf, detail); err != nil {
		t.Fatalf("fprintJSON() error = %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// These fields should be omitted when empty.
	omittedFields := []string{
		"plist_path", "program", "program_args", "watch_paths",
		"queue_directories", "working_directory", "stdout_path",
		"stderr_path", "environment", "blame",
	}
	for _, field := range omittedFields {
		if _, ok := raw[field]; ok {
			t.Errorf("field %q should be omitted when empty, got %s", field, string(raw[field]))
		}
	}

	// These fields should always be present.
	requiredFields := []string{"label", "domain", "type", "status", "pid", "last_exit_status", "run_at_load", "disabled"}
	for _, field := range requiredFields {
		if _, ok := raw[field]; !ok {
			t.Errorf("field %q should always be present", field)
		}
	}
}
