package plist

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// ExportBundle is a portable JSON representation of a launch agent/daemon.
type ExportBundle struct {
	Version   int              `json:"version"`
	ExportedAt string          `json:"exported_at"`
	Label     string           `json:"label"`
	Domain    string           `json:"domain"`
	Type      string           `json:"type"`
	PlistPath string           `json:"plist_path"`
	Plist     LaunchAgentPlist `json:"plist"`
}

// NewExportBundle creates an ExportBundle from a parsed plist and its metadata.
func NewExportBundle(pl *LaunchAgentPlist, plistPath, domain, serviceType string) *ExportBundle {
	return &ExportBundle{
		Version:    1,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Label:      pl.Label,
		Domain:     domain,
		Type:       serviceType,
		PlistPath:  plistPath,
		Plist:      *pl,
	}
}

// WriteBundle serializes an ExportBundle to the given writer as JSON.
func WriteBundle(w io.Writer, bundle *ExportBundle) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(bundle); err != nil {
		return fmt.Errorf("failed to encode export bundle: %w", err)
	}
	return nil
}

// WriteBundleToFile serializes an ExportBundle to a file.
func WriteBundleToFile(path string, bundle *ExportBundle) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create export file %s: %w", path, err)
	}
	defer f.Close()

	return WriteBundle(f, bundle)
}

// ReadBundle reads an ExportBundle from the given reader.
func ReadBundle(r io.Reader) (*ExportBundle, error) {
	var bundle ExportBundle
	dec := json.NewDecoder(r)
	if err := dec.Decode(&bundle); err != nil {
		return nil, fmt.Errorf("failed to decode export bundle: %w", err)
	}

	if bundle.Version == 0 {
		return nil, fmt.Errorf("invalid export bundle: missing version")
	}
	if bundle.Label == "" {
		return nil, fmt.Errorf("invalid export bundle: missing label")
	}

	return &bundle, nil
}

// ReadBundleFromFile reads an ExportBundle from a file.
func ReadBundleFromFile(path string) (*ExportBundle, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open export file %s: %w", path, err)
	}
	defer f.Close()

	return ReadBundle(f)
}
