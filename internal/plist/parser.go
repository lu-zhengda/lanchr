package plist

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	goplist "howett.net/plist"
)

// Parser reads plist files from disk and returns structured data.
type Parser struct{}

// NewParser creates a new plist parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse reads a plist file (XML, binary, or OpenStep format) and returns
// a LaunchAgentPlist struct.
func (p *Parser) Parse(path string) (*LaunchAgentPlist, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plist %s: %w", path, err)
	}
	defer f.Close()

	var result LaunchAgentPlist
	decoder := goplist.NewDecoder(f)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode plist %s: %w", path, err)
	}

	return &result, nil
}

// ParseAll reads all *.plist files from a directory.
// Errors for individual files are collected rather than failing on the first error.
func (p *Parser) ParseAll(dir string) ([]*LaunchAgentPlist, []error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to read directory %s: %w", dir, err)}
	}

	var plists []*LaunchAgentPlist
	var errs []error

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".plist") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		pl, err := p.Parse(path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		plists = append(plists, pl)
	}

	return plists, errs
}
