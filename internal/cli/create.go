package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhengda-lu/lanchr/internal/plist"
)

var (
	createLabel      string
	createProgram    string
	createArgs       string
	createInterval   int
	createCalendar   string
	createRunAtLoad  bool
	createKeepAlive  bool
	createStdout     string
	createStderr     string
	createWorkingDir string
	createEnv        []string
	createTemplate   string
	createOutput     string
	createLoad       bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Scaffold a new launch agent plist from templates",
	Long:  "Create a new launch agent plist using built-in templates. Supports simple, interval, calendar, keepalive, and watcher templates.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if createLabel == "" {
			return fmt.Errorf("--label is required")
		}
		if createProgram == "" {
			return fmt.Errorf("--program is required")
		}

		var pl plist.LaunchAgentPlist

		// Start from a template if specified.
		if createTemplate != "" {
			tmpl, err := plist.GetTemplate(createTemplate)
			if err != nil {
				return err
			}
			pl = tmpl.Plist
		}

		// Apply explicit flags (override template defaults).
		pl.Label = createLabel
		pl.Program = createProgram

		if createArgs != "" {
			pl.ProgramArguments = append([]string{createProgram}, strings.Split(createArgs, ",")...)
		}

		if createInterval > 0 {
			pl.StartInterval = createInterval
		}

		if createCalendar != "" {
			// Parse cron-like spec: "minute hour day month weekday"
			cal, err := parseCalendarSpec(createCalendar)
			if err != nil {
				return fmt.Errorf("invalid calendar spec: %w", err)
			}
			pl.StartCalendarInterval = cal
		}

		if createRunAtLoad {
			pl.RunAtLoad = true
		}

		if createKeepAlive {
			pl.KeepAlive = true
		}

		if createStdout != "" {
			pl.StandardOutPath = createStdout
		}

		if createStderr != "" {
			pl.StandardErrorPath = createStderr
		}

		if createWorkingDir != "" {
			pl.WorkingDirectory = createWorkingDir
		}

		if len(createEnv) > 0 {
			pl.EnvironmentVariables = make(map[string]string)
			for _, env := range createEnv {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					pl.EnvironmentVariables[parts[0]] = parts[1]
				}
			}
		}

		// Apply defaults (log paths).
		plist.ApplyDefaults(&pl)

		// Determine output path.
		outputPath := createOutput
		if outputPath == "" {
			home, _ := os.UserHomeDir()
			outputPath = filepath.Join(home, "Library", "LaunchAgents", pl.Label+".plist")
		}

		// Write the plist.
		writer := plist.NewWriter()
		if err := writer.Write(&pl, outputPath); err != nil {
			return fmt.Errorf("failed to write plist: %w", err)
		}

		fmt.Printf("Created %s\n", outputPath)

		// Optionally load the plist.
		if createLoad {
			_, manager, _ := buildDeps()
			if err := manager.Load(outputPath); err != nil {
				return fmt.Errorf("failed to load plist: %w", err)
			}
			fmt.Printf("Loaded %s\n", pl.Label)
		}

		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createLabel, "label", "l", "", "Service label (e.g., com.example.myagent)")
	createCmd.Flags().StringVarP(&createProgram, "program", "p", "", "Executable path")
	createCmd.Flags().StringVarP(&createArgs, "args", "a", "", "Program arguments (comma-separated)")
	createCmd.Flags().IntVar(&createInterval, "interval", 0, "StartInterval in seconds")
	createCmd.Flags().StringVar(&createCalendar, "calendar", "", "StartCalendarInterval (cron-like: \"minute hour\")")
	createCmd.Flags().BoolVar(&createRunAtLoad, "run-at-load", false, "Set RunAtLoad to true")
	createCmd.Flags().BoolVar(&createKeepAlive, "keep-alive", false, "Set KeepAlive to true")
	createCmd.Flags().StringVar(&createStdout, "stdout", "", "StandardOutPath")
	createCmd.Flags().StringVar(&createStderr, "stderr", "", "StandardErrorPath")
	createCmd.Flags().StringVar(&createWorkingDir, "working-dir", "", "WorkingDirectory")
	createCmd.Flags().StringArrayVar(&createEnv, "env", nil, "Environment variables (KEY=VAL, repeatable)")
	createCmd.Flags().StringVar(&createTemplate, "template", "", "Built-in template: simple, interval, calendar, keepalive, watcher")
	createCmd.Flags().StringVarP(&createOutput, "output", "o", "", "Output path (default: ~/Library/LaunchAgents/<label>.plist)")
	createCmd.Flags().BoolVar(&createLoad, "load", false, "Bootstrap the plist after creation")
}

// parseCalendarSpec parses a simplified cron-like spec: "minute hour".
func parseCalendarSpec(spec string) (map[string]interface{}, error) {
	parts := strings.Fields(spec)
	result := make(map[string]interface{})

	if len(parts) >= 1 && parts[0] != "*" {
		var minute int
		_, err := fmt.Sscanf(parts[0], "%d", &minute)
		if err != nil {
			return nil, fmt.Errorf("invalid minute: %s", parts[0])
		}
		result["Minute"] = minute
	}

	if len(parts) >= 2 && parts[1] != "*" {
		var hour int
		_, err := fmt.Sscanf(parts[1], "%d", &hour)
		if err != nil {
			return nil, fmt.Errorf("invalid hour: %s", parts[1])
		}
		result["Hour"] = hour
	}

	if len(parts) >= 3 && parts[2] != "*" {
		var day int
		_, err := fmt.Sscanf(parts[2], "%d", &day)
		if err != nil {
			return nil, fmt.Errorf("invalid day: %s", parts[2])
		}
		result["Day"] = day
	}

	if len(parts) >= 4 && parts[3] != "*" {
		var month int
		_, err := fmt.Sscanf(parts[3], "%d", &month)
		if err != nil {
			return nil, fmt.Errorf("invalid month: %s", parts[3])
		}
		result["Month"] = month
	}

	if len(parts) >= 5 && parts[4] != "*" {
		var weekday int
		_, err := fmt.Sscanf(parts[4], "%d", &weekday)
		if err != nil {
			return nil, fmt.Errorf("invalid weekday: %s", parts[4])
		}
		result["Weekday"] = weekday
	}

	return result, nil
}
