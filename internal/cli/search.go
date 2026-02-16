package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/lanchr/internal/agent"
)

var (
	searchRegex bool
	searchPath  bool
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search services by label, path, or plist content",
	Long:  "Search across label, program path, program arguments, and plist filename. Supports glob patterns and regex.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scanner, _, _ := buildDeps()

		query := args[0]
		services, err := scanner.ScanAll()
		if err != nil {
			return fmt.Errorf("failed to scan services: %w", err)
		}

		var matches []agent.Service

		if searchRegex {
			re, err := regexp.Compile(query)
			if err != nil {
				return fmt.Errorf("invalid regex %q: %w", query, err)
			}
			for _, svc := range services {
				if re.MatchString(svc.Label) ||
					re.MatchString(svc.BinaryPath()) ||
					re.MatchString(svc.PlistPath) ||
					re.MatchString(formatProgramArgs(svc.ProgramArgs)) {
					matches = append(matches, svc)
				}
			}
		} else if searchPath {
			q := strings.ToLower(query)
			for _, svc := range services {
				if matchGlob(q, strings.ToLower(svc.BinaryPath())) ||
					matchGlob(q, strings.ToLower(svc.PlistPath)) {
					matches = append(matches, svc)
				}
			}
		} else {
			q := strings.ToLower(query)
			for _, svc := range services {
				if strings.Contains(strings.ToLower(svc.Label), q) ||
					strings.Contains(strings.ToLower(svc.BinaryPath()), q) ||
					strings.Contains(strings.ToLower(svc.PlistPath), q) ||
					strings.Contains(strings.ToLower(formatProgramArgs(svc.ProgramArgs)), q) {
					matches = append(matches, svc)
				}
			}
		}

		if jsonFlag {
			return printJSON(toJSONServices(matches))
		}

		if len(matches) == 0 {
			fmt.Println("No services found matching query.")
			return nil
		}

		fmt.Printf("Found %d service(s):\n\n", len(matches))
		return outputTable(matches)
	},
}

func init() {
	searchCmd.Flags().BoolVar(&searchRegex, "regex", false, "Use regex matching")
	searchCmd.Flags().BoolVar(&searchPath, "path", false, "Search by path with glob patterns")
}

// matchGlob performs a simple glob match (only supports * wildcard).
func matchGlob(pattern, s string) bool {
	if pattern == "" {
		return s == ""
	}

	// Simple containment if no glob characters.
	if !strings.Contains(pattern, "*") {
		return strings.Contains(s, pattern)
	}

	// Split on * and check if all parts appear in order.
	parts := strings.Split(pattern, "*")
	pos := 0
	for _, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(s[pos:], part)
		if idx < 0 {
			return false
		}
		pos += idx + len(part)
	}
	return true
}
