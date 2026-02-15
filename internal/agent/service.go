package agent

import (
	"strings"

	"github.com/zhengda-lu/lanchr/internal/platform"
)

// Status represents the runtime state of a service.
type Status int

const (
	StatusStopped  Status = iota
	StatusRunning
	StatusError
	StatusDisabled
)

// String returns a human-readable status name.
func (s Status) String() string {
	switch s {
	case StatusStopped:
		return "stopped"
	case StatusRunning:
		return "running"
	case StatusError:
		return "error"
	case StatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

// Indicator returns a single character for the status.
func (s Status) Indicator() string {
	switch s {
	case StatusRunning:
		return "*"
	case StatusStopped:
		return "-"
	case StatusError:
		return "!"
	case StatusDisabled:
		return "x"
	default:
		return "?"
	}
}

// CalendarInterval represents a StartCalendarInterval entry.
type CalendarInterval struct {
	Minute  *int
	Hour    *int
	Day     *int
	Weekday *int
	Month   *int
}

// KeepAliveConditions holds structured KeepAlive conditions.
type KeepAliveConditions struct {
	SuccessfulExit  *bool
	Crashed         *bool
	PathState       map[string]bool
	OtherJobEnabled map[string]bool
}

// Service represents a macOS launch agent or daemon with both plist data
// and runtime state from launchctl.
type Service struct {
	Label             string
	Domain            platform.Domain
	Type              platform.ServiceType
	Status            Status
	PID               int // -1 if not running
	LastExitStatus    int
	PlistPath         string
	Program           string
	ProgramArgs       []string
	RunAtLoad         bool
	KeepAlive         interface{} // bool or KeepAliveConditions
	StartInterval     int
	CalendarInterval  []CalendarInterval
	WatchPaths        []string
	QueueDirectories  []string
	StandardOutPath   string
	StandardErrorPath string
	WorkingDirectory  string
	EnvironmentVars   map[string]string
	UserName          string
	GroupName         string
	Disabled          bool
	ExitTimeout       int
	ThrottleInterval  int
	Nice              int
	ProcessType       string
	MachServices      map[string]interface{}
	Sockets           map[string]interface{}
	BlameLine         string
}

// IsApple returns true if the service label starts with "com.apple.".
func (s *Service) IsApple() bool {
	return strings.HasPrefix(s.Label, "com.apple.")
}

// IsSIPProtected returns true if the service plist is under /System/Library.
func (s *Service) IsSIPProtected() bool {
	return platform.IsSIPProtected(s.PlistPath)
}

// ServiceTarget returns the launchctl service target for this service.
func (s *Service) ServiceTarget() string {
	return platform.ServiceTarget(s.Domain, s.Label)
}

// DomainTarget returns the launchctl domain target for this service.
func (s *Service) DomainTarget() string {
	return platform.DomainTarget(s.Domain)
}

// HasPlist returns true if the service has a known plist path.
func (s *Service) HasPlist() bool {
	return s.PlistPath != ""
}

// BinaryPath returns the effective binary path.
func (s *Service) BinaryPath() string {
	if s.Program != "" {
		return s.Program
	}
	if len(s.ProgramArgs) > 0 {
		return s.ProgramArgs[0]
	}
	return ""
}
