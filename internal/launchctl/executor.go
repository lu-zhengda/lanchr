package launchctl

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// Executor runs launchctl commands and returns structured output.
type Executor interface {
	// List returns parsed output of "launchctl list".
	List() ([]ListEntry, error)

	// PrintService returns parsed output of "launchctl print <service-target>".
	PrintService(serviceTarget string) (*ServiceInfo, error)

	// PrintDisabled returns the disabled services map for a domain.
	PrintDisabled(domainTarget string) (map[string]bool, error)

	// Blame returns the reason a service was launched.
	Blame(serviceTarget string) (string, error)

	// Enable enables a service.
	Enable(serviceTarget string) error

	// Disable disables a service.
	Disable(serviceTarget string) error

	// Bootstrap loads a plist into a domain.
	Bootstrap(domainTarget string, plistPath string) error

	// Bootout removes a service from a domain.
	Bootout(serviceTarget string) error

	// Kickstart restarts a service.
	Kickstart(serviceTarget string, kill bool) error

	// Kill sends a signal to a service.
	Kill(signal string, serviceTarget string) error
}

// ListEntry is a row from "launchctl list".
type ListEntry struct {
	PID    int    // -1 if not running
	Status int    // last exit status
	Label  string
}

// ServiceInfo holds parsed output from "launchctl print <service-target>".
type ServiceInfo struct {
	State        string // "running", "waiting", etc.
	PID          int
	Path         string
	BundleID     string
	Program      string
	Type         string // "LaunchAgent", "LaunchDaemon"
	Runs         int
	LastExitCode string
	ExitTimeout  int
	Domain       string
}

// CmdRunner abstracts shell command execution for testability.
type CmdRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// RealCmdRunner executes real shell commands.
type RealCmdRunner struct{}

// Run executes a command and returns its combined stdout.
func (r *RealCmdRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = io.Discard
	return cmd.Output()
}

// DefaultExecutor shells out to /bin/launchctl.
type DefaultExecutor struct {
	runner CmdRunner
}

// NewDefaultExecutor creates an executor that uses real shell commands.
func NewDefaultExecutor() *DefaultExecutor {
	return &DefaultExecutor{runner: &RealCmdRunner{}}
}

// NewExecutorWithRunner creates an executor with a custom command runner (for testing).
func NewExecutorWithRunner(runner CmdRunner) *DefaultExecutor {
	return &DefaultExecutor{runner: runner}
}

func (e *DefaultExecutor) run(args ...string) ([]byte, error) {
	ctx := context.Background()
	out, err := e.runner.Run(ctx, "launchctl", args...)
	if err != nil {
		return out, fmt.Errorf("launchctl %v: %w", args, err)
	}
	return out, nil
}

// Enable enables a service.
func (e *DefaultExecutor) Enable(serviceTarget string) error {
	_, err := e.run("enable", serviceTarget)
	return err
}

// Disable disables a service.
func (e *DefaultExecutor) Disable(serviceTarget string) error {
	_, err := e.run("disable", serviceTarget)
	return err
}

// Bootstrap loads a plist into a domain.
func (e *DefaultExecutor) Bootstrap(domainTarget string, plistPath string) error {
	_, err := e.run("bootstrap", domainTarget, plistPath)
	return err
}

// Bootout removes a service from a domain.
func (e *DefaultExecutor) Bootout(serviceTarget string) error {
	_, err := e.run("bootout", serviceTarget)
	return err
}

// Kickstart restarts a service.
func (e *DefaultExecutor) Kickstart(serviceTarget string, kill bool) error {
	if kill {
		_, err := e.run("kickstart", "-kp", serviceTarget)
		return err
	}
	_, err := e.run("kickstart", "-p", serviceTarget)
	return err
}

// Kill sends a signal to a service.
func (e *DefaultExecutor) Kill(signal string, serviceTarget string) error {
	_, err := e.run("kill", signal, serviceTarget)
	return err
}

// Blame returns the reason a service was launched.
func (e *DefaultExecutor) Blame(serviceTarget string) (string, error) {
	out, err := e.run("blame", serviceTarget)
	if err != nil {
		return "", err
	}
	return trimOutput(out), nil
}
