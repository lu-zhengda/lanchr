package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var editReload bool

var editCmd = &cobra.Command{
	Use:   "edit <label>",
	Short: "Open the plist in $EDITOR with validation on save",
	Long:  "Open the plist for a service in your preferred editor ($EDITOR or $VISUAL). Optionally reload the service after editing.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scanner, manager, _ := buildDeps()

		label := args[0]
		svc, err := scanner.FindByLabel(label)
		if err != nil {
			return fmt.Errorf("failed to find service %q: %w", label, err)
		}

		if svc.PlistPath == "" {
			return fmt.Errorf("service %q has no plist on disk", label)
		}

		if svc.IsSIPProtected() {
			return fmt.Errorf("cannot edit %q: service is SIP-protected", label)
		}

		// Determine editor.
		editor := os.Getenv("VISUAL")
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}
		if editor == "" {
			editor = "vi"
		}

		// Open the plist in the editor.
		editExec := exec.Command(editor, svc.PlistPath)
		editExec.Stdin = os.Stdin
		editExec.Stdout = os.Stdout
		editExec.Stderr = os.Stderr

		if err := editExec.Run(); err != nil {
			return fmt.Errorf("failed to run editor: %w", err)
		}

		// Validate the plist after editing.
		validateCmd := exec.Command("plutil", "-lint", svc.PlistPath)
		validateOut, err := validateCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Warning: plist validation failed:\n%s\n", string(validateOut))
		} else {
			fmt.Println("Plist validation passed.")
		}

		// Optionally reload.
		if editReload {
			fmt.Printf("Reloading %s...\n", label)

			// Bootout then bootstrap.
			_ = manager.Unload(label)
			if err := manager.Load(svc.PlistPath); err != nil {
				return fmt.Errorf("failed to reload service: %w", err)
			}
			fmt.Println("Service reloaded.")
		}

		return nil
	},
}

func init() {
	editCmd.Flags().BoolVar(&editReload, "reload", false, "Bootout and bootstrap the service after editing")
}
