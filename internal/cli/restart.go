package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart <label>",
	Short: "Force restart a running service",
	Long:  "Equivalent to launchctl kickstart -k. Stops and starts the service.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, manager, _ := buildDeps()

		label := args[0]
		if err := manager.Restart(label); err != nil {
			return err
		}

		if jsonFlag {
			return printJSON(jsonAction{OK: true, Action: "restart", Label: label})
		}

		fmt.Printf("Restarted %s\n", label)
		return nil
	},
}
