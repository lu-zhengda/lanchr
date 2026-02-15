//go:build darwin

package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

// Domain represents the launchd domain a service belongs to.
type Domain int

const (
	DomainUser   Domain = iota // ~/Library/LaunchAgents
	DomainGlobal               // /Library/LaunchAgents, /Library/LaunchDaemons
	DomainSystem               // /System/Library/LaunchAgents, /System/Library/LaunchDaemons
)

// String returns a human-readable name for the domain.
func (d Domain) String() string {
	switch d {
	case DomainUser:
		return "user"
	case DomainGlobal:
		return "global"
	case DomainSystem:
		return "system"
	default:
		return "unknown"
	}
}

// ServiceType distinguishes launch agents from launch daemons.
type ServiceType int

const (
	TypeAgent  ServiceType = iota
	TypeDaemon
)

// String returns "agent" or "daemon".
func (t ServiceType) String() string {
	switch t {
	case TypeAgent:
		return "agent"
	case TypeDaemon:
		return "daemon"
	default:
		return "unknown"
	}
}

// PlistDir describes a directory that contains plist files.
type PlistDir struct {
	Path   string
	Domain Domain
	Type   ServiceType
}

// CurrentUID returns the effective user ID.
func CurrentUID() int {
	return os.Getuid()
}

// GUIDomainTarget returns the GUI domain target for the current user.
func GUIDomainTarget() string {
	return fmt.Sprintf("gui/%d", CurrentUID())
}

// UserDomainTarget returns the user domain target for the current user.
func UserDomainTarget() string {
	return fmt.Sprintf("user/%d", CurrentUID())
}

// ServiceTarget builds the launchctl service target string for a given domain and label.
func ServiceTarget(domain Domain, label string) string {
	switch domain {
	case DomainUser, DomainGlobal:
		return fmt.Sprintf("gui/%d/%s", CurrentUID(), label)
	case DomainSystem:
		return fmt.Sprintf("system/%s", label)
	default:
		return fmt.Sprintf("gui/%d/%s", CurrentUID(), label)
	}
}

// DomainTarget returns the domain-level target (without a service label).
func DomainTarget(domain Domain) string {
	switch domain {
	case DomainUser, DomainGlobal:
		return GUIDomainTarget()
	case DomainSystem:
		return "system"
	default:
		return GUIDomainTarget()
	}
}

// PlistDirectories returns all known plist directories on macOS.
func PlistDirectories() []PlistDir {
	home, _ := os.UserHomeDir()
	return []PlistDir{
		{Path: filepath.Join(home, "Library", "LaunchAgents"), Domain: DomainUser, Type: TypeAgent},
		{Path: "/Library/LaunchAgents", Domain: DomainGlobal, Type: TypeAgent},
		{Path: "/Library/LaunchDaemons", Domain: DomainGlobal, Type: TypeDaemon},
		{Path: "/System/Library/LaunchAgents", Domain: DomainSystem, Type: TypeAgent},
		{Path: "/System/Library/LaunchDaemons", Domain: DomainSystem, Type: TypeDaemon},
	}
}

// IsSIPProtected returns true if the path is under /System/Library.
func IsSIPProtected(path string) bool {
	return len(path) > 15 && path[:15] == "/System/Library"
}

// DomainFromPath determines the domain from a plist file path.
func DomainFromPath(path string) Domain {
	home, _ := os.UserHomeDir()
	userAgents := filepath.Join(home, "Library", "LaunchAgents")

	switch {
	case len(path) >= len(userAgents) && path[:len(userAgents)] == userAgents:
		return DomainUser
	case len(path) >= 23 && path[:23] == "/Library/LaunchAgents/":
		return DomainGlobal
	case len(path) >= 24 && path[:24] == "/Library/LaunchDaemons/":
		return DomainGlobal
	case len(path) >= 30 && path[:30] == "/System/Library/LaunchAgents/":
		return DomainSystem
	case len(path) >= 31 && path[:31] == "/System/Library/LaunchDaemons/":
		return DomainSystem
	default:
		return DomainUser
	}
}

// TypeFromPath determines whether a plist is an agent or daemon from its path.
func TypeFromPath(path string) ServiceType {
	if len(path) >= 24 && path[:24] == "/Library/LaunchDaemons/" {
		return TypeDaemon
	}
	if len(path) >= 31 && path[:31] == "/System/Library/LaunchDaemons/" {
		return TypeDaemon
	}
	return TypeAgent
}

// CheckDarwin is a no-op on macOS. It exists so callers can verify the platform.
func CheckDarwin() error {
	return nil
}
