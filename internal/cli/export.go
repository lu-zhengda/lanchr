package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/lanchr/internal/plist"
)

var exportCmd = &cobra.Command{
	Use:   "export <label> [file]",
	Short: "Export a launch agent/daemon to a portable JSON bundle",
	Long:  "Export a launch agent/daemon (its plist file + metadata) to a portable JSON bundle.\nIf no output file is specified, the bundle is written to stdout.",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		scanner, _, _ := buildDeps()

		label := args[0]
		svc, err := scanner.FindByLabel(label)
		if err != nil {
			return fmt.Errorf("failed to find service %q: %w", label, err)
		}

		if svc.PlistPath == "" {
			return fmt.Errorf("service %q has no plist on disk; cannot export", label)
		}

		parser := plist.NewParser()
		pl, err := parser.Parse(svc.PlistPath)
		if err != nil {
			return fmt.Errorf("failed to parse plist for %q: %w", label, err)
		}

		bundle := plist.NewExportBundle(pl, svc.PlistPath, svc.Domain.String(), svc.Type.String())

		// Write to file or stdout.
		if len(args) == 2 {
			outputPath := args[1]
			if err := plist.WriteBundleToFile(outputPath, bundle); err != nil {
				return fmt.Errorf("failed to write export bundle: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Exported %s to %s\n", label, outputPath)
		} else {
			if err := plist.WriteBundle(os.Stdout, bundle); err != nil {
				return fmt.Errorf("failed to write export bundle: %w", err)
			}
		}

		return nil
	},
}
