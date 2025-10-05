package file_test

import (
	"testing"

	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/types"
)

func TestFileFeature_Type(t *testing.T) {
	feature := file.New()

	if feature.Type() != types.TYPE_FILE {
		t.Errorf("expected type %s, got %s", types.TYPE_FILE, feature.Type())
	}
}

func TestFileFeature_Executor(t *testing.T) {
	feature := file.New()

	executor := feature.Executor()
	if executor == nil {
		t.Fatal("executor should not be nil")
	}
}

func TestFileFeature_Validate(t *testing.T) {
	feature := file.New()

	tests := []struct {
		name    string
		config  *file.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &file.Config{
				Path:   "/tmp/test.ini",
				Format: "ini",
			},
			wantErr: false,
		},
		{
			name: "missing path",
			config: &file.Config{
				Format: "ini",
			},
			wantErr: true,
		},
		{
			name: "missing format",
			config: &file.Config{
				Path: "/tmp/test.ini",
			},
			wantErr: true,
		},
		{
			name: "unsupported format",
			config: &file.Config{
				Path:   "/tmp/test.txt",
				Format: "txt",
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

func TestFileFeature_NewTarget(t *testing.T) {
	feature := file.New()

	config := &file.Config{
		Path:   "/tmp/test.ini",
		Format: "ini",
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

	if target.GetType() != types.TYPE_FILE {
		t.Errorf("expected type %s, got %s", types.TYPE_FILE, target.GetType())
	}
}
