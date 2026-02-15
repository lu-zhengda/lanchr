package launchctl

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
)

// PrintService parses the output of "launchctl print <service-target>".
// The output is a structured dump of service properties.
// We parse defensively, treating missing fields as optional.
func (e *DefaultExecutor) PrintService(serviceTarget string) (*ServiceInfo, error) {
	out, err := e.run("print", serviceTarget)
	if err != nil {
		return nil, err
	}

	return parsePrintServiceOutput(out), nil
}

func parsePrintServiceOutput(data []byte) *ServiceInfo {
	info := &ServiceInfo{
		PID: -1,
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse "key = value" patterns.
		if idx := strings.Index(line, " = "); idx != -1 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+3:])

			switch key {
			case "state":
				info.State = value
			case "pid":
				if pid, err := strconv.Atoi(value); err == nil {
					info.PID = pid
				}
			case "path":
				info.Path = value
			case "bundle identifier":
				info.BundleID = value
			case "program":
				info.Program = value
			case "type":
				info.Type = value
			case "runs":
				if runs, err := strconv.Atoi(value); err == nil {
					info.Runs = runs
				}
			case "last exit code":
				info.LastExitCode = value
			case "exit timeout":
				// Value may be something like "5" or "5 seconds".
				valParts := strings.Fields(value)
				if len(valParts) > 0 {
					if timeout, err := strconv.Atoi(valParts[0]); err == nil {
						info.ExitTimeout = timeout
					}
				}
			case "domain":
				info.Domain = value
			}
		}
	}

	return info
}
