package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:   "disable <label>",
	Short: "Disable a service without unloading it",
	Long:  "Disable a service. The disabled state persists across reboots. This does NOT unload the plist.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, manager, _ := buildDeps()

		label := args[0]
		if err := manager.Disable(label); err != nil {
			return err
		}

		if jsonFlag {
			return printJSON(jsonAction{OK: true, Action: "disable", Label: label})
		}

		fmt.Printf("Disabled %s\n", label)
		return nil
	},
}
