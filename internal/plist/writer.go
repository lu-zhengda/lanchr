package plist

import (
	"fmt"
	"os"

	goplist "howett.net/plist"
)

// Writer generates plist XML files.
type Writer struct{}

// NewWriter creates a new plist writer.
func NewWriter() *Writer {
	return &Writer{}
}

// Write serializes a LaunchAgentPlist to XML format at the given path.
func (w *Writer) Write(pl *LaunchAgentPlist, path string) error {
	if errs := w.Validate(pl); len(errs) > 0 {
		return fmt.Errorf("failed to validate plist: %s", errs[0].Message)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create plist file %s: %w", path, err)
	}
	defer f.Close()

	encoder := goplist.NewEncoderForFormat(f, goplist.XMLFormat)
	encoder.Indent("\t")
	if err := encoder.Encode(pl); err != nil {
		return fmt.Errorf("failed to encode plist: %w", err)
	}

	return nil
}

// WriteWithoutValidation serializes a LaunchAgentPlist to XML format at the
// given path, skipping binary-exists validation. This is useful for importing
// plists where the referenced binary may not be installed yet.
func (w *Writer) WriteWithoutValidation(pl *LaunchAgentPlist, path string) error {
	if pl.Label == "" {
		return fmt.Errorf("failed to validate plist: Label is required")
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create plist file %s: %w", path, err)
	}
	defer f.Close()

	encoder := goplist.NewEncoderForFormat(f, goplist.XMLFormat)
	encoder.Indent("\t")
	if err := encoder.Encode(pl); err != nil {
		return fmt.Errorf("failed to encode plist: %w", err)
	}

	return nil
}

// Validate checks a plist for common issues.
func (w *Writer) Validate(pl *LaunchAgentPlist) []ValidationError {
	var errs []ValidationError

	if pl.Label == "" {
		errs = append(errs, ValidationError{
			Field:   "Label",
			Message: "Label is required",
		})
	}

	if pl.Program == "" && len(pl.ProgramArguments) == 0 {
		errs = append(errs, ValidationError{
			Field:   "Program",
			Message: "either Program or ProgramArguments is required",
		})
	}

	programPath := pl.ProgramPath()
	if programPath != "" {
		if _, err := os.Stat(programPath); os.IsNotExist(err) {
			errs = append(errs, ValidationError{
				Field:   "Program",
				Message: fmt.Sprintf("binary not found at %s", programPath),
			})
		}
	}

	return errs
}
