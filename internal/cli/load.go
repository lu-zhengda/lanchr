package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

var loadCmd = &cobra.Command{
	Use:   "load <path>",
	Short: "Bootstrap a plist into the appropriate domain",
	Long:  "Load (bootstrap) a plist file into the correct domain based on its path.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, manager, _ := buildDeps()

		plistPath, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		if err := manager.Load(plistPath); err != nil {
			return err
		}

		fmt.Printf("Loaded %s\n", plistPath)
		return nil
	},
}
