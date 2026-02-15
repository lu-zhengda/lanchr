package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/lu-zhengda/lanchr/internal/agent"
	"github.com/lu-zhengda/lanchr/internal/launchctl"
	"github.com/lu-zhengda/lanchr/internal/platform"
	"github.com/lu-zhengda/lanchr/internal/plist"
	"github.com/lu-zhengda/lanchr/internal/tui"
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
		if shell, _ := cmd.Root().Flags().GetString("generate-completion"); shell != "" {
			return nil
		}
		return platform.CheckDarwin()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if shell, _ := cmd.Flags().GetString("generate-completion"); shell != "" {
			switch shell {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			default:
				return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
			}
		}
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
	rootCmd.Flags().String("generate-completion", "", "Generate shell completion (bash, zsh, fish)")
	rootCmd.Flags().MarkHidden("generate-completion")

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
