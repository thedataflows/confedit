package dconf_test

import (
	"testing"

	"github.com/thedataflows/confedit/internal/features/dconf"
	"github.com/thedataflows/confedit/internal/types"
)

func TestDconfFeature_Type(t *testing.T) {
	feature := dconf.New()

	if feature.Type() != types.TYPE_DCONF {
		t.Errorf("expected type %s, got %s", types.TYPE_DCONF, feature.Type())
	}
}

func TestDconfFeature_Executor(t *testing.T) {
	feature := dconf.New()

	executor := feature.Executor()
	if executor == nil {
		t.Fatal("executor should not be nil")
	}
}

func TestDconfFeature_Validate(t *testing.T) {
	feature := dconf.New()

	tests := []struct {
		name    string
		config  *dconf.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &dconf.Config{
				Schema: "/org/gnome/desktop/interface",
			},
			wantErr: false,
		},
		{
			name:    "missing schema",
			config:  &dconf.Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(tb *testing.T) {
			err := feature.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				tb.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDconfFeature_NewTarget(t *testing.T) {
	feature := dconf.New()

	config := &dconf.Config{
		Schema: "/org/gnome/desktop/interface",
	}

	target, err := feature.NewTarget("test-target", config)
	if err != nil {
		t.Fatalf("NewTarget() error = %v", err)
	}

	if target == nil {
		t.Fatal("target should not be nil")
	}

	if target.GetName() != "test-target" {
		t.Errorf("expected name 'test-target', got %s", target.GetName())
	}

	if target.GetType() != types.TYPE_DCONF {
		t.Errorf("expected type %s, got %s", types.TYPE_DCONF, target.GetType())
	}
}
