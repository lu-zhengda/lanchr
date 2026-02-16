package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/lanchr/internal/agent"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose broken plists, orphaned agents, and missing binaries",
	Long:  "Run a suite of health checks across all launch agents and daemons and print a diagnostic report.",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _, doctor := buildDeps()

		findings, err := doctor.Check()
		if err != nil {
			return fmt.Errorf("failed to run doctor: %w", err)
		}

		if jsonFlag {
			return printJSON(toJSONDoctor(findings))
		}

		if len(findings) == 0 {
			fmt.Println("DOCTOR REPORT")
			fmt.Println("=============")
			fmt.Println()
			fmt.Println("All services passed health checks.")
			return nil
		}

		critical, warning, okCount := agent.CountBySeverity(findings)
		totalServices := critical + warning + okCount

		fmt.Println("DOCTOR REPORT")
		fmt.Println("=============")
		fmt.Println()

		if critical > 0 {
			fmt.Printf("CRITICAL (%d)\n", critical)
			for _, f := range findings {
				if f.Severity == agent.SeverityCritical {
					fmt.Printf("  [!] %s: %s\n", f.Label, f.Message)
					if f.Suggestion != "" {
						fmt.Printf("      Suggestion: %s\n", f.Suggestion)
					}
				}
			}
			fmt.Println()
		}

		if warning > 0 {
			fmt.Printf("WARNING (%d)\n", warning)
			for _, f := range findings {
				if f.Severity == agent.SeverityWarning {
					fmt.Printf("  [~] %s: %s\n", f.Label, f.Message)
					if f.Suggestion != "" {
						fmt.Printf("      Suggestion: %s\n", f.Suggestion)
					}
				}
			}
			fmt.Println()
		}

		// Count services that passed (estimated from total minus findings).
		_ = totalServices
		fmt.Println("Run 'lanchr list' to see all services.")

		return nil
	},
}
