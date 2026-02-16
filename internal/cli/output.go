package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// printJSON encodes v as indented JSON to stdout.
func printJSON(v any) error {
	return fprintJSON(os.Stdout, v)
}

// fprintJSON encodes v as indented JSON to w.
func fprintJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}
