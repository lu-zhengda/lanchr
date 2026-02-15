# daemons - macOS Launch Agent/Daemon Manager

A CLI + TUI tool for managing macOS launch agents and daemons across all domains.

---

## Table of Contents

1. [Overview](#overview)
2. [CLI Commands](#cli-commands)
3. [Package Structure](#package-structure)
4. [Data Sources & macOS Commands](#data-sources--macos-commands)
5. [Core Domain Model](#core-domain-model)
6. [Package Design](#package-design)
7. [TUI Design](#tui-design)
8. [Plist Templates](#plist-templates)
9. [Error Handling & Edge Cases](#error-handling--edge-cases)

---

## Overview

`daemons` provides a unified interface for inspecting, managing, and troubleshooting
macOS launch agents and daemons. It replaces scattered `launchctl` invocations with
a coherent CLI and an interactive TUI dashboard.

Target audience: developers and system administrators who manage custom agents,
debug third-party daemons, or want visibility into what runs on their Mac.

---

## CLI Commands

```
daemons                        # launch TUI dashboard (default)
daemons list                   # list all agents/daemons
daemons info <label>           # detailed info for a specific service
daemons search <query>         # search by label, binary path, or plist content
daemons enable <label>         # enable a disabled service
daemons disable <label>        # disable a service (without unloading the plist)
daemons logs <label>           # tail stdout/stderr logs for a service
daemons doctor                 # diagnose broken plists, orphaned agents, missing binaries
daemons create                 # scaffold a new launch agent plist from templates
daemons edit <label>           # open the plist in $EDITOR with validation on save
daemons restart <label>        # kickstart (stop + start) a service
daemons load <path>            # bootstrap a plist into the appropriate domain
daemons unload <label>         # bootout a service from its domain
```

### Command Details

#### `daemons list`

```
daemons list [flags]

Flags:
  -d, --domain <domain>   Filter by domain: user, global, system (default: all)
  -s, --status <status>   Filter by status: running, stopped, error
  -t, --type <type>       Filter by type: agent, daemon
      --json              Output as JSON
      --no-apple          Hide com.apple.* services
```

Output columns: `STATUS | PID | LABEL | DOMAIN | BINARY`

Example:
```
STATUS  PID    LABEL                              DOMAIN   BINARY
  *     647    com.apple.Finder                   user     /System/Library/CoreServices/Finder.app/...
  *     1404   com.apple.cloudphotod              user     /usr/libexec/cloudphotod
  -     -      com.docker.helper                  user     /Applications/Docker.app/...
  !     -      com.example.broken                 user     /usr/local/bin/missing-binary
```

Status indicators:
- `*` (green) - running
- `-` (dim) - stopped/idle
- `!` (red) - error (missing binary, bad plist, non-zero last exit)

#### `daemons info <label>`

Displays all parsed plist keys plus live runtime state from `launchctl print`.

```
daemons info com.apple.Finder

Label:               com.apple.Finder
Domain:              gui/501 (user)
Type:                LaunchAgent
State:               running
PID:                 647
Plist Path:          /System/Library/LaunchAgents/com.apple.Finder.plist
Program:             /System/Library/CoreServices/Finder.app/Contents/MacOS/Finder
Arguments:           (none)
Run At Load:         false
Keep Alive:          successful-exit = false
Start Interval:      (none)
Calendar Interval:   (none)
Watch Paths:         (none)
Working Directory:   (none)
Stdout Path:         (none)
Stderr Path:         (none)
Environment:         OSLogRateLimit=64
Exit Timeout:        5s
Last Exit Code:      (never exited)
Runs:                1
Blame:               non-ipc demand
```

#### `daemons search <query>`

Searches across label, program path, program arguments, and plist filename.
Supports glob patterns and regex.

```
daemons search docker
daemons search --regex "com\\.google\\..*"
daemons search --path "/usr/local/bin/*"
```

#### `daemons enable <label>` / `daemons disable <label>`

```
daemons enable com.example.myagent
daemons disable com.example.myagent

Flags:
  -d, --domain <domain>   Target domain (default: auto-detect from plist location)
```

These use `launchctl enable/disable` under the hood. They do NOT delete the plist
file. The disabled state persists across reboots via launchd's internal override
database.

#### `daemons logs <label>`

```
daemons logs com.example.myagent

Flags:
  -f, --follow            Follow log output (like tail -f)
  -n, --lines <n>         Number of lines to show (default: 50)
      --stderr            Show only stderr
      --stdout            Show only stdout
      --unified            Use macOS unified logging (log show --predicate)
```

Strategy:
1. Parse plist for `StandardOutPath` and `StandardErrorPath` -- tail those files.
2. If no explicit paths, fall back to unified logging:
   `log show --predicate 'process == "<binary-name>"' --last 1h`

#### `daemons doctor`

Runs a suite of health checks and prints a diagnostic report.

```
daemons doctor

Flags:
      --fix               Attempt to auto-fix issues (e.g., fix permissions)
      --domain <domain>   Check only a specific domain
```

Checks performed:
1. **Missing binaries** -- Program path in plist does not exist on disk.
2. **Bad permissions** -- Plist not owned by root (for /Library) or current user
   (for ~/Library). World-writable plists.
3. **Invalid XML** -- Plist fails to parse (malformed XML/binary plist).
4. **Orphaned overrides** -- Services in `launchctl print-disabled` that have no
   matching plist on disk.
5. **Duplicate labels** -- Multiple plists across domains with the same Label.
6. **Missing Label key** -- Plist missing the required Label key.
7. **Filename mismatch** -- Plist filename does not match its Label key
   (convention violation, not fatal).
8. **Crashed services** -- Services with non-zero last exit status.
9. **Stale log files** -- StandardOutPath/StandardErrorPath pointing to
   non-existent directories.
10. **SIP-protected but modified** -- Plists in /System/Library that appear
    to have been tampered with (informational only).

Output format:
```
DOCTOR REPORT
=============

CRITICAL (2)
  [!] com.example.myagent: binary not found at /usr/local/bin/myagent
  [!] com.example.broken: plist parse error: unexpected EOF

WARNING (3)
  [~] com.example.task: filename "task.plist" does not match label "com.example.task"
  [~] com.google.keystone.agent: world-writable plist
  [~] com.old.service: override exists but no plist found on disk

OK (145)
  All other services passed checks.
```

#### `daemons create`

Interactive (or flag-driven) scaffolding for new launch agent plists.

```
daemons create

Flags:
  -l, --label <label>         Service label (e.g., com.example.myagent)
  -p, --program <path>        Executable path
  -a, --args <args>           Program arguments (comma-separated)
      --interval <seconds>    StartInterval
      --calendar <spec>       StartCalendarInterval (cron-like: "0 * * * *")
      --run-at-load           Set RunAtLoad to true
      --keep-alive            Set KeepAlive to true
      --stdout <path>         StandardOutPath
      --stderr <path>         StandardErrorPath
      --working-dir <path>    WorkingDirectory
      --env <KEY=VAL>         Environment variables (repeatable)
      --template <name>       Use a built-in template: simple, interval, calendar,
                              keepalive, watcher
  -o, --output <path>         Output path (default: ~/Library/LaunchAgents/<label>.plist)
      --load                  Bootstrap the plist after creation
```

When run without flags, enters an interactive wizard in the terminal.

#### `daemons edit <label>`

Opens the plist in `$EDITOR` (or `$VISUAL`). On save, validates the plist
with `plutil -lint` and optionally reloads the service.

```
daemons edit com.example.myagent

Flags:
      --reload    Bootout and bootstrap the service after editing
```

#### `daemons restart <label>`

```
daemons restart com.example.myagent
```

Equivalent to `launchctl kickstart -k <service-target>`.

#### `daemons load <path>` / `daemons unload <label>`

```
daemons load ~/Library/LaunchAgents/com.example.myagent.plist
daemons unload com.example.myagent
```

Wrappers around `launchctl bootstrap` and `launchctl bootout` with
auto-detection of the correct domain target.

---

## Package Structure

```
daemons/
  cmd/
    daemons/
      main.go                    # Entry point, calls cli.Execute()
  internal/
    cli/                         # Cobra command definitions
      root.go                    # Root command, launches TUI when no subcommand
      list.go                    # "list" subcommand
      info.go                    # "info" subcommand
      search.go                  # "search" subcommand
      enable.go                  # "enable" subcommand
      disable.go                 # "disable" subcommand
      logs.go                    # "logs" subcommand
      doctor.go                  # "doctor" subcommand
      create.go                  # "create" subcommand
      edit.go                    # "edit" subcommand
      restart.go                 # "restart" subcommand
      load.go                    # "load" subcommand
      unload.go                  # "unload" subcommand
    agent/                       # Core domain: service representation + operations
      service.go                 # Service struct, Domain enum, Status enum
      manager.go                 # ServiceManager: enable, disable, restart, load, unload
      scanner.go                 # Scan all plist directories, correlate with launchctl state
      doctor.go                  # Health check logic
    plist/                       # Plist parsing and generation
      parser.go                  # Parse XML/binary plists into structured Go types
      writer.go                  # Generate plist XML from Go structs
      template.go                # Built-in plist templates for "create"
      types.go                   # LaunchAgentPlist struct with all known keys
    launchctl/                   # Shell interface to launchctl
      executor.go                # Run launchctl commands, parse output
      list.go                    # Parse "launchctl list" output
      print.go                   # Parse "launchctl print" output
      disabled.go                # Parse "launchctl print-disabled" output
    logs/                        # Log file access
      tailer.go                  # Tail StandardOutPath/StandardErrorPath files
      unified.go                 # Query macOS unified logging via "log" command
    tui/                         # Bubbletea TUI
      app.go                     # Main Bubbletea model, top-level Update/View
      list.go                    # Service list view with filtering/sorting
      detail.go                  # Service detail panel (info view)
      logs.go                    # Inline log viewer
      doctor.go                  # Doctor report view
      search.go                  # Search overlay / filter bar
      help.go                    # Help overlay with keybindings
      styles.go                  # Lipgloss styles and theme
      keys.go                    # Key bindings
    platform/                    # Platform detection and guards
      darwin.go                  # macOS-specific checks, UID resolution
      notdarwin.go               # Build-tag guarded: error on non-macOS
  go.mod
  go.sum
```

---

## Data Sources & macOS Commands

### Plist File Locations

| Directory | Domain | Type | Privileges |
|-----------|--------|------|------------|
| `~/Library/LaunchAgents/` | Per-user agents | Agent | User |
| `/Library/LaunchAgents/` | Global agents (all users) | Agent | Root to install, user to load |
| `/Library/LaunchDaemons/` | Global daemons | Daemon | Root |
| `/System/Library/LaunchAgents/` | Apple system agents | Agent | SIP-protected, read-only |
| `/System/Library/LaunchDaemons/` | Apple system daemons | Daemon | SIP-protected, read-only |

### launchctl Commands Used

#### Listing and Querying

```bash
# List all services in the current user's GUI domain (modern API)
launchctl print gui/$(id -u)
# Output: domain metadata + list of services with PID, status, label

# List all services (legacy, but reliable for PID + exit status)
launchctl list
# Output: PID \t Status \t Label (tab-separated, one per line)

# Detailed info about a specific service
launchctl print gui/<uid>/<label>
launchctl print system/<label>
# Output: multi-line structured dump of service properties

# List disabled services in a domain
launchctl print-disabled gui/<uid>
launchctl print-disabled system
# Output: dictionary of label => enabled/disabled

# Why was a service launched?
launchctl blame gui/<uid>/<label>
launchctl blame system/<label>
# Output: single line, e.g., "non-ipc demand", "timer", "keepalive"
```

#### Service Lifecycle

```bash
# Enable a disabled service (persists across reboot)
launchctl enable gui/<uid>/<label>
launchctl enable system/<label>

# Disable a service (persists across reboot, does NOT unload)
launchctl disable gui/<uid>/<label>
launchctl disable system/<label>

# Load (bootstrap) a plist into a domain
launchctl bootstrap gui/<uid> /path/to/plist
launchctl bootstrap system /path/to/plist

# Unload (bootout) a service from a domain
launchctl bootout gui/<uid>/<label>
launchctl bootout system/<label>
# Alternative: bootout by plist path
launchctl bootout gui/<uid> /path/to/plist

# Force restart a running service
launchctl kickstart -k gui/<uid>/<label>
launchctl kickstart -kp gui/<uid>/<label>  # also print PID

# Send a signal to a running service
launchctl kill SIGTERM gui/<uid>/<label>
```

#### Diagnostics

```bash
# Validate a plist file
plutil -lint /path/to/plist

# Convert plist to readable format (for binary plists)
plutil -p /path/to/plist

# Convert binary plist to XML
plutil -convert xml1 -o /tmp/readable.plist /path/to/binary.plist

# Query unified logging for a process
log show --predicate 'process == "myagent"' --last 1h --style compact
log stream --predicate 'process == "myagent"' --style compact
```

#### Domain Target Resolution

The tool must resolve the correct domain target for each service:

```
Plist in ~/Library/LaunchAgents/       => gui/<uid>/<label>
Plist in /Library/LaunchAgents/        => gui/<uid>/<label>  (user context)
Plist in /Library/LaunchDaemons/       => system/<label>
Plist in /System/Library/LaunchAgents/ => gui/<uid>/<label>  (user context)
Plist in /System/Library/LaunchDaemons/ => system/<label>
```

For `enable`/`disable`, the domain target differs slightly:
- User agents: `gui/<uid>/<label>` or `user/<uid>/<label>`
- System daemons: `system/<label>`

### Plist Parsing

macOS plists come in three formats:
1. **XML** -- parse with `encoding/xml` or Apple's plist format
2. **Binary** -- convert first with `plutil -convert xml1`, or use a Go plist library
3. **JSON** -- rare but valid; parse with `encoding/json`

Recommended Go library: `howett.net/plist` -- handles all three formats natively
without shelling out to `plutil`.

---

## Core Domain Model

### Service

```go
// internal/agent/service.go

type Domain int

const (
    DomainUser   Domain = iota  // ~/Library/LaunchAgents
    DomainGlobal                // /Library/LaunchAgents, /Library/LaunchDaemons
    DomainSystem                // /System/Library/LaunchAgents, /System/Library/LaunchDaemons
)

type ServiceType int

const (
    TypeAgent  ServiceType = iota
    TypeDaemon
)

type Status int

const (
    StatusStopped Status = iota
    StatusRunning
    StatusError
    StatusDisabled
)

type Service struct {
    Label           string
    Domain          Domain
    Type            ServiceType
    Status          Status
    PID             int          // -1 if not running
    LastExitStatus  int
    PlistPath       string
    Program         string
    ProgramArgs     []string
    RunAtLoad       bool
    KeepAlive       interface{}  // bool or KeepAliveConditions
    StartInterval   int
    CalendarInterval []CalendarInterval
    WatchPaths      []string
    QueueDirectories []string
    StandardOutPath string
    StandardErrorPath string
    WorkingDirectory string
    EnvironmentVars map[string]string
    UserName        string
    GroupName       string
    Disabled        bool
    ExitTimeout     int
    ThrottleInterval int
    Nice            int
    ProcessType     string
    MachServices    map[string]interface{}
    Sockets         map[string]interface{}
    BlameLine       string       // output of launchctl blame
}

type CalendarInterval struct {
    Minute  *int
    Hour    *int
    Day     *int
    Weekday *int
    Month   *int
}

type KeepAliveConditions struct {
    SuccessfulExit *bool
    Crashed        *bool
    PathState      map[string]bool
    OtherJobEnabled map[string]bool
}
```

### LaunchAgentPlist

```go
// internal/plist/types.go

type LaunchAgentPlist struct {
    Label                string                 `plist:"Label"`
    Disabled             bool                   `plist:"Disabled,omitempty"`
    Program              string                 `plist:"Program,omitempty"`
    ProgramArguments     []string               `plist:"ProgramArguments,omitempty"`
    EnableGlobbing       bool                   `plist:"EnableGlobbing,omitempty"`
    EnvironmentVariables map[string]string      `plist:"EnvironmentVariables,omitempty"`
    WorkingDirectory     string                 `plist:"WorkingDirectory,omitempty"`
    StandardOutPath      string                 `plist:"StandardOutPath,omitempty"`
    StandardErrorPath    string                 `plist:"StandardErrorPath,omitempty"`
    StandardInPath       string                 `plist:"StandardInPath,omitempty"`
    RunAtLoad            bool                   `plist:"RunAtLoad,omitempty"`
    KeepAlive            interface{}            `plist:"KeepAlive,omitempty"`
    StartInterval        int                    `plist:"StartInterval,omitempty"`
    StartCalendarInterval interface{}           `plist:"StartCalendarInterval,omitempty"`
    StartOnMount         bool                   `plist:"StartOnMount,omitempty"`
    WatchPaths           []string               `plist:"WatchPaths,omitempty"`
    QueueDirectories     []string               `plist:"QueueDirectories,omitempty"`
    UserName             string                 `plist:"UserName,omitempty"`
    GroupName            string                 `plist:"GroupName,omitempty"`
    Umask                interface{}            `plist:"Umask,omitempty"`
    RootDirectory        string                 `plist:"RootDirectory,omitempty"`
    ExitTimeOut          int                    `plist:"ExitTimeOut,omitempty"`
    ThrottleInterval     int                    `plist:"ThrottleInterval,omitempty"`
    InitGroups           *bool                  `plist:"InitGroups,omitempty"`
    Nice                 int                    `plist:"Nice,omitempty"`
    ProcessType          string                 `plist:"ProcessType,omitempty"`
    AbandonProcessGroup  bool                   `plist:"AbandonProcessGroup,omitempty"`
    LowPriorityIO        bool                   `plist:"LowPriorityIO,omitempty"`
    LowPriorityBackgroundIO bool               `plist:"LowPriorityBackgroundIO,omitempty"`
    LaunchOnlyOnce       bool                   `plist:"LaunchOnlyOnce,omitempty"`
    MachServices         map[string]interface{} `plist:"MachServices,omitempty"`
    Sockets              map[string]interface{} `plist:"Sockets,omitempty"`
    LaunchEvents         map[string]interface{} `plist:"LaunchEvents,omitempty"`
    HardResourceLimits   map[string]int         `plist:"HardResourceLimits,omitempty"`
    SoftResourceLimits   map[string]int         `plist:"SoftResourceLimits,omitempty"`
    EnableTransactions   bool                   `plist:"EnableTransactions,omitempty"`
    EnablePressuredExit  bool                   `plist:"EnablePressuredExit,omitempty"`
    Debug                bool                   `plist:"Debug,omitempty"`
    WaitForDebugger      bool                   `plist:"WaitForDebugger,omitempty"`
    LimitLoadToSessionType interface{}          `plist:"LimitLoadToSessionType,omitempty"`
    LimitLoadToHardware  map[string][]string    `plist:"LimitLoadToHardware,omitempty"`
    LimitLoadFromHardware map[string][]string   `plist:"LimitLoadFromHardware,omitempty"`
    inetdCompatibility   map[string]bool        `plist:"inetdCompatibility,omitempty"`
    AssociatedBundleIdentifiers interface{}     `plist:"AssociatedBundleIdentifiers,omitempty"`
}
```

---

## Package Design

### internal/launchctl (Shell Interface)

This package is the ONLY package that shells out to `launchctl` and `plutil`.
All other packages use it as an abstraction boundary. This makes testing feasible
via interface injection.

```go
// internal/launchctl/executor.go

// Executor runs launchctl commands and returns raw output.
type Executor interface {
    // List returns parsed output of "launchctl list".
    List() ([]ListEntry, error)

    // PrintDomain returns parsed output of "launchctl print <domain-target>".
    PrintDomain(domainTarget string) (*DomainInfo, error)

    // PrintService returns parsed output of "launchctl print <service-target>".
    PrintService(serviceTarget string) (*ServiceInfo, error)

    // PrintDisabled returns the disabled services map for a domain.
    PrintDisabled(domainTarget string) (map[string]bool, error)

    // Blame returns the reason a service was launched.
    Blame(serviceTarget string) (string, error)

    // Enable enables a service.
    Enable(serviceTarget string) error

    // Disable disables a service.
    Disable(serviceTarget string) error

    // Bootstrap loads a plist into a domain.
    Bootstrap(domainTarget string, plistPath string) error

    // Bootout removes a service from a domain.
    Bootout(serviceTarget string) error

    // Kickstart restarts a service.
    Kickstart(serviceTarget string, kill bool) error

    // Kill sends a signal to a service.
    Kill(signal string, serviceTarget string) error
}

// ListEntry is a row from "launchctl list".
type ListEntry struct {
    PID    int    // -1 if not running
    Status int    // last exit status
    Label  string
}

// DefaultExecutor shells out to /bin/launchctl.
type DefaultExecutor struct{}
```

Parsing `launchctl list` output:

```go
// internal/launchctl/list.go

func (e *DefaultExecutor) List() ([]ListEntry, error) {
    // Runs: launchctl list
    // Parses tab-separated output: PID \t Status \t Label
    // First line is header, skip it
    // PID of "-" means not running (set to -1)
}
```

Parsing `launchctl print` output:

```go
// internal/launchctl/print.go

func (e *DefaultExecutor) PrintService(serviceTarget string) (*ServiceInfo, error) {
    // Runs: launchctl print gui/501/com.example.myagent
    // Parses key = value pairs from the structured output
    // Extracts: state, pid, path, program, runs, last exit code, etc.
    // NOTE: Apple says this output is NOT API and may change between releases.
    // We parse defensively, treating missing fields as optional.
}

type ServiceInfo struct {
    State          string // "running", "waiting", etc.
    PID            int
    Path           string // plist path
    BundleID       string
    Program        string
    Type           string // "LaunchAgent", "LaunchDaemon"
    Runs           int
    LastExitCode   string
    ExitTimeout    int
    Domain         string
}
```

### internal/plist (Plist Parsing & Generation)

```go
// internal/plist/parser.go

// Parser reads plist files from disk and returns structured data.
type Parser struct{}

// Parse reads a plist file (XML, binary, or JSON format) and returns
// a LaunchAgentPlist struct.
func (p *Parser) Parse(path string) (*LaunchAgentPlist, error) {
    // Uses howett.net/plist to handle all formats.
    // Falls back to shelling out to plutil -convert xml1 if the library fails.
}

// ParseAll reads all plists from a directory.
func (p *Parser) ParseAll(dir string) ([]*LaunchAgentPlist, error) {
    // Lists *.plist files in dir, calls Parse on each.
    // Collects errors per file rather than failing on first error.
}
```

```go
// internal/plist/writer.go

// Writer generates plist XML files.
type Writer struct{}

// Write serializes a LaunchAgentPlist to XML format at the given path.
func (w *Writer) Write(plist *LaunchAgentPlist, path string) error {
    // Uses howett.net/plist to encode as XML.
    // Validates required fields (Label, Program or ProgramArguments).
    // Sets file permissions to 0644.
}

// Validate checks a plist for common issues.
func (w *Writer) Validate(plist *LaunchAgentPlist) []ValidationError {
    // Checks: Label present, Program or ProgramArguments present,
    // binary exists on disk, no duplicate keys, etc.
}
```

```go
// internal/plist/template.go

type Template struct {
    Name        string
    Description string
    Plist       LaunchAgentPlist
}

func BuiltinTemplates() []Template {
    return []Template{
        {Name: "simple", Description: "Run once at login", ...},
        {Name: "interval", Description: "Run every N seconds", ...},
        {Name: "calendar", Description: "Run on a cron-like schedule", ...},
        {Name: "keepalive", Description: "Always running, restart on crash", ...},
        {Name: "watcher", Description: "Run when watched paths change", ...},
    }
}
```

### internal/agent (Core Business Logic)

```go
// internal/agent/scanner.go

// Scanner discovers all launch agents and daemons across all domains
// and correlates plist data with live launchctl state.
type Scanner struct {
    plistParser *plist.Parser
    launchctl   launchctl.Executor
}

// ScanAll returns all services found across all plist directories,
// enriched with live runtime state from launchctl.
func (s *Scanner) ScanAll() ([]Service, error) {
    // 1. Scan all plist directories:
    //    ~/Library/LaunchAgents      => DomainUser,   TypeAgent
    //    /Library/LaunchAgents       => DomainGlobal,  TypeAgent
    //    /Library/LaunchDaemons      => DomainGlobal,  TypeDaemon
    //    /System/Library/LaunchAgents     => DomainSystem, TypeAgent
    //    /System/Library/LaunchDaemons    => DomainSystem, TypeDaemon
    //
    // 2. Parse each plist with plist.Parser
    //
    // 3. Call launchctl.List() for PID + exit status
    //
    // 4. Call launchctl.PrintDisabled("gui/<uid>") and
    //    launchctl.PrintDisabled("system") for disabled state
    //
    // 5. Correlate: match plist Label to launchctl list entries
    //
    // 6. For services found in launchctl but with no plist on disk,
    //    create a Service with a note that it may be an XPC service
    //    or dynamically registered.
}

// ScanDomain returns services from a specific domain only.
func (s *Scanner) ScanDomain(domain Domain) ([]Service, error)
```

```go
// internal/agent/manager.go

// Manager performs lifecycle operations on services.
type Manager struct {
    launchctl launchctl.Executor
    scanner   *Scanner
}

func (m *Manager) Enable(label string) error {
    // 1. Find the service by label (via scanner)
    // 2. Determine the service target: gui/<uid>/<label> or system/<label>
    // 3. Call launchctl.Enable(serviceTarget)
}

func (m *Manager) Disable(label string) error {
    // 1. Find the service by label
    // 2. Determine the service target
    // 3. Call launchctl.Disable(serviceTarget)
}

func (m *Manager) Restart(label string) error {
    // 1. Find the service by label
    // 2. Determine the service target
    // 3. Call launchctl.Kickstart(serviceTarget, kill: true)
}

func (m *Manager) Load(plistPath string) error {
    // 1. Parse the plist to get the Label
    // 2. Determine domain from path:
    //    ~/Library/LaunchAgents => gui/<uid>
    //    /Library/LaunchDaemons => system
    // 3. Call launchctl.Bootstrap(domainTarget, plistPath)
}

func (m *Manager) Unload(label string) error {
    // 1. Find the service by label
    // 2. Determine the service target
    // 3. Call launchctl.Bootout(serviceTarget)
}
```

```go
// internal/agent/doctor.go

type Severity int

const (
    SeverityOK Severity = iota
    SeverityWarning
    SeverityCritical
)

type Finding struct {
    Severity    Severity
    Label       string
    PlistPath   string
    Message     string
    Suggestion  string
}

// Doctor runs health checks across all services.
type Doctor struct {
    scanner *Scanner
    parser  *plist.Parser
}

func (d *Doctor) Check() ([]Finding, error) {
    // Run all checks, collect findings, sort by severity.
}

func (d *Doctor) checkMissingBinaries(services []Service) []Finding
func (d *Doctor) checkPermissions(services []Service) []Finding
func (d *Doctor) checkPlistValidity(dirs []string) []Finding
func (d *Doctor) checkOrphanedOverrides(services []Service, disabled map[string]bool) []Finding
func (d *Doctor) checkDuplicateLabels(services []Service) []Finding
func (d *Doctor) checkFilenameMismatch(services []Service) []Finding
func (d *Doctor) checkCrashedServices(services []Service) []Finding
func (d *Doctor) checkStaleLogPaths(services []Service) []Finding
```

### internal/logs (Log Access)

```go
// internal/logs/tailer.go

// Tailer reads log files specified in plist StandardOutPath/StandardErrorPath.
type Tailer struct{}

// Tail returns the last N lines from the given file path.
func (t *Tailer) Tail(path string, lines int) ([]string, error)

// Follow streams new lines from the given file path to the channel.
// The caller should cancel the context to stop streaming.
func (t *Tailer) Follow(ctx context.Context, path string) (<-chan string, error)
```

```go
// internal/logs/unified.go

// UnifiedLog queries macOS unified logging for a process.
type UnifiedLog struct{}

// Show returns recent log entries for a process.
func (u *UnifiedLog) Show(processName string, duration time.Duration, lines int) ([]LogEntry, error) {
    // Runs: log show --predicate 'process == "<name>"' --last <duration> --style compact
}

// Stream opens a live stream of log entries for a process.
func (u *UnifiedLog) Stream(ctx context.Context, processName string) (<-chan LogEntry, error) {
    // Runs: log stream --predicate 'process == "<name>"' --style compact
}

type LogEntry struct {
    Timestamp time.Time
    Process   string
    PID       int
    Level     string // Default, Info, Debug, Error, Fault
    Message   string
}
```

### internal/platform (Platform Guards)

```go
// internal/platform/darwin.go
//go:build darwin

package platform

func CurrentUID() int {
    return os.Getuid()
}

func GUIDomainTarget() string {
    return fmt.Sprintf("gui/%d", CurrentUID())
}

func UserDomainTarget() string {
    return fmt.Sprintf("user/%d", CurrentUID())
}

func ServiceTarget(domain Domain, label string) string {
    switch domain {
    case DomainUser, DomainGlobal:
        // User and global agents run in the GUI domain
        return fmt.Sprintf("gui/%d/%s", CurrentUID(), label)
    case DomainSystem:
        return fmt.Sprintf("system/%s", label)
    }
}

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
```

```go
// internal/platform/notdarwin.go
//go:build !darwin

package platform

import "errors"

var ErrNotMacOS = errors.New("daemons requires macOS")

// All functions return ErrNotMacOS on non-macOS platforms.
```

---

## TUI Design

The TUI is built with Bubbletea (github.com/charmbracelet/bubbletea) and
Lipgloss (github.com/charmbracelet/lipgloss) for styling.

### Layout

```
+----------------------------------------------------------------------+
| daemons v0.1.0                                         [?] help      |
+----------------------------------------------------------------------+
| [/] Search  [Tab] Switch pane  [d] Domain filter  [s] Status filter  |
+----------------------------------------------------------------------+
|                                                                      |
| USER AGENTS (~/Library/LaunchAgents)                      3 services |
| ---------------------------------------------------------------      |
|   * 12345  com.google.keystone.agent          /usr/local/bin/ks...   |
|   -    -   com.docker.helper                  /Applications/Doc...   |
| > * 67890  com.example.myagent                /usr/local/bin/my...   |
|                                                                      |
| GLOBAL AGENTS (/Library/LaunchAgents)                     0 services |
| ---------------------------------------------------------------      |
|   (none)                                                             |
|                                                                      |
| GLOBAL DAEMONS (/Library/LaunchDaemons)                   3 services |
| ---------------------------------------------------------------      |
|   *  1234  com.docker.vmnetd                  /Library/Privileg...   |
|   *  5678  dev.orbstack.OrbStack.privhelper   /Library/Privileg...   |
|   -    -   com.docker.socket                  /Library/Privileg...   |
|                                                                      |
| SYSTEM AGENTS (/System/Library/LaunchAgents)            145 services |
| ---------------------------------------------------------------      |
|   *   647  com.apple.Finder                   /System/Library/C...   |
|   *   798  com.apple.mediaremoteagent         /usr/libexec/medi...   |
|   ...                                                                |
|                                                                      |
+----------------------------------------------------------------------+
| e:enable  d:disable  r:restart  l:logs  i:info  D:doctor  q:quit    |
+----------------------------------------------------------------------+
```

### Views / Modes

1. **List View** (default) -- Service list grouped by domain, scrollable.
2. **Detail View** -- Full info panel for the selected service (press `i` or `Enter`).
3. **Log View** -- Inline log viewer for the selected service (press `l`).
4. **Doctor View** -- Diagnostic report (press `D`).
5. **Search Overlay** -- Filter bar that narrows the list as you type (press `/`).
6. **Help Overlay** -- Keybinding reference (press `?`).

### Key Bindings

```
Navigation:
  j / Down      Move cursor down
  k / Up        Move cursor up
  g             Jump to top
  G             Jump to bottom
  Tab           Cycle through domain groups
  Enter / i     Open detail view for selected service
  Esc           Go back to list view

Actions:
  e             Enable selected service
  x             Disable selected service
  r             Restart selected service (kickstart -k)
  l             View logs for selected service
  L             Load a plist (prompts for path)
  U             Unload selected service

Filtering:
  /             Open search bar
  d             Cycle domain filter: all -> user -> global -> system
  s             Cycle status filter: all -> running -> stopped -> error
  a             Toggle show/hide Apple services (com.apple.*)

Views:
  D             Run doctor and show report
  ?             Show help overlay
  q / Ctrl+C    Quit
```

### TUI Model Architecture

```go
// internal/tui/app.go

type viewMode int

const (
    viewList viewMode = iota
    viewDetail
    viewLogs
    viewDoctor
)

type Model struct {
    // Data
    services     []agent.Service
    filtered     []agent.Service    // services after applying filters
    scanner      *agent.Scanner
    manager      *agent.Manager

    // View state
    mode         viewMode
    cursor       int
    scroll       int
    width        int
    height       int

    // Sub-models
    listModel    ListModel
    detailModel  DetailModel
    logModel     LogModel
    doctorModel  DoctorModel
    searchModel  SearchModel
    helpModel    HelpModel

    // Filters
    domainFilter *agent.Domain      // nil = all
    statusFilter *agent.Status      // nil = all
    hideApple    bool
    searchQuery  string

    // State
    showSearch   bool
    showHelp     bool
    err          error
}

func (m Model) Init() tea.Cmd {
    // Scan all services on startup
    return m.scanServices
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if m.showSearch {
            return m.searchModel.Update(msg)
        }
        if m.showHelp {
            return m.helpModel.Update(msg)
        }
        switch m.mode {
        case viewList:
            return m.updateList(msg)
        case viewDetail:
            return m.updateDetail(msg)
        case viewLogs:
            return m.updateLogs(msg)
        case viewDoctor:
            return m.updateDoctor(msg)
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    case servicesScanResult:
        m.services = msg.services
        m.applyFilters()
    case doctorResult:
        m.doctorModel.findings = msg.findings
        m.mode = viewDoctor
    }
    return m, nil
}
```

### Styles

```go
// internal/tui/styles.go

var (
    // Status indicators
    StatusRunning  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // green
    StatusStopped  = lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // gray
    StatusError    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red
    StatusDisabled = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // orange

    // Domain headers
    DomainHeader = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("99")).
        MarginTop(1)

    // Selected row
    Selected = lipgloss.NewStyle().
        Background(lipgloss.Color("236")).
        Bold(true)

    // Section count badge
    CountBadge = lipgloss.NewStyle().
        Foreground(lipgloss.Color("245")).
        Italic(true)

    // Help bar at bottom
    HelpBar = lipgloss.NewStyle().
        Foreground(lipgloss.Color("241")).
        Background(lipgloss.Color("235")).
        Padding(0, 1)

    // Title bar
    TitleBar = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("255")).
        Background(lipgloss.Color("62")).
        Padding(0, 1).
        Width(80)

    // Doctor severity colors
    FindingCritical = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
    FindingWarning  = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
    FindingOK       = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)
```

---

## Plist Templates

### simple -- Run once at login

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>Program</key>
    <string>{{.Program}}</string>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/{{.Label}}.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/{{.Label}}.stderr.log</string>
</dict>
</plist>
```

### interval -- Run every N seconds

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>Program</key>
    <string>{{.Program}}</string>
    <key>StartInterval</key>
    <integer>{{.StartInterval}}</integer>
    <key>StandardOutPath</key>
    <string>/tmp/{{.Label}}.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/{{.Label}}.stderr.log</string>
</dict>
</plist>
```

### calendar -- Run on a cron-like schedule

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>Program</key>
    <string>{{.Program}}</string>
    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>{{.Hour}}</integer>
        <key>Minute</key>
        <integer>{{.Minute}}</integer>
    </dict>
    <key>StandardOutPath</key>
    <string>/tmp/{{.Label}}.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/{{.Label}}.stderr.log</string>
</dict>
</plist>
```

### keepalive -- Always running, restart on crash

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>Program</key>
    <string>{{.Program}}</string>
    <key>KeepAlive</key>
    <dict>
        <key>Crashed</key>
        <true/>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/{{.Label}}.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/{{.Label}}.stderr.log</string>
</dict>
</plist>
```

### watcher -- Run when watched paths change

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>Program</key>
    <string>{{.Program}}</string>
    <key>WatchPaths</key>
    <array>
        <string>{{.WatchPath}}</string>
    </array>
    <key>StandardOutPath</key>
    <string>/tmp/{{.Label}}.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/{{.Label}}.stderr.log</string>
</dict>
</plist>
```

---

## Error Handling & Edge Cases

### Permission Escalation

Operations on system daemons (`/Library/LaunchDaemons`, `/System/Library/*`) require
root privileges. The tool should:

1. Detect when an operation requires elevated privileges.
2. Prompt the user and explain why `sudo` is needed.
3. Re-execute the specific `launchctl` command via `sudo`, NOT the entire tool.

```go
func (m *Manager) Enable(label string) error {
    svc, err := m.scanner.FindByLabel(label)
    if err != nil {
        return fmt.Errorf("failed to find service %q: %w", label, err)
    }
    target := platform.ServiceTarget(svc.Domain, svc.Label)
    err = m.launchctl.Enable(target)
    if err != nil && isPermissionDenied(err) {
        return fmt.Errorf("failed to enable %q: operation requires sudo (system daemon): %w", label, err)
    }
    return err
}
```

### SIP-Protected Services

Services in `/System/Library/` are protected by System Integrity Protection.
The tool should:

1. Mark these services as read-only in the UI.
2. Refuse to attempt `enable`/`disable`/`edit` on SIP-protected services.
3. Display a clear error message explaining why.

### Binary Plist Files

Some plists (especially Apple system ones) are in binary format. The tool must
handle all three plist formats transparently:

1. Use `howett.net/plist` which handles XML, binary, and OpenStep formats.
2. Fall back to `plutil -convert xml1` if the Go library cannot parse a file.

### Agents Without Plists

`launchctl list` may return services that have no plist on disk:
- XPC services embedded in app bundles
- Dynamically registered services
- Services from app extensions

The tool should list these but mark them as "no plist" and disable operations
that require plist access (edit, doctor checks).

### Large Number of Services

A typical macOS installation has 400-600 services. The scanner should:

1. Parse plists in parallel using a worker pool (bounded by `runtime.NumCPU()`).
2. Cache scan results and provide a refresh mechanism (key `R` in TUI).
3. Never re-scan automatically; only on user request or tool startup.

### Race Conditions

Service state can change between scanning and performing an action. The tool should:

1. Re-check service state before performing lifecycle operations.
2. Handle "already running", "already stopped", "not found" errors gracefully.
3. Display stale-data warnings in the TUI status bar.

---

## Dependencies

```
github.com/spf13/cobra          # CLI framework
github.com/charmbracelet/bubbletea  # TUI framework
github.com/charmbracelet/lipgloss   # TUI styling
github.com/charmbracelet/bubbles    # TUI components (textinput, viewport, spinner, table)
howett.net/plist                 # Plist parsing (XML, binary, OpenStep)
```

No other external dependencies. The tool shells out to `launchctl`, `plutil`,
and `log` (for unified logging) -- all of which are guaranteed present on macOS.

---

## Build & Run

```bash
# Build
go build -o daemons ./cmd/daemons

# Run TUI
./daemons

# Run CLI commands
./daemons list --no-apple
./daemons info com.example.myagent
./daemons doctor
./daemons create --label com.example.task --program /usr/local/bin/task --interval 300
```

Build tags ensure the tool only compiles on macOS:

```go
// cmd/daemons/main.go
//go:build darwin
```
