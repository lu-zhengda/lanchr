package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var unloadCmd = &cobra.Command{
	Use:   "unload <label>",
	Short: "Remove a service from its domain",
	Long:  "Unload (bootout) a service from the running launchd domain.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, manager, _ := buildDeps()

		label := args[0]
		if err := manager.Unload(label); err != nil {
			return err
		}

		if jsonFlag {
			return printJSON(jsonAction{OK: true, Action: "unload", Label: label})
		}

		fmt.Printf("Unloaded %s\n", label)
		return nil
	},
}
