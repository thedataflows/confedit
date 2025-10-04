package types

// BaseTarget contains common fields for all target types
type BaseTarget[T TargetConfig] struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Config   T                      `json:"config"`
}

// GetName implements AnyTarget interface
func (bt *BaseTarget[T]) GetName() string {
	return bt.Name
}

// GetType implements AnyTarget interface
func (bt *BaseTarget[T]) GetType() string {
	return bt.Type
}

// GetMetadata implements AnyTarget interface
func (bt *BaseTarget[T]) GetMetadata() map[string]interface{} {
	return bt.Metadata
}

// Validate implements AnyTarget interface
func (bt *BaseTarget[T]) Validate() error {
	return bt.Config.Validate()
}

// GetConfig returns the typed configuration
func (bt *BaseTarget[T]) GetConfig() T {
	return bt.Config
}
