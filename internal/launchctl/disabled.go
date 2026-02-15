package launchctl

import (
	"bufio"
	"bytes"
	"strings"
)

// PrintDisabled returns the disabled services map for a domain.
// It parses the output of "launchctl print-disabled <domain-target>".
// The output contains lines like:
//
//	"com.example.service" => disabled
//	"com.example.other" => enabled
func (e *DefaultExecutor) PrintDisabled(domainTarget string) (map[string]bool, error) {
	out, err := e.run("print-disabled", domainTarget)
	if err != nil {
		// On error, return an empty map rather than failing entirely.
		// Some domains may not be accessible without privileges.
		return make(map[string]bool), nil
	}

	return parsePrintDisabledOutput(out), nil
}

func parsePrintDisabledOutput(data []byte) map[string]bool {
	result := make(map[string]bool)
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line == "{" || line == "}" {
			continue
		}

		// Lines look like: "com.example.service" => disabled
		// or: "com.example.service" => true
		parts := strings.SplitN(line, "=>", 2)
		if len(parts) != 2 {
			continue
		}

		label := strings.TrimSpace(parts[0])
		label = strings.Trim(label, "\"")

		value := strings.TrimSpace(parts[1])
		// "disabled" or "true" means the service is disabled.
		isDisabled := value == "disabled" || value == "true"

		result[label] = isDisabled
	}

	return result
}
