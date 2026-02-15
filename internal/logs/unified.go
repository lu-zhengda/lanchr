package logs

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// LogEntry represents a single entry from the macOS unified logging system.
type LogEntry struct {
	Timestamp time.Time
	Process   string
	PID       int
	Level     string // Default, Info, Debug, Error, Fault
	Message   string
}

// String returns a formatted log entry.
func (e LogEntry) String() string {
	ts := e.Timestamp.Format("2006-01-02 15:04:05.000")
	return fmt.Sprintf("%s [%s] %s", ts, e.Level, e.Message)
}

// UnifiedLog queries the macOS unified logging system for a process.
type UnifiedLog struct{}

// NewUnifiedLog creates a new unified log reader.
func NewUnifiedLog() *UnifiedLog {
	return &UnifiedLog{}
}

// Show returns recent log entries for a process.
func (u *UnifiedLog) Show(processName string, duration time.Duration, lines int) ([]string, error) {
	lastArg := formatDuration(duration)
	predicate := fmt.Sprintf("process == %q", processName)

	cmd := exec.Command("log", "show",
		"--predicate", predicate,
		"--last", lastArg,
		"--style", "compact",
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query unified log for %q: %w", processName, err)
	}

	var result []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		// Skip header lines from log show.
		if strings.HasPrefix(line, "Filtering the log data") ||
			strings.HasPrefix(line, "Skipping info and debug") ||
			strings.HasPrefix(line, "Timestamp") ||
			line == "" {
			continue
		}
		result = append(result, line)
	}

	// Return only the last N lines.
	if lines > 0 && lines < len(result) {
		result = result[len(result)-lines:]
	}

	return result, nil
}

// Stream opens a live stream of log entries for a process.
// The caller should cancel the context to stop streaming.
func (u *UnifiedLog) Stream(ctx context.Context, processName string) (<-chan string, error) {
	predicate := fmt.Sprintf("process == %q", processName)

	cmd := exec.CommandContext(ctx, "log", "stream",
		"--predicate", predicate,
		"--style", "compact",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start log stream: %w", err)
	}

	ch := make(chan string, 100)

	go func() {
		defer close(ch)
		defer cmd.Wait() //nolint:errcheck

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			// Skip header lines.
			if strings.HasPrefix(line, "Filtering the log data") || line == "" {
				continue
			}
			select {
			case ch <- line:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// formatDuration converts a Go duration to a "log show --last" format (e.g., "1h", "30m").
func formatDuration(d time.Duration) string {
	if d >= time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}
