# Internal Package Structure

This directory contains the internal implementation of confedit using a **feature-based architecture**.

## Architecture Overview

The architecture is built around three key concepts:

1. **Features**: Self-contained modules for each target type (file, dconf, systemd, sed)
2. **Interfaces**: Well-defined contracts between components
3. **Registry**: Central discovery and management of features

## Core Interfaces

### Executor (`engine/executor.go`)

The `Executor` interface defines how to apply, validate, and read configuration:

```go
type Executor interface {
    Apply(target types.AnyTarget, diff *state.ConfigDiff) error
    Validate(target types.AnyTarget) error
    GetCurrentState(target types.AnyTarget) (map[string]interface{}, error)
}
```

### Feature (`features/feature.go`)

The `Feature` interface defines a complete configuration target feature:

```go
type Feature interface {
    Type() string
    Executor() engine.Executor
    NewTarget(name string, config interface{}) (types.AnyTarget, error)
    ValidateConfig(config interface{}) error
}
```

## Feature Modules

Each feature is a self-contained package that owns:

- **Interface implementation** - Implements `features.Feature`
- **Executor** - Implements `engine.Executor`
- **Types** - Feature-specific type definitions
- **Utilities** - Feature-specific helper functions
- **Tests** - Comprehensive unit tests

### File Feature (`features/file/`)

The file feature handles configuration files in multiple formats.

**Structure:**

```tree
file/
├── file.go          # Feature interface implementation
├── executor.go      # Executor implementation
├── types.go         # File-specific types (Config, Target, NewTarget)
├── backup.go        # Backup utilities
├── formats/         # Format-specific parsers
│   ├── format.go    # Parser interface
│   ├── registry.go  # Parser registry
│   ├── ini/
│   │   └── iniparser/  # Advanced INI parser with comments
│   ├── yaml/        # YAML format (direct implementation)
│   ├── toml/        # TOML format (direct implementation)
│   ├── json/        # JSON format (direct implementation)
│   └── xml/         # XML format (direct implementation)
└── *_test.go        # Tests
```

**Supported Formats:**

- INI (with configurable options)
- YAML
- TOML
- JSON
- XML

**Example Usage:**

```go
// Create the file feature
feature := file.New()

// Get the executor
executor := feature.Executor()

// Apply configuration
err := executor.Apply(target, diff)
```

## Feature Registry

The registry provides centralized feature management:

```go
// Create registry
registry := features.NewRegistry()

// Register features
registry.Register(file.New())
registry.Register(dconf.New())
registry.Register(systemd.New())
registry.Register(sed.New())

// Get feature by type
feature, err := registry.Get("file")

// Get executor by type (convenience method)
executor, err := registry.Executor("file")
```

## Type Organization

Types are organized using **dependency inversion**:

- **`internal/types/`** - Shared interfaces and contracts
  - `AnyTarget` - Base target interface
  - `TargetConfig` - Configuration interface
  - `BaseTarget[T]` - Generic target implementation
  - `SystemConfig` - Top-level configuration
  - Type constants (`TYPE_FILE`, `TYPE_DCONF`, etc.)

- **Feature packages** - Concrete implementations
  - `file.Config`, `file.Target` - File-specific types
  - `dconf.Config`, `dconf.Target` - DConf-specific types
  - `systemd.Config`, `systemd.Target` - Systemd-specific types
  - `sed.Config`, `sed.Target` - Sed-specific types

This enables:

- Features own their types (co-location)
- No duplication between central types and feature types
- Clean dependency graph (types → features → config/reconciler)

## Design Principles

### 1. **Self-Contained Modules**

Each feature is independent and owns all its code (types, logic, tests). No cross-feature dependencies.

### 2. **Interface-Driven**

Components interact through well-defined interfaces in `internal/types`, enabling loose coupling.

### 3. **Testability**

Each feature can be tested in isolation without external dependencies.

### 4. **Extensibility**

New features are added by implementing interfaces, not modifying existing code.

### 5. **Simplicity**

Minimize complexity. Use the simplest solution that works correctly.

### 6. **Feature Ownership**

Each feature package owns its configuration types, eliminating duplication and centralizing related code.

## Adding a New Feature

To add a new feature (e.g., "git"):

1. **Create the package**: `internal/features/git/`

