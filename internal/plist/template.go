package plist

import (
	"fmt"
	"strings"
)

// Template describes a built-in plist template for the "create" command.
type Template struct {
	Name        string
	Description string
	Plist       LaunchAgentPlist
}

// BuiltinTemplates returns the set of built-in plist templates.
func BuiltinTemplates() []Template {
	return []Template{
		{
			Name:        "simple",
			Description: "Run once at login",
			Plist: LaunchAgentPlist{
				RunAtLoad: true,
			},
		},
		{
			Name:        "interval",
			Description: "Run every N seconds",
			Plist: LaunchAgentPlist{
				StartInterval: 300,
			},
		},
		{
			Name:        "calendar",
			Description: "Run on a cron-like schedule",
			Plist: LaunchAgentPlist{
				StartCalendarInterval: map[string]interface{}{
					"Hour":   0,
					"Minute": 0,
				},
			},
		},
		{
			Name:        "keepalive",
			Description: "Always running, restart on crash",
			Plist: LaunchAgentPlist{
				RunAtLoad: true,
				KeepAlive: map[string]interface{}{
					"Crashed": true,
				},
			},
		},
		{
			Name:        "watcher",
			Description: "Run when watched paths change",
			Plist: LaunchAgentPlist{
				WatchPaths: []string{"/tmp/watched"},
			},
		},
		{
			Name:        "monitor-cpu",
			Description: "Monitor CPU usage via pstop (alerts above 80%)",
			Plist: LaunchAgentPlist{
				Program:       "/opt/homebrew/bin/pstop",
				ProgramArguments: []string{"/opt/homebrew/bin/pstop", "watch", "--alert", "--cpu", "80", "--json"},
				StartInterval: 300,
				RunAtLoad:     true,
			},
		},
		{
			Name:        "monitor-ports",
			Description: "Monitor open ports via whport",
			Plist: LaunchAgentPlist{
				Program:       "/opt/homebrew/bin/whport",
				ProgramArguments: []string{"/opt/homebrew/bin/whport", "watch", "--alert", "--json"},
				StartInterval: 600,
				RunAtLoad:     true,
			},
		},
		{
			Name:        "monitor-security",
			Description: "Security audit via macdog",
			Plist: LaunchAgentPlist{
				Program:       "/opt/homebrew/bin/macdog",
				ProgramArguments: []string{"/opt/homebrew/bin/macdog", "audit", "--watch", "--json"},
				StartInterval: 3600,
				RunAtLoad:     true,
			},
		},
		{
			Name:        "monitor-disk",
			Description: "Monitor disk space via macbroom (alert below 10G free)",
			Plist: LaunchAgentPlist{
				Program:       "/opt/homebrew/bin/macbroom",
				ProgramArguments: []string{"/opt/homebrew/bin/macbroom", "watch", "--free", "10G", "--json"},
				StartInterval: 1800,
				RunAtLoad:     true,
			},
		},
	}
}

// GetTemplate returns a template by name, or an error if not found.
func GetTemplate(name string) (*Template, error) {
	for _, t := range BuiltinTemplates() {
		if t.Name == name {
			return &t, nil
		}
	}
	names := make([]string, 0, len(BuiltinTemplates()))
	for _, t := range BuiltinTemplates() {
		names = append(names, t.Name)
	}
	return nil, fmt.Errorf("unknown template %q; available: %s", name, strings.Join(names, ", "))
}

// ApplyDefaults fills in standard log paths for a plist based on its label.
func ApplyDefaults(pl *LaunchAgentPlist) {
	if pl.StandardOutPath == "" {
		pl.StandardOutPath = fmt.Sprintf("/tmp/%s.stdout.log", pl.Label)
	}
	if pl.StandardErrorPath == "" {
		pl.StandardErrorPath = fmt.Sprintf("/tmp/%s.stderr.log", pl.Label)
	}
}
