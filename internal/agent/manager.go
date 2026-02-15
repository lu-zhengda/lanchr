package agent

import (
	"fmt"

	"github.com/lu-zhengda/lanchr/internal/launchctl"
	"github.com/lu-zhengda/lanchr/internal/platform"
	"github.com/lu-zhengda/lanchr/internal/plist"
)

// Manager performs lifecycle operations on services.
type Manager struct {
	launchctl launchctl.Executor
	scanner   *Scanner
	parser    *plist.Parser
}

// NewManager creates a new service manager.
func NewManager(executor launchctl.Executor, scanner *Scanner, parser *plist.Parser) *Manager {
	return &Manager{
		launchctl: executor,
		scanner:   scanner,
		parser:    parser,
	}
}

// Enable enables a service by its label.
func (m *Manager) Enable(label string) error {
	svc, err := m.scanner.FindByLabel(label)
	if err != nil {
		return fmt.Errorf("failed to find service %q: %w", label, err)
	}

	if svc.IsSIPProtected() {
		return fmt.Errorf("cannot enable %q: service is SIP-protected", label)
	}

	target := platform.ServiceTarget(svc.Domain, svc.Label)
	if err := m.launchctl.Enable(target); err != nil {
		if launchctl.IsPermissionDenied(err) {
			return fmt.Errorf("failed to enable %q: operation requires sudo (system daemon): %w", label, err)
		}
		return fmt.Errorf("failed to enable %q: %w", label, err)
	}
	return nil
}

// Disable disables a service by its label.
func (m *Manager) Disable(label string) error {
	svc, err := m.scanner.FindByLabel(label)
	if err != nil {
		return fmt.Errorf("failed to find service %q: %w", label, err)
	}

	if svc.IsSIPProtected() {
		return fmt.Errorf("cannot disable %q: service is SIP-protected", label)
	}

	target := platform.ServiceTarget(svc.Domain, svc.Label)
	if err := m.launchctl.Disable(target); err != nil {
		if launchctl.IsPermissionDenied(err) {
			return fmt.Errorf("failed to disable %q: operation requires sudo (system daemon): %w", label, err)
		}
		return fmt.Errorf("failed to disable %q: %w", label, err)
	}
	return nil
}

// Restart force-restarts a service by its label using kickstart.
func (m *Manager) Restart(label string) error {
	svc, err := m.scanner.FindByLabel(label)
	if err != nil {
		return fmt.Errorf("failed to find service %q: %w", label, err)
	}

	if svc.IsSIPProtected() {
		return fmt.Errorf("cannot restart %q: service is SIP-protected", label)
	}

	target := platform.ServiceTarget(svc.Domain, svc.Label)
	if err := m.launchctl.Kickstart(target, true); err != nil {
		if launchctl.IsPermissionDenied(err) {
			return fmt.Errorf("failed to restart %q: operation requires sudo: %w", label, err)
		}
		return fmt.Errorf("failed to restart %q: %w", label, err)
	}
	return nil
}

// Load bootstraps a plist into the appropriate domain.
func (m *Manager) Load(plistPath string) error {
	pl, err := m.parser.Parse(plistPath)
	if err != nil {
		return fmt.Errorf("failed to parse plist %s: %w", plistPath, err)
	}

	domain := platform.DomainFromPath(plistPath)
	domainTarget := platform.DomainTarget(domain)

	if err := m.launchctl.Bootstrap(domainTarget, plistPath); err != nil {
		if launchctl.IsPermissionDenied(err) {
			return fmt.Errorf("failed to load %q: operation requires sudo: %w", pl.Label, err)
		}
		return fmt.Errorf("failed to load %q: %w", pl.Label, err)
	}
	return nil
}

// Unload removes a service from its domain.
func (m *Manager) Unload(label string) error {
	svc, err := m.scanner.FindByLabel(label)
	if err != nil {
		return fmt.Errorf("failed to find service %q: %w", label, err)
	}

	if svc.IsSIPProtected() {
		return fmt.Errorf("cannot unload %q: service is SIP-protected", label)
	}

	target := platform.ServiceTarget(svc.Domain, svc.Label)
	if err := m.launchctl.Bootout(target); err != nil {
		if launchctl.IsPermissionDenied(err) {
			return fmt.Errorf("failed to unload %q: operation requires sudo: %w", label, err)
		}
		return fmt.Errorf("failed to unload %q: %w", label, err)
	}
	return nil
}

// Info returns detailed information about a service by looking up both
// the plist data and the live runtime state via launchctl print.
func (m *Manager) Info(label string) (*Service, error) {
	svc, err := m.scanner.FindByLabel(label)
	if err != nil {
		return nil, fmt.Errorf("failed to find service %q: %w", label, err)
	}

	// Try to enrich with launchctl print data.
	target := platform.ServiceTarget(svc.Domain, svc.Label)
	info, err := m.launchctl.PrintService(target)
	if err == nil && info != nil {
		if info.PID > 0 {
			svc.PID = info.PID
			svc.Status = StatusRunning
		}
		if info.State == "running" {
			svc.Status = StatusRunning
		}
		if info.Program != "" && svc.Program == "" {
			svc.Program = info.Program
		}
	}

	// Try to get blame info.
	blame, err := m.launchctl.Blame(target)
	if err == nil {
		svc.BlameLine = blame
	}

	return svc, nil
}
