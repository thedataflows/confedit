package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/types"
)

type ApplyCmdTestSuite struct {
	suite.Suite
}

func TestApplyCmdTestSuite(t *testing.T) {
	suite.Run(t, new(ApplyCmdTestSuite))
}

func (s *ApplyCmdTestSuite) TestBackupFlag_FileTarget() {
	tests := []struct {
		name           string
		backupFlag     bool
		originalBackup bool
		expectedBackup bool
	}{
		{
			name:           "backup flag true overrides false",
			backupFlag:     true,
			originalBackup: false,
			expectedBackup: true,
		},
		{
			name:           "backup flag true with true",
			backupFlag:     true,
			originalBackup: true,
			expectedBackup: true,
		},
		{
			name:           "backup flag false preserves original",
			backupFlag:     false,
			originalBackup: true,
			expectedBackup: true,
		},
		{
			name:           "backup flag false with false",
			backupFlag:     false,
			originalBackup: false,
			expectedBackup: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Create a test file target
			fileTarget := file.NewTarget("test-target", "/tmp/test.conf", "ini")
			require.NotNil(s.T(), fileTarget)
			require.NotNil(s.T(), fileTarget.GetConfig())

			// Set the original backup value
			fileTarget.GetConfig().Backup = tt.originalBackup

			targets := []types.AnyTarget{fileTarget}

			// Apply backup override logic
			if tt.backupFlag {
				for i := range targets {
					if targets[i].GetType() == types.TYPE_FILE {
						if fileTargetTyped, ok := targets[i].(*file.Target); ok {
							fileTargetTyped.GetConfig().Backup = true
						}
					}
				}
			}

			// Verify the result
			fileTargetTyped, ok := targets[0].(*file.Target)
			require.True(s.T(), ok, "target should be a FileConfigTarget")
			assert.Equal(s.T(), tt.expectedBackup, fileTargetTyped.GetConfig().Backup)
		})
	}
}

func (s *ApplyCmdTestSuite) TestBackupFlag_MultipleTargets() {
	fileTarget1 := file.NewTarget("target1", "/tmp/test1.conf", "ini")
	fileTarget2 := file.NewTarget("target2", "/tmp/test2.conf", "yaml")

	// Set different initial backup values
	fileTarget1.GetConfig().Backup = false
	fileTarget2.GetConfig().Backup = true

	targets := []types.AnyTarget{fileTarget1, fileTarget2}

	// Apply backup flag = true
	backupFlag := true
	if backupFlag {
		for i := range targets {
			if targets[i].GetType() == types.TYPE_FILE {
				if fileTargetTyped, ok := targets[i].(*file.Target); ok {
					fileTargetTyped.GetConfig().Backup = true
				}
			}
		}
	}

	// Verify both targets have backup enabled
	for i, target := range targets {
		fileTargetTyped, ok := target.(*file.Target)
		require.True(s.T(), ok, "target %d should be a FileConfigTarget", i)
		assert.True(s.T(), fileTargetTyped.GetConfig().Backup, "target %d should have backup enabled", i)
	}
}

func (s *ApplyCmdTestSuite) TestBackupFlag_NonFileTargets() {
	// Create a mock non-file target (using file target but with different type)
	fileTarget := file.NewTarget("test-target", "/tmp/test.conf", "ini")
	fileTarget.Type = types.TYPE_DCONF // Change to non-file type

	targets := []types.AnyTarget{fileTarget}

	// Apply backup flag = true
	backupFlag := true
	if backupFlag {
		for i := range targets {
			if targets[i].GetType() == types.TYPE_FILE {
				if fileTargetTyped, ok := targets[i].(*file.Target); ok {
					fileTargetTyped.GetConfig().Backup = true
				}
			}
		}
	}

	// Verify the backup flag was not applied to non-file targets
	fileTargetTyped, ok := targets[0].(*file.Target)
	require.True(s.T(), ok, "target should be a FileConfigTarget")
	assert.False(s.T(), fileTargetTyped.GetConfig().Backup, "backup should not be applied to non-file targets")
}

func (s *ApplyCmdTestSuite) TestBackupFlag_EdgeCases() {
	s.Run("empty targets slice", func() {
		targets := []types.AnyTarget{}

		// Apply backup flag = true (should not panic)
		backupFlag := true
		if backupFlag {
			for i := range targets {
				if targets[i].GetType() == types.TYPE_FILE {
					if fileTargetTyped, ok := targets[i].(*file.Target); ok {
						fileTargetTyped.GetConfig().Backup = true
					}
				}
			}
		}

		assert.Empty(s.T(), targets)
	})

	s.Run("nil target in slice", func() {
		targets := []types.AnyTarget{nil}

		// Apply backup flag = true (should not panic on nil check)
		backupFlag := true
		if backupFlag {
			for i := range targets {
				if targets[i] != nil && targets[i].GetType() == types.TYPE_FILE {
					if fileTargetTyped, ok := targets[i].(*file.Target); ok {
						fileTargetTyped.GetConfig().Backup = true
					}
				}
			}
		}

		assert.Nil(s.T(), targets[0])
	})
}
