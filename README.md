# lanchr

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform: macOS](https://img.shields.io/badge/Platform-macOS-lightgrey.svg)](https://github.com/lu-zhengda/lanchr)
[![Homebrew](https://img.shields.io/badge/Homebrew-lu--zhengda/tap-orange.svg)](https://github.com/lu-zhengda/homebrew-tap)

macOS launch agent & daemon manager — inspect, create, and troubleshoot launchd services from the terminal.

## Install

```bash
brew tap lu-zhengda/tap
brew install lanchr
```

## Usage

```
$ lanchr doctor
DOCTOR REPORT
=============

CRITICAL (6)
  [!] com.apple.cvmsCompAgent_arm64_1: binary not found at /System/Library/...
      Suggestion: Remove or update the plist to point to a valid binary
  [!] com.apple.menuextra.battery.helper: binary not found at /System/Library/...
      Suggestion: Remove or update the plist to point to a valid binary
  [!] com.apple.knowledgeconstructiond: last exit status: -9
      Suggestion: Check logs for the service to diagnose the crash

WARNING (71)
  [~] com.apple.akd: duplicate label found in 2 plists
      Suggestion: Remove duplicate plists or use unique labels

Run 'lanchr list' to see all services.

$ lanchr list --no-apple -s running
STATUS  PID     LABEL                                       DOMAIN    BINARY
  *     27280   com.openssh.ssh-agent                       system    /usr/bin/ssh-agent
  *     2022    application.com.dwarvesv.minimalbar...      user
  *     634     application.com.google.Chrome...            user
  *     76742   application.com.raycast.macos...            user
```

## Commands

| Command | Description | Example |
|---------|-------------|---------|
| `list` | List all services | `lanchr list --no-apple` |
| `list -d <domain>` | Filter by domain (user/global/system) | `lanchr list -d user` |
| `list -s <status>` | Filter by status (running/stopped/error) | `lanchr list -s error` |
| `info <label>` | Detailed service info (all plist keys + runtime) | `lanchr info com.example.myapp` |
| `search <query>` | Search by label, path, or content | `lanchr search redis` |
| `enable <label>` | Enable a disabled service (persists) | `lanchr enable com.example.myapp` |
| `disable <label>` | Disable a service (persists) | `lanchr disable com.example.myapp` |
| `load <path>` | Bootstrap a plist file | `lanchr load ~/Library/LaunchAgents/com.example.plist` |
| `unload <label>` | Bootout a service | `lanchr unload com.example.myapp` |
| `restart <label>` | Force restart a running service | `lanchr restart com.example.myapp` |
| `logs <label>` | View service logs | `lanchr logs com.example.myapp -f` |
| `doctor` | Diagnose broken plists and orphaned agents | `lanchr doctor` |
| `create` | Scaffold a new plist from template | See below |
| `edit <label>` | Open plist in $EDITOR | `lanchr edit com.example.myapp` |

### Creating Launch Agents

Use `lanchr create` with templates instead of writing plist XML manually:

```bash
# Simple agent that runs a script at login
lanchr create -l com.me.backup -p /usr/local/bin/backup.sh --template simple --run-at-load

# Interval-based agent (every 30 minutes)
lanchr create -l com.me.sync -p /usr/local/bin/sync.sh --template interval --interval 1800

# Calendar-based agent (daily at 9am)
lanchr create -l com.me.report -p /usr/local/bin/report.sh --template calendar --calendar "Hour=9 Minute=0"

# Keep-alive agent (restart on crash)
lanchr create -l com.me.server -p /usr/local/bin/server --template keepalive --keep-alive

# File watcher agent
lanchr create -l com.me.watcher -p /usr/local/bin/process.sh --template watcher
```

Additional flags: `--stdout <path>`, `--stderr <path>`, `--env KEY=VAL`, `--load` (bootstrap after creation).

## Diagnostic Workflow

1. `lanchr doctor` — identify broken plists, orphaned agents, missing binaries
2. `lanchr list -s error` — find services in error state
3. `lanchr logs <label> -f` — follow logs for a specific service
4. `lanchr restart <label>` — restart a misbehaving service

## TUI

Launch `lanchr` without arguments for interactive service management. Browse services, filter by domain/status, and manage agents with a keyboard-driven interface.

## License

[MIT](LICENSE)
