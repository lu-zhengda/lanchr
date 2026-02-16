package cli

import "github.com/lu-zhengda/lanchr/internal/agent"

// ---------------------------------------------------------------------------
// Service JSON types (list, search, info)
// ---------------------------------------------------------------------------

// jsonService is the compact representation used by list and search.
type jsonService struct {
	Label          string `json:"label"`
	Domain         string `json:"domain"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	PID            int    `json:"pid"`
	LastExitStatus int    `json:"last_exit_status"`
	PlistPath      string `json:"plist_path,omitempty"`
	Program        string `json:"program,omitempty"`
}

// toJSONServices converts a slice of agent.Service to JSON-serializable form.
func toJSONServices(services []agent.Service) []jsonService {
	out := make([]jsonService, 0, len(services))
	for _, svc := range services {
		out = append(out, jsonService{
			Label:          svc.Label,
			Domain:         svc.Domain.String(),
			Type:           svc.Type.String(),
			Status:         svc.Status.String(),
			PID:            svc.PID,
			LastExitStatus: svc.LastExitStatus,
			PlistPath:      svc.PlistPath,
			Program:        svc.BinaryPath(),
		})
	}
	return out
}

// jsonServiceDetail is the full representation used by info.
type jsonServiceDetail struct {
	Label             string            `json:"label"`
	Domain            string            `json:"domain"`
	Type              string            `json:"type"`
	Status            string            `json:"status"`
	PID               int               `json:"pid"`
	LastExitStatus    int               `json:"last_exit_status"`
	PlistPath         string            `json:"plist_path,omitempty"`
	Program           string            `json:"program,omitempty"`
	ProgramArgs       []string          `json:"program_args,omitempty"`
	RunAtLoad         bool              `json:"run_at_load"`
	KeepAlive         interface{}       `json:"keep_alive"`
	StartInterval     int               `json:"start_interval,omitempty"`
	WatchPaths        []string          `json:"watch_paths,omitempty"`
	QueueDirectories  []string          `json:"queue_directories,omitempty"`
	WorkingDirectory  string            `json:"working_directory,omitempty"`
	StandardOutPath   string            `json:"stdout_path,omitempty"`
	StandardErrorPath string            `json:"stderr_path,omitempty"`
	EnvironmentVars   map[string]string `json:"environment,omitempty"`
	ExitTimeout       int               `json:"exit_timeout,omitempty"`
	Disabled          bool              `json:"disabled"`
	BlameLine         string            `json:"blame,omitempty"`
}

// toJSONServiceDetail converts an agent.Service to its full JSON representation.
func toJSONServiceDetail(svc *agent.Service) jsonServiceDetail {
	return jsonServiceDetail{
		Label:             svc.Label,
		Domain:            svc.Domain.String(),
		Type:              svc.Type.String(),
		Status:            svc.Status.String(),
		PID:               svc.PID,
		LastExitStatus:    svc.LastExitStatus,
		PlistPath:         svc.PlistPath,
		Program:           svc.BinaryPath(),
		ProgramArgs:       svc.ProgramArgs,
		RunAtLoad:         svc.RunAtLoad,
		KeepAlive:         svc.KeepAlive,
		StartInterval:     svc.StartInterval,
		WatchPaths:        svc.WatchPaths,
		QueueDirectories:  svc.QueueDirectories,
		WorkingDirectory:  svc.WorkingDirectory,
		StandardOutPath:   svc.StandardOutPath,
		StandardErrorPath: svc.StandardErrorPath,
		EnvironmentVars:   svc.EnvironmentVars,
		ExitTimeout:       svc.ExitTimeout,
		Disabled:          svc.Disabled,
		BlameLine:         svc.BlameLine,
	}
}

// ---------------------------------------------------------------------------
// Action JSON type (enable, disable, restart, load, unload)
// ---------------------------------------------------------------------------

type jsonAction struct {
	OK     bool   `json:"ok"`
	Action string `json:"action"`
	Label  string `json:"label"`
}

// ---------------------------------------------------------------------------
// Doctor JSON types
// ---------------------------------------------------------------------------

type jsonDoctor struct {
	Findings []jsonFinding `json:"findings"`
	Summary  jsonSummary   `json:"summary"`
}

type jsonFinding struct {
	Severity   string `json:"severity"`
	Label      string `json:"label"`
	PlistPath  string `json:"plist_path,omitempty"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

type jsonSummary struct {
	Critical int `json:"critical"`
	Warning  int `json:"warning"`
	OK       int `json:"ok"`
}

// toJSONDoctor converts doctor findings to a JSON-serializable structure.
func toJSONDoctor(findings []agent.Finding) jsonDoctor {
	critical, warning, ok := agent.CountBySeverity(findings)

	jf := make([]jsonFinding, 0, len(findings))
	for _, f := range findings {
		jf = append(jf, jsonFinding{
			Severity:   f.Severity.String(),
			Label:      f.Label,
			PlistPath:  f.PlistPath,
			Message:    f.Message,
			Suggestion: f.Suggestion,
		})
	}

	return jsonDoctor{
		Findings: jf,
		Summary: jsonSummary{
			Critical: critical,
			Warning:  warning,
			OK:       ok,
		},
	}
}

// ---------------------------------------------------------------------------
// Create JSON type
// ---------------------------------------------------------------------------

type jsonCreate struct {
	OK        bool   `json:"ok"`
	Action    string `json:"action"`
	Label     string `json:"label"`
	PlistPath string `json:"plist_path"`
	Loaded    bool   `json:"loaded"`
}

// ---------------------------------------------------------------------------
// Import JSON type
// ---------------------------------------------------------------------------

type jsonImport struct {
	OK        bool   `json:"ok"`
	Action    string `json:"action"`
	Label     string `json:"label"`
	PlistPath string `json:"plist_path"`
	Loaded    bool   `json:"loaded"`
}

// ---------------------------------------------------------------------------
// Logs JSON type
// ---------------------------------------------------------------------------

type jsonLogs struct {
	Label  string   `json:"label"`
	Source string   `json:"source"`
	Lines  []string `json:"lines"`
}

// ---------------------------------------------------------------------------
// Edit JSON type
// ---------------------------------------------------------------------------

type jsonEdit struct {
	OK              bool   `json:"ok"`
	Action          string `json:"action"`
	Label           string `json:"label"`
	PlistPath       string `json:"plist_path"`
	ValidationOK    bool   `json:"validation_ok"`
	Reloaded        bool   `json:"reloaded"`
}
