package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/lanchr/internal/agent"
	"github.com/lu-zhengda/lanchr/internal/platform"
)

var (
	listDomain  string
	listStatus  string
	listType    string
	listJSON    bool
	listNoApple bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all agents and daemons",
	Long:  "List all launch agents and daemons across all domains with their status, PID, and binary path.",
	RunE: func(cmd *cobra.Command, args []string) error {
		scanner, _, _ := buildDeps()

		services, err := scanner.ScanAll()
		if err != nil {
			return fmt.Errorf("failed to scan services: %w", err)
		}

		// Apply filters.
		var filtered []agent.Service
		for _, svc := range services {
			if listDomain != "" {
				switch listDomain {
				case "user":
					if svc.Domain != platform.DomainUser {
						continue
					}
				case "global":
					if svc.Domain != platform.DomainGlobal {
						continue
					}
				case "system":
					if svc.Domain != platform.DomainSystem {
						continue
					}
				}
			}

			if listStatus != "" {
				switch listStatus {
				case "running":
					if svc.Status != agent.StatusRunning {
						continue
					}
				case "stopped":
					if svc.Status != agent.StatusStopped {
						continue
					}
				case "error":
					if svc.Status != agent.StatusError {
						continue
					}
				}
			}

			if listType != "" {
				switch listType {
				case "agent":
					if svc.Type != platform.TypeAgent {
						continue
					}
				case "daemon":
					if svc.Type != platform.TypeDaemon {
						continue
					}
				}
			}

			if listNoApple && svc.IsApple() {
				continue
			}

			filtered = append(filtered, svc)
		}

		if listJSON {
			return outputJSON(filtered)
		}

		return outputTable(filtered)
	},
}

func init() {
	listCmd.Flags().StringVarP(&listDomain, "domain", "d", "", "Filter by domain: user, global, system")
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "Filter by status: running, stopped, error")
	listCmd.Flags().StringVarP(&listType, "type", "t", "", "Filter by type: agent, daemon")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
	listCmd.Flags().BoolVar(&listNoApple, "no-apple", false, "Hide com.apple.* services")
}

func outputTable(services []agent.Service) error {
	fmt.Printf("%-6s  %-6s  %-42s  %-8s  %s\n", "STATUS", "PID", "LABEL", "DOMAIN", "BINARY")

	for _, svc := range services {
		indicator := svc.Status.Indicator()
		pid := "-"
		if svc.PID > 0 {
			pid = fmt.Sprintf("%d", svc.PID)
		}

		binary := svc.BinaryPath()
		if len(binary) > 50 {
			binary = binary[:47] + "..."
		}

		label := svc.Label
		if len(label) > 42 {
			label = label[:39] + "..."
		}

		fmt.Printf("  %s     %-6s  %-42s  %-8s  %s\n",
			indicator, pid, label, svc.Domain.String(), binary)
	}

	return nil
}

type jsonService struct {
	Label          string `json:"label"`
	Domain         string `json:"domain"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	PID            int    `json:"pid"`
	LastExitStatus int    `json:"last_exit_status"`
	PlistPath      string `json:"plist_path,omitempty"`
	Program        string `json:"program,omitempty"`
}

func outputJSON(services []agent.Service) error {
	var out []jsonService
	for _, svc := range services {
		out = append(out, jsonService{
			Label:          svc.Label,
			Domain:         svc.Domain.String(),
			Type:           svc.Type.String(),
			Status:         svc.Status.String(),
			PID:            svc.PID,
			LastExitStatus: svc.LastExitStatus,
			PlistPath:      svc.PlistPath,
			Program:        svc.BinaryPath(),
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// formatProgramArgs joins program arguments for display.
func formatProgramArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return strings.Join(args, " ")
}
