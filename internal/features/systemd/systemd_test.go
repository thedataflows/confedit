package systemd_test

import (
	"testing"

	"github.com/thedataflows/confedit/internal/features/systemd"
	"github.com/thedataflows/confedit/internal/types"
)

func TestSystemdFeature_Type(t *testing.T) {
	feature := systemd.New()

	if feature.Type() != types.TYPE_SYSTEMD {
		t.Errorf("expected type %s, got %s", types.TYPE_SYSTEMD, feature.Type())
	}
}

func TestSystemdFeature_Executor(t *testing.T) {
	feature := systemd.New()

	executor := feature.Executor()
	if executor == nil {
		t.Fatal("executor should not be nil")
	}
}

func TestSystemdFeature_ValidateConfig(t *testing.T) {
	feature := systemd.New()

	tests := []struct {
		name    string
		config  *systemd.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &systemd.Config{
				Unit:    "nginx.service",
				Section: "Service",
			},
			wantErr: false,
		},
		{
			name: "missing unit",
			config: &systemd.Config{
				Section: "Service",
			},
			wantErr: true,
		},
		{
			name: "missing section",
			config: &systemd.Config{
				Unit: "nginx.service",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(tb *testing.T) {
			err := feature.ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				tb.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSystemdFeature_NewTarget(t *testing.T) {
	feature := systemd.New()

	config := &systemd.Config{
		Unit:    "nginx.service",
		Section: "Service",
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

	if target.GetType() != types.TYPE_SYSTEMD {
		t.Errorf("expected type %s, got %s", types.TYPE_SYSTEMD, target.GetType())
	}
}
