package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/lanchr/internal/logs"
)

var (
	logsFollow  bool
	logsLines   int
	logsStderr  bool
	logsStdout  bool
	logsUnified bool
)

var logsCmd = &cobra.Command{
	Use:   "logs <label>",
	Short: "View logs for a service",
	Long:  "Tail stdout/stderr logs for a service. Falls back to unified logging if no explicit log paths are set.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scanner, _, _ := buildDeps()

		label := args[0]
		svc, err := scanner.FindByLabel(label)
		if err != nil {
			return fmt.Errorf("failed to find service %q: %w", label, err)
		}

		tailer := logs.NewTailer()
		unified := logs.NewUnifiedLog()

		// Determine which log sources to use.
		if logsUnified || (svc.StandardOutPath == "" && svc.StandardErrorPath == "") {
			// Use unified logging.
			processName := filepath.Base(svc.BinaryPath())
			if processName == "" || processName == "." {
				processName = label
			}

			if logsFollow {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				ch, err := unified.Stream(ctx, processName)
				if err != nil {
					return fmt.Errorf("failed to stream logs: %w", err)
				}

				for line := range ch {
					fmt.Println(line)
				}
				return nil
			}

			lines, err := unified.Show(processName, time.Hour, logsLines)
			if err != nil {
				return fmt.Errorf("failed to show unified logs: %w", err)
			}

			for _, line := range lines {
				fmt.Println(line)
			}
			return nil
		}

		// Use file-based logs.
		if !logsStderr && svc.StandardOutPath != "" {
			fmt.Printf("--- stdout: %s ---\n", svc.StandardOutPath)
			if logsFollow {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				ch, err := tailer.Follow(ctx, svc.StandardOutPath)
				if err != nil {
					return fmt.Errorf("failed to follow stdout: %w", err)
				}
				for line := range ch {
					fmt.Println(line)
				}
				return nil
			}

			lines, err := tailer.Tail(svc.StandardOutPath, logsLines)
			if err != nil {
				fmt.Printf("  (error reading: %v)\n", err)
			} else {
				for _, line := range lines {
					fmt.Println(line)
				}
			}
		}

		if !logsStdout && svc.StandardErrorPath != "" {
			fmt.Printf("--- stderr: %s ---\n", svc.StandardErrorPath)
			if logsFollow {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				ch, err := tailer.Follow(ctx, svc.StandardErrorPath)
				if err != nil {
					return fmt.Errorf("failed to follow stderr: %w", err)
				}
				for line := range ch {
					fmt.Println(line)
				}
				return nil
			}

			lines, err := tailer.Tail(svc.StandardErrorPath, logsLines)
			if err != nil {
				fmt.Printf("  (error reading: %v)\n", err)
			} else {
				for _, line := range lines {
					fmt.Println(line)
				}
			}
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output (like tail -f)")
	logsCmd.Flags().IntVarP(&logsLines, "lines", "n", 50, "Number of lines to show")
	logsCmd.Flags().BoolVar(&logsStderr, "stderr", false, "Show only stderr")
	logsCmd.Flags().BoolVar(&logsStdout, "stdout", false, "Show only stdout")
	logsCmd.Flags().BoolVar(&logsUnified, "unified", false, "Use macOS unified logging (log show --predicate)")
}
