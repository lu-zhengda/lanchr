package plist

import "fmt"

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
	}
}

// GetTemplate returns a template by name, or an error if not found.
func GetTemplate(name string) (*Template, error) {
	for _, t := range BuiltinTemplates() {
		if t.Name == name {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("unknown template %q; available: simple, interval, calendar, keepalive, watcher", name)
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
