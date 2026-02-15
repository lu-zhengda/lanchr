//go:build !darwin

package platform

import "errors"

// ErrNotMacOS is returned on non-macOS platforms.
var ErrNotMacOS = errors.New("lanchr requires macOS")

// Domain represents the launchd domain a service belongs to.
type Domain int

const (
	DomainUser   Domain = iota
	DomainGlobal
	DomainSystem
)

func (d Domain) String() string { return "unsupported" }

// ServiceType distinguishes launch agents from launch daemons.
type ServiceType int

const (
	TypeAgent  ServiceType = iota
	TypeDaemon
)

func (t ServiceType) String() string { return "unsupported" }

// PlistDir describes a directory that contains plist files.
type PlistDir struct {
	Path   string
	Domain Domain
	Type   ServiceType
}

func CurrentUID() int                      { return -1 }
func GUIDomainTarget() string              { return "" }
func UserDomainTarget() string             { return "" }
func ServiceTarget(_ Domain, _ string) string { return "" }
func DomainTarget(_ Domain) string         { return "" }
func PlistDirectories() []PlistDir         { return nil }
func IsSIPProtected(_ string) bool         { return false }
func DomainFromPath(_ string) Domain       { return DomainUser }
func TypeFromPath(_ string) ServiceType    { return TypeAgent }

// CheckDarwin returns an error on non-macOS platforms.
func CheckDarwin() error {
	return ErrNotMacOS
}
