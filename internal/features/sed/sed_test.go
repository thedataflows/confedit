package sed_test

import (
	"testing"

	"github.com/thedataflows/confedit/internal/features/sed"
	"github.com/thedataflows/confedit/internal/types"
)

func TestSedFeature_Type(t *testing.T) {
	feature := sed.New()

	if feature.Type() != types.TYPE_SED {
		t.Errorf("expected type %s, got %s", types.TYPE_SED, feature.Type())
	}
}

func TestSedFeature_Executor(t *testing.T) {
	feature := sed.New()

	executor := feature.Executor()
	if executor == nil {
		t.Fatal("executor should not be nil")
	}
}

func TestSedFeature_Validate(t *testing.T) {
	feature := sed.New()

	tests := []struct {
		name    string
		config  *sed.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &sed.Config{
				Path:     "/tmp/test.txt",
				Commands: []string{"s/foo/bar/g"},
			},
			wantErr: false,
		},
		{
			name: "missing path",
			config: &sed.Config{
				Commands: []string{"s/foo/bar/g"},
			},
			wantErr: true,
		},
		{
			name: "missing commands",
			config: &sed.Config{
				Path: "/tmp/test.txt",
			},
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

func TestSedFeature_NewTarget(t *testing.T) {
	feature := sed.New()

	config := &sed.Config{
		Path:     "/tmp/test.txt",
		Commands: []string{"s/foo/bar/g"},
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

	if target.GetType() != types.TYPE_SED {
		t.Errorf("expected type %s, got %s", types.TYPE_SED, target.GetType())
	}
}
