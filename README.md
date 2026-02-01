# Rollbar CLI

A command-line interface for [Rollbar](https://rollbar.com) focused on **reading and listing items and occurrences**. Optimized for both AI coding agents and human users.

## Installation

### Arch Linux (AUR)

```bash
yay -S rollbar-cli
```

### macOS (Homebrew)

```bash
brew install robzolkos/tap/rollbar-cli
```

### Debian/Ubuntu

```bash
# Download the .deb for your architecture (amd64 or arm64)
curl -LO https://github.com/robzolkos/rollbar-cli/releases/latest/download/rollbar-cli_VERSION_amd64.deb
sudo dpkg -i rollbar-cli_VERSION_amd64.deb
```

### Fedora/RHEL

```bash
# Download the .rpm for your architecture (x86_64 or aarch64)
curl -LO https://github.com/robzolkos/rollbar-cli/releases/latest/download/rollbar-cli-VERSION-1.x86_64.rpm
sudo rpm -i rollbar-cli-VERSION-1.x86_64.rpm
```

### Windows

Download `rollbar-windows-amd64.exe` from [GitHub Releases](https://github.com/robzolkos/rollbar-cli/releases), rename it to `rollbar.exe`, and add it to your PATH.

### With Go

```bash
go install github.com/robzolkos/rollbar-cli/cmd/rollbar@latest
```

### From Source

```bash
git clone https://github.com/robzolkos/rollbar-cli.git
cd rollbar-cli
make build
./bin/rollbar --help
```

## Quick Start

1. **Get a read token** from Rollbar:
   - Go to your project's Settings → Access Tokens
   - Create a token with "read" scope (or use an existing one)

2. **Configure the CLI**:
   ```bash
   # Option 1: Environment variable
   export ROLLBAR_ACCESS_TOKEN="your-token"

   # Option 2: Create a config file
   rollbar init
   rollbar config set access_token your-token
   ```

3. **Test your setup**:
   ```bash
   rollbar whoami
   ```

4. **List recent errors**:
   ```bash
   rollbar items --since "8 hours ago" --level error,critical
   ```

## Commands

### List Items (Errors)

```bash
# List active items
rollbar items

# Filter by status, level, environment
rollbar items --status active --level error,critical --env production

# Time-based filtering
rollbar items --since "8 hours ago"
rollbar items --since 24h
rollbar items --from "2026-01-30T09:00:00" --to "2026-01-30T17:00:00"

# Search by title
rollbar items --query "TypeError"

# Sort by occurrence count
rollbar items --sort occurrences --limit 10
```

### Get Item Details

```bash
# Get by project counter
rollbar item 123

# Include recent occurrences
rollbar item 123 --occurrences 5
```

### Generate AI Context

The `context` command generates comprehensive markdown with everything needed to fix a bug:

```bash
# Output to stdout
rollbar context 123

# Write to file
rollbar context 123 --out bug-context.md

# Copy to clipboard (macOS)
rollbar context 123 | pbcopy
```

### View Occurrences

```bash
# List occurrences for an item
rollbar occurrences --item 123

# Get single occurrence details
rollbar occurrence 453568801204
```

## Output Formats

| Format | Flag | Use Case |
|--------|------|----------|
| Table | `--output table` (default) | Human-readable in terminal |
| JSON | `--output json` | Scripting and piping to jq |
| Compact | `--output compact` | Token-efficient for AI agents |
| Markdown | `--output markdown` | Documentation and context files |

Use `--ai` as shorthand for `--output compact --no-color`.

## Configuration

### Config File Discovery

The CLI looks for configuration in this order:

1. `--config` flag (explicit path)
2. `.rollbar.yaml` in current directory
3. Walk up directory tree looking for `.rollbar.yaml`
4. `~/.config/rollbar/config.yaml` (global)
5. Environment variables

### Config File Format

```yaml
# .rollbar.yaml
access_token: "your-read-token"
project_id: 12345
default_environment: "production"

output:
  format: "table"
  color: "auto"
```

### Environment Variables

- `ROLLBAR_ACCESS_TOKEN` - Your read token
- `ROLLBAR_ENVIRONMENT` - Default environment filter

## AI Agent Integration

This CLI is designed for AI coding agents. Key features:

- **Token-efficient output**: Use `--ai` flag for compact output
- **Structured context**: `rollbar context` generates comprehensive bug reports
- **Time-based queries**: Natural language durations like "8 hours ago"
- **Exit codes**: Proper exit codes for scripting (0=success, 1=error, 2=auth error, 3=not found, 4=rate limited)

### Example Workflows

```bash
# "What errors happened overnight?"
rollbar items --since "12 hours ago" --level error,critical --env production --ai

# "Get context for the most frequent error"
rollbar items --sort occurrences --limit 1 --output json | jq -r '.[0].counter' | xargs rollbar context

# "Find all TypeError issues"
rollbar items --query "TypeError" --level error --ai
```

## Claude Code Skill

This CLI includes an agent skill for Claude Code. When installed, Claude can automatically investigate Rollbar errors when you ask about production issues.

### Install the Skill

```bash
npx skills add robzolkos/rollbar-cli
```

Or manually copy to your skills directory:

```bash
# Personal (all projects)
cp -r skills/rollbar ~/.claude/skills/

# Project-specific
mkdir -p .claude/skills
cp -r skills/rollbar .claude/skills/
```

### What the Skill Enables

Once installed, Claude will automatically use the Rollbar CLI when you ask:

- "What errors happened overnight?"
- "Investigate error #123"
- "What's breaking in production?"
- "Get me context to fix this Rollbar issue"

Claude will query Rollbar, analyze the errors, and provide context for bug fixes—all without you needing to run commands manually.

## Shell Completions

Generate completions for your shell:

```bash
# Bash
rollbar completion bash > /etc/bash_completion.d/rollbar

# Zsh
rollbar completion zsh > "${fpath[1]}/_rollbar"

# Fish
rollbar completion fish > ~/.config/fish/completions/rollbar.fish
```

## Development

```bash
# Run tests
make test

# Run with coverage
make test-cover

# Lint
make lint

# Build
make build

# Run E2E tests (requires ROLLBAR_E2E_TOKEN)
make test-e2e
```

## License

MIT
