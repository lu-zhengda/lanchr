package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/lanchr/internal/plist"
)

var importLoad bool

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import a launch agent/daemon from an export bundle",
	Long:  "Import a launch agent/daemon from a JSON export bundle. Copies the plist to\n~/Library/LaunchAgents/ and optionally loads it.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bundlePath := args[0]

		bundle, err := plist.ReadBundleFromFile(bundlePath)
		if err != nil {
			return fmt.Errorf("failed to read export bundle: %w", err)
		}

		// Determine output path: always place in user's LaunchAgents.
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		outputDir := filepath.Join(home, "Library", "LaunchAgents")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
		}

		outputPath := filepath.Join(outputDir, bundle.Label+".plist")

		// Check if the file already exists.
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("plist already exists at %s; remove it first or use a different label", outputPath)
		}

		// Write the plist from the bundle.
		writer := plist.NewWriter()
		if err := writer.WriteWithoutValidation(&bundle.Plist, outputPath); err != nil {
			return fmt.Errorf("failed to write plist: %w", err)
		}

		// Optionally load the plist.
		loaded := false
		if importLoad {
			_, manager, _ := buildDeps()
			if err := manager.Load(outputPath); err != nil {
				if !jsonFlag {
					fmt.Printf("Imported %s to %s\n", bundle.Label, outputPath)
				}
				return fmt.Errorf("failed to load plist: %w", err)
			}
			loaded = true
		}

		if jsonFlag {
			return printJSON(jsonImport{
				OK:        true,
				Action:    "import",
				Label:     bundle.Label,
				PlistPath: outputPath,
				Loaded:    loaded,
			})
		}

		fmt.Printf("Imported %s to %s\n", bundle.Label, outputPath)
		if loaded {
			fmt.Printf("Loaded %s\n", bundle.Label)
		}
		return nil
	},
}

func init() {
	importCmd.Flags().BoolVar(&importLoad, "load", false, "Bootstrap the plist after import")
}