2. **Define types**: `git/types.go`

   ```go
   package git

   import "github.com/thedataflows/confedit/internal/types"

   // Config represents the configuration for a git target
   type Config struct {
       Repository string `json:"repository"`
       Branch     string `json:"branch"`
       // ...
   }

   // Type implements TargetConfig interface
   func (c *Config) Type() string {
       return types.TYPE_GIT  // Add constant to internal/types/core.go
   }

   // Validate implements TargetConfig interface
   func (c *Config) Validate() error {
       if c.Repository == "" {
           return fmt.Errorf("repository is required")
       }
       return nil
   }

   // Target is a type alias for git-based configuration targets
   type Target = types.BaseTarget[*Config]

   // NewTarget creates a new git configuration target
   func NewTarget(name, repository, branch string) *Target {
       return &Target{
           Name:     name,
           Type:     types.TYPE_GIT,
           Metadata: make(map[string]interface{}),
           Config: &Config{
               Repository: repository,
               Branch:     branch,
           },
       }
   }
   ```

3. **Implement executor**: `git/executor.go`

   ```go
   type Executor struct { /* ... */ }

   func (e *Executor) Apply(target types.AnyTarget, diff *state.ConfigDiff) error {
       gitTarget, ok := target.(*Target)
       if !ok {
           return fmt.Errorf("not a git target")
       }
       // Implementation...
   }
   func (e *Executor) Validate(target types.AnyTarget) error { /* ... */ }
   func (e *Executor) GetCurrentState(target types.AnyTarget) (map[string]interface{}, error) { /* ... */ }
   ```

4. **Implement feature**: `git/git.go`

   ```go
   type Feature struct {
       executor engine.Executor
   }

   func New() features.Feature {
       return &Feature{
           executor: NewExecutor(),
       }
   }

   func (f *Feature) Type() string { return types.TYPE_GIT }
   func (f *Feature) Executor() engine.Executor { return f.executor }
   func (f *Feature) NewTarget(name string, config interface{}) (types.AnyTarget, error) {
       gitConfig, ok := config.(*Config)
       if !ok {
           return nil, fmt.Errorf("invalid config type")
       }
       return &Target{
           Name:     name,
           Type:     types.TYPE_GIT,
           Metadata: make(map[string]interface{}),
           Config:   gitConfig,
       }, nil
   }
   // ...
   ```

5. **Add to type constants**: `internal/types/core.go`

   ```go
   const TYPE_GIT = "git"
   ```

6. **Register in config loader**: `internal/config/loader.go` (createTarget function)

   ```go
   case types.TYPE_GIT:
       config := &git.Config{}
       // decode and return git.Target
   ```

7. **Register it**: In CLI initialization (`cmd/root.go`)

   ```go
   registry.Register(git.New())
   ```

8. **Add tests**: `git/*_test.go`

That's it! The feature is now available throughout the application.

## Adding a New File Format

To add a new file format (e.g., "properties"):

1. **Create package**: `internal/features/file/formats/properties/`

2. **Implement parser**: `properties/parser.go`

   ```go
   type Parser struct { /* ... */ }

   func New() formats.Parser {
       return &Parser{}
   }

   func (p *Parser) Unmarshal(data []byte) (map[string]interface{}, error) { /* ... */ }
   func (p *Parser) Marshal(data map[string]interface{}, w io.Writer) error { /* ... */ }
   ```

3. **Register it**: In `file/file.go`

   ```go
   registry.Register("properties", properties.New())
   ```

## Testing

Run all internal package tests:

```bash
go test ./internal/... -v
```

Run specific feature tests:

```bash
go test ./internal/features/file/... -v
```

Run with coverage:

```bash
go test ./internal/... -cover
```

## Implementation Status

- ✅ **File Feature**: Fully implemented with format parsers (INI, YAML, TOML, JSON, XML)
- ✅ **DConf Feature**: Fully implemented with dconf-cli integration
- ✅ **Systemd Feature**: Fully implemented with systemctl integration
- ✅ **Sed Feature**: Fully implemented with sed command execution
- ✅ **Configuration System**: CUE-based config loading with schema validation
- ✅ **State Management**: State persistence and diffing
- ✅ **Reconciliation Engine**: High-level orchestration with hook support
- ✅ **Type System**: Clean separation between interfaces and implementations
- ✅ **CLI**: Full command set (apply, list, status, generate, init, version)

All core features are complete and tested. The architecture supports easy addition of new features.

## Best Practices

1. **Keep features independent** - No cross-feature imports
2. **Use interfaces** - Depend on interfaces, not implementations
3. **Write tests first** - Test-driven development for new code
4. **Document assumptions** - Use unit tests to validate assumptions
5. **Keep it simple** - Prefer simplicity over clever solutions

## Questions?

See the main [README.md](../README.md)
