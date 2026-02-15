package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable <label>",
	Short: "Enable a disabled service",
	Long:  "Enable a service that was previously disabled. The enabled state persists across reboots.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, manager, _ := buildDeps()

		label := args[0]
		if err := manager.Enable(label); err != nil {
			return err
		}

		fmt.Printf("Enabled %s\n", label)
		return nil
	},
}
