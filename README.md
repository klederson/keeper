# Keeper

Backup daemon & CLI tool for developers. Configure, schedule, and monitor backups with an interactive terminal interface.

## Features

- **Interactive CLI** — Configure backups with guided forms, no manual YAML editing
- **rsync + SSH** — Battle-tested backup engine with incremental transfers
- **Scheduler** — Cron-based scheduling with systemd integration
- **TUI Dashboard** — Real-time monitoring with a beautiful terminal UI
- **Reports** — Track backup history, success rates, and transfer stats

## Quick Start

```bash
# Build
make build

# Initialize configuration
keeper init

# Add a backup job
keeper add

# Test the connection (dry-run)
keeper test my-project

# Run a backup
keeper run my-project

# Start the daemon (scheduler)
keeper daemon start
```

## Commands

| Command | Description |
|---------|-------------|
| `keeper init` | Interactive setup wizard |
| `keeper add` | Add a new backup job |
| `keeper list` | List all backup jobs |
| `keeper edit <job>` | Edit a backup job |
| `keeper remove <job>` | Remove a backup job |
| `keeper run <job>` | Run a backup now |
| `keeper run --all` | Run all backup jobs |
| `keeper test <job>` | Dry-run (verify without transferring) |
| `keeper status` | Status of all jobs |
| `keeper logs [job]` | View backup logs |
| `keeper dashboard` | Interactive TUI dashboard |
| `keeper daemon start` | Start the scheduler daemon |
| `keeper daemon stop` | Stop the daemon |
| `keeper daemon status` | Check daemon status |
| `keeper doctor` | Check dependencies & connectivity |

## Configuration

Config lives at `~/.config/keeper/config.yaml`. See [configs/keeper.example.yaml](configs/keeper.example.yaml) for a full example.

## Daemon (systemd)

```bash
# Install the service (user-level)
mkdir -p ~/.config/systemd/user
cp deployments/keeper.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now keeper
```

## Requirements

- Go 1.22+
- rsync
- ssh

## License

MIT
