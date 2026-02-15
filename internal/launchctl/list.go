package launchctl

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
)

// List parses the output of "launchctl list".
// The output is tab-separated: PID \t Status \t Label
// The first line is a header and is skipped.
// A PID of "-" means the service is not running (stored as -1).
func (e *DefaultExecutor) List() ([]ListEntry, error) {
	out, err := e.run("list")
	if err != nil {
		return nil, err
	}

	return parseListOutput(out), nil
}

func parseListOutput(data []byte) []ListEntry {
	var entries []ListEntry
	scanner := bufio.NewScanner(bytes.NewReader(data))

	// Skip the header line.
	if scanner.Scan() {
		// discard header
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		fields := strings.SplitN(line, "\t", 3)
		if len(fields) < 3 {
			// Try space-separated as a fallback.
			fields = strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
		}

		entry := ListEntry{
			PID:    -1,
			Status: 0,
			Label:  strings.TrimSpace(fields[len(fields)-1]),
		}

		pidStr := strings.TrimSpace(fields[0])
		if pidStr != "-" && pidStr != "" {
			if pid, err := strconv.Atoi(pidStr); err == nil {
				entry.PID = pid
			}
		}

		statusStr := strings.TrimSpace(fields[1])
		if statusStr != "-" && statusStr != "" {
			if status, err := strconv.Atoi(statusStr); err == nil {
				entry.Status = status
			}
		}

		entries = append(entries, entry)
	}

	return entries
}
