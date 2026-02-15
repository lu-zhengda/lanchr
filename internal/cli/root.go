package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/zhengda-lu/lanchr/internal/agent"
	"github.com/zhengda-lu/lanchr/internal/launchctl"
	"github.com/zhengda-lu/lanchr/internal/platform"
	"github.com/zhengda-lu/lanchr/internal/plist"
	"github.com/zhengda-lu/lanchr/internal/tui"
)

var (
	// Set via ldflags at build time.
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "lanchr",
	Short:   "A macOS launch agent/daemon manager",
	Long:    "lanchr provides a unified interface for inspecting, managing, and troubleshooting\nmacOS launch agents and daemons. Launch without subcommands for interactive TUI mode.",
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "help" || cmd.Flags().Changed("version") {
			return nil
		}
		return platform.CheckDarwin()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := launchctl.NewDefaultExecutor()
		parser := plist.NewParser()
		scanner := agent.NewScanner(parser, exec)
		manager := agent.NewManager(exec, scanner, parser)
		doctor := agent.NewDoctor(scanner)

		model := tui.New(scanner, manager, doctor, version)
		p := tea.NewProgram(model, tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("lanchr %s\n", version))
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(enableCmd)
	rootCmd.AddCommand(disableCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(unloadCmd)
}

// buildDeps creates the common dependencies for CLI commands.
func buildDeps() (*agent.Scanner, *agent.Manager, *agent.Doctor) {
	exec := launchctl.NewDefaultExecutor()
	parser := plist.NewParser()
	scanner := agent.NewScanner(parser, exec)
	manager := agent.NewManager(exec, scanner, parser)
	doctor := agent.NewDoctor(scanner)
	return scanner, manager, doctor
}
