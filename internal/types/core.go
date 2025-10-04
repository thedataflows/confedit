package types

// Core domain interfaces and system configuration

const (
	TYPE_FILE    = "file"
	TYPE_DCONF   = "dconf"
	TYPE_SYSTEMD = "systemd"
	TYPE_SED     = "sed"
)

// AnyTarget is a union type for all possible target types
type AnyTarget interface {
	GetName() string
	GetType() string
	GetMetadata() map[string]interface{}
	Validate() error
}

// TargetConfig is the interface that all target configuration types must implement
type TargetConfig interface {
	Type() string
	Validate() error
}

// SystemConfig represents the top-level configuration structure
type SystemConfig struct {
	Targets   []AnyTarget            `json:"targets"`
	Variables map[string]interface{} `json:"variables,omitempty"`
	Hooks     *Hooks                 `json:"hooks,omitempty"`
}
