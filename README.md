# confedit

A powerful CLI tool for managing and applying configuration files across different formats and systems. Built with CUE for type-safe configuration definition and Go for high performance.

## Features

### Core Capabilities

- **🔧 Multi-Format Support**: INI, YAML, TOML, JSON, XML configuration files
- **📋 CUE Configuration**: Type-safe configuration definition with schema validation
- **🎯 Target Management**: Apply configurations to files, dconf, systemd, sed operations
- **🔄 Deep Merging**: Automatically merge multiple CUE files with conflict resolution
- **📊 Status Checking**: Compare desired vs actual configuration state
- **💾 Automatic Backups**: Optional backup creation with checksums before modifications
- **🎨 Clean CLI**: Modern interface with subcommands (apply, status, list, generate)

### Configuration Features

- **Variables**: Define and reuse values across targets with `variables: { ... }`
- **Hooks**: Execute custom scripts before/after applying configurations
- **Multi-File**: Split configurations across multiple CUE files for better organization
- **Format Options**: Control spacing, indentation, pretty printing per format
- **Discovery**: List and inspect all configured targets in multiple output formats

## Usage

### CLI Commands

**Show all available commands:**

```bash
./confedit --help
```

**List configured targets:**

```bash
# Basic list (table format)
./confedit list

# Detailed information
./confedit list --long

# JSON output
./confedit list --format json

# YAML output with details
./confedit list --format yaml --long
```

**Check configuration status:**

```bash
# Check all targets
./confedit status

# Check specific targets
./confedit status target1 target2

# Use custom config directory
./confedit status --config /path/to/config
```

**Apply configurations:**

```bash
# Apply all configurations
./confedit apply

# Apply specific targets
./confedit apply target1 target2

# Dry run (preview changes)
./confedit apply --dry-run

# Force backup creation
./confedit apply --backup

# Use custom config directory
./confedit apply --config /path/to/config
```

**Generate configurations:**

```bash
# Generate configuration from differences between files
./confedit generate --type file source.ini target.ini

# Generate sed configuration from file differences
./confedit generate --type sed original.conf modified.conf

# Generate with custom name and output
./confedit generate --type file --name my-config --output my-config.cue source.yaml target.yaml

# Specify file format explicitly
./confedit generate --type file --file-format ini source.ini target.ini
```

**Show version:**

```bash
./confedit version
```

### Configuration Structure

Configurations are defined using CUE files. See complete examples in [`testdata/`](testdata/).

**Basic structure:**

- `variables: { ... }` - Define reusable values across targets
- `targets: [ ... ]` - Configuration targets (file, dconf, systemd, sed)
- `hooks: { ... }` - Optional pre/post apply commands

**Multi-file support:**
Split configurations across multiple `.cue` files for better organization. The tool automatically discovers and merges all CUE files in the config directory, merging targets with the same name.

### Target Types

**`file`** - Configuration file management

- Supported formats: INI, YAML, TOML, JSON, XML
- Features: Backup support, ownership/permissions control, format-specific options
- Use cases: Application configs, system settings, any structured file

**`dconf`** - GNOME/GTK settings

- Purpose: Manage GNOME desktop and GTK application settings
- Features: User-specific or system-wide settings, automatic dconf updates
- Use cases: Desktop environment configuration, GNOME app preferences

**`systemd`** - systemd service configuration

- Purpose: Manage systemd unit properties and settings
- Features: Automatic daemon-reload after changes, property management
- Use cases: Service configuration, unit file modifications

**`sed`** - Text file editing with sed

- Purpose: Apply sed commands for precise text file edits
- Features: Multiple sed operations, backup support, idempotent changes
- Use cases: Complex find/replace, line insertion/deletion, regex-based edits

## Examples

Complete working examples are in [`testdata/`](testdata/). All examples can be tested without modifying your system.

### Basic Usage Examples

**INI file configuration** - [`testdata/01-example-config.cue`](testdata/01-example-config.cue)

```bash
./confedit list -c testdata/01-example-config.cue
./confedit status -c testdata/01-example-config.cue
```

**Variables** - [`testdata/variables-config.cue`](testdata/variables-config.cue)

```bash
./confedit list -c testdata/variables-config.cue
```

**Multi-file merging** - Automatically merges all `.cue` files in a directory

```bash
./confedit list -c testdata/
```

**Hooks** - Pre/post apply commands - [`testdata/hooks-config/`](testdata/hooks-config/)

```bash
./confedit list -c testdata/hooks-config/
./confedit apply -c testdata/hooks-config/ --dry-run
```

**Sed operations** - [`testdata/sed-example-config.cue`](testdata/sed-example-config.cue)

```bash
./confedit list -c testdata/sed-example-config.cue
./confedit apply -c testdata/sed-example-config.cue --dry-run
```

### Common Workflows

**Safe configuration changes:**

```bash
# 1. Check what would change
./confedit status -c testdata/

# 2. Preview changes (dry-run)
./confedit apply --dry-run -c testdata/

# 3. Apply with backups
./confedit apply --backup -c testdata/
```

**Working with specific targets:**

```bash
# Check specific targets only
./confedit status example-ini-config -c testdata/

# Apply only specific targets
./confedit apply example-ini-config --backup -c testdata/
```

**Export configuration information:**

```bash
# List all targets (table format)
./confedit list -c testdata/

# Detailed information
./confedit list --long -c testdata/

# Export as JSON
./confedit list --format json -c testdata/ > targets.json

# Export as YAML
./confedit list --format yaml --long -c testdata/
```

## Development

**Build from source:**

```bash
git clone https://github.com/thedataflows/confedit.git
cd confedit
go build .
```

**Development workflow:**

This repo uses [mise-en-place](https://github.com/jdx/mise)

```bash
# Set environment using mise
mise up

# Run in development mode
LOG_LEVEL=debug go run .

# Build binary
go build .

# Run tests
go test ./...
```

**Requirements:**

- Go 1.25+

## Docker

**Build Docker image:**

```bash
docker build -t confedit .
```

**Run with Docker:**

```bash
# Run interactively
docker run --rm -it confedit --help
```

**Use pre-built image from GitHub Container Registry:**

```bash
docker pull ghcr.io/thedataflows/confedit:latest
docker run --rm ghcr.io/thedataflows/confedit:latest version
```

## License

[MIT License](LICENSE)
