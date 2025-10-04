# confedit

A powerful CLI tool for managing and applying configuration files across different formats and systems. Built with CUE for type-safe configuration definition and Go for high performance.

## Features

- **🔧 Multi-Format Support**: INI, YAML, TOML, JSON configuration files
- **📋 CUE Configuration**: Type-safe configuration definition with schema validation
- **🎯 Target Management**: Apply configurations to files, dconf, systemd, sed operations
- **🔄 Deep Merging**: Merge multiple CUE files with conflict resolution and source tracking
- **📊 Status Checking**: Compare desired vs actual configuration state
- **💾 Backup Support**: Automatic backup creation with checksums
- **🎨 Modern CLI**: Clean interface with subcommands using Kong framework
- **🔍 Configuration Discovery**: List and inspect configured targets
- **⚙️ Format-Specific Options**: INI spacing control, indentation, pretty printing, etc.
- **🔍 Sed Operations**: Apply sed commands to any text file for precise edits

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
```

**Show version:**

```bash
./confedit version
```

### Configuration Structure

Configurations are defined using CUE files with the following structure. See the complete examples in the [`testdata/`](testdata/) directory.

**Basic structure:**

- **Variables**: Define reusable values with `variables: { ... }`
- **Targets**: Configuration targets with `targets: [ ... ]`
- **Hooks**: Optional pre/post commands with `hooks: { ... }`

### Multi-File Configuration

You can split configurations across multiple CUE files for better organization. The tool automatically merges configurations with the same target name.

### Target Types

**File targets** (`type: "file"`):

- **Formats**: `ini`, `yaml`, `toml`, `json`, `xml`
- **Features**: Backup support, ownership/permissions, format-specific options

**dconf targets** (`type: "dconf"`):

- **Purpose**: GNOME/GTK application settings
- **Features**: User-specific or system-wide settings

**systemd targets** (`type: "systemd"`):

- **Purpose**: systemd service configuration
- **Features**: Automatic daemon reload, property management

**sed targets** (`type: "sed"`):

- **Purpose**: Apply sed commands to any text file for precise edits
- **Features**: Multiple sed commands, backup support, idempotent operations

## Examples

For complete working examples, see the [`testdata/`](testdata/) directory which contains real configuration files you can test with:

### Basic INI Configuration

See [`testdata/01-example-config.cue`](testdata/01-example-config.cue) for a complete example showing:

- INI file target configuration
- Custom options (spacing control)
- Content with sections and root-level keys
- Special INI values (commented keys, deleted keys)

**Test it:**

```bash
./confedit list -c testdata/01-example-config.cue
./confedit status -c testdata/01-example-config.cue
```

### Multi-Target Configuration with Variables

See [`testdata/variables-config.cue`](testdata/variables-config.cue) for an example showing:

- Variable definitions and usage
- Variable interpolation in target configurations
- Shared configuration patterns

**Test it:**

```bash
./confedit list -c testdata/variables-config.cue
```

### Multi-File Merging

The testdata directory demonstrates automatic file merging:

- [`testdata/01-example-config.cue`](testdata/01-example-config.cue)
- [`testdata/02-example-config.cue`](testdata/02-example-config.cue)
- [`testdata/variables-config.cue`](testdata/variables-config.cue)

**Test merging:**

```bash
# Automatically discovers and merges all .cue files in the directory
./confedit list -c testdata/
```

### Hooks Configuration

See [`testdata/hooks-config/hooks.cue`](testdata/hooks-config/hooks.cue) for an example showing:

- Pre-apply and post-apply hooks
- Shell script execution
- Environment validation
- Error handling

**Test it:**

```bash
./confedit list -c testdata/hooks-config/
./confedit apply -c testdata/hooks-config/ --dry-run
```

### Sed Operations

See [`testdata/sed-example-config.cue`](testdata/sed-example-config.cue) for an example showing:

- Sed command configuration
- Multiple sed operations (find/replace, delete, insert)
- Backup support
- Variable usage in sed targets

**Test it:**

```bash
./confedit list -c testdata/sed-example-config.cue
./confedit status -c testdata/sed-example-config.cue
./confedit apply -c testdata/sed-example-config.cue --dry-run
```

### Real-World Usage Patterns

**Check what would change:**

```bash
# See what changes would be made using testdata examples
./confedit status -c testdata/

# Check specific targets only
./confedit status example-ini-config -c testdata/

# Use different config directory
./confedit status -c ./testdata/hooks-config
```

**Apply configurations safely:**

```bash
# Preview changes first using testdata
./confedit apply --dry-run -c testdata/

# Apply with backups
./confedit apply --backup -c testdata/

# Apply only specific targets
./confedit apply example-ini-config --backup -c testdata/

# Use hooks configuration
./confedit apply -c testdata/hooks-config/ --dry-run
```

**Inspect configuration:**

```bash
# List all targets from testdata
./confedit list -c testdata/

# Show detailed information
./confedit list --long -c testdata/

# Export as JSON for processing
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

**Dependencies:**

- Go 1.24+

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
