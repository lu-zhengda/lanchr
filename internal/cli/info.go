package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <label>",
	Short: "Show detailed info for a specific service",
	Long:  "Display all parsed plist keys plus live runtime state from launchctl print.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, manager, _ := buildDeps()

		label := args[0]
		svc, err := manager.Info(label)
		if err != nil {
			return fmt.Errorf("failed to get info for %q: %w", label, err)
		}

		printField := func(name, value string) {
			fmt.Printf("%-22s %s\n", name+":", value)
		}

		printField("Label", svc.Label)
		printField("Domain", svc.Domain.String())
		printField("Type", svc.Type.String())
		printField("State", svc.Status.String())

		if svc.PID > 0 {
			printField("PID", fmt.Sprintf("%d", svc.PID))
		} else {
			printField("PID", "-")
		}

		if svc.PlistPath != "" {
			printField("Plist Path", svc.PlistPath)
		}

		if svc.Program != "" {
			printField("Program", svc.Program)
		}

		if len(svc.ProgramArgs) > 0 {
			printField("Arguments", strings.Join(svc.ProgramArgs, " "))
		} else {
			printField("Arguments", "(none)")
		}

		printField("Run At Load", fmt.Sprintf("%v", svc.RunAtLoad))

		if svc.KeepAlive != nil {
			printField("Keep Alive", fmt.Sprintf("%v", svc.KeepAlive))
		} else {
			printField("Keep Alive", "(none)")
		}

		if svc.StartInterval > 0 {
			printField("Start Interval", fmt.Sprintf("%ds", svc.StartInterval))
		} else {
			printField("Start Interval", "(none)")
		}

		if len(svc.WatchPaths) > 0 {
			printField("Watch Paths", strings.Join(svc.WatchPaths, ", "))
		} else {
			printField("Watch Paths", "(none)")
		}

		if svc.WorkingDirectory != "" {
			printField("Working Directory", svc.WorkingDirectory)
		} else {
			printField("Working Directory", "(none)")
		}

		if svc.StandardOutPath != "" {
			printField("Stdout Path", svc.StandardOutPath)
		} else {
			printField("Stdout Path", "(none)")
		}

		if svc.StandardErrorPath != "" {
			printField("Stderr Path", svc.StandardErrorPath)
		} else {
			printField("Stderr Path", "(none)")
		}

		if len(svc.EnvironmentVars) > 0 {
			var envParts []string
			for k, v := range svc.EnvironmentVars {
				envParts = append(envParts, fmt.Sprintf("%s=%s", k, v))
			}
			printField("Environment", strings.Join(envParts, ", "))
		}

		if svc.ExitTimeout > 0 {
			printField("Exit Timeout", fmt.Sprintf("%ds", svc.ExitTimeout))
		}

		printField("Last Exit Code", fmt.Sprintf("%d", svc.LastExitStatus))
		printField("Disabled", fmt.Sprintf("%v", svc.Disabled))

		if svc.BlameLine != "" {
			printField("Blame", svc.BlameLine)
		}

		return nil
	},
}
