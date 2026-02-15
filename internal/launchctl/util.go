package launchctl

import "strings"

// trimOutput removes trailing whitespace and newlines from command output.
func trimOutput(data []byte) string {
	return strings.TrimSpace(string(data))
}

// IsPermissionDenied checks if an error is likely a permission denied error.
func IsPermissionDenied(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "Operation not permitted") ||
		strings.Contains(msg, "not privileged") ||
		strings.Contains(msg, "Could not write configuration") ||
		strings.Contains(msg, "exit status 36")
}
