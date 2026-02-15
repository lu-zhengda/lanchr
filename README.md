# lanchr

A macOS launch agent/daemon manager — inspect, manage, and troubleshoot launchd services.

## Install

```bash
brew tap lu-zhengda/tap
brew install lanchr
```

## Quick Start

```bash
lanchr           # Launch interactive TUI
lanchr --help    # Show all commands
```

## Commands

| Command   | Description                              |
|-----------|------------------------------------------|
| `list`    | List all agents and daemons              |
| `info`    | Show detailed service information        |
| `search`  | Search services by name/label            |
| `enable`  | Enable a service                         |
| `disable` | Disable a service                        |
| `logs`    | View service logs                        |
| `doctor`  | Diagnose service issues                  |
| `create`  | Create a new launch agent/daemon         |
| `edit`    | Edit an existing service                 |
| `restart` | Restart a service                        |
| `load`    | Load a service                           |
| `unload`  | Unload a service                         |

## TUI

Launch without arguments for interactive mode. Browse services, filter by domain/status, and manage agents with a keyboard-driven interface.

<!-- Screenshot placeholder: ![lanchr TUI](docs/tui.png) -->

## License

MIT
