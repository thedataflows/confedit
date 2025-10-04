package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StateManagerTestSuite struct {
	suite.Suite
	manager *Manager
}

func TestStateManagerTestSuite(t *testing.T) {
	suite.Run(t, new(StateManagerTestSuite))
}

func (s *StateManagerTestSuite) SetupTest() {
	s.manager = NewManager("")
}

func (s *StateManagerTestSuite) TestNewManager() {
	manager := NewManager("/some/path")
	assert.NotNil(s.T(), manager)
}

func (s *StateManagerTestSuite) TestFilterManagedKeys_BasicFiltering() {
	current := map[string]interface{}{
		"managed_key":   "value1",
		"unmanaged_key": "value2",
	}
	desired := map[string]interface{}{
		"managed_key": "new_value",
	}

	filtered := s.manager.filterManagedKeys(current, desired)

	assert.Len(s.T(), filtered, 1)
	assert.Equal(s.T(), "value1", filtered["managed_key"])
	assert.NotContains(s.T(), filtered, "unmanaged_key")
}

func (s *StateManagerTestSuite) TestFilterManagedKeys_INIRootSection() {
	// Simulate INI parser output with root section
	current := map[string]interface{}{
		"": map[string]interface{}{
			"direct_key": "value1",
			"other_key":  "value2",
		},
		"section1": map[string]interface{}{
			"key1": "value3",
		},
	}
	desired := map[string]interface{}{
		"direct_key": "new_value",
		"section1": map[string]interface{}{
			"key1": "new_value3",
		},
	}

	filtered := s.manager.filterManagedKeys(current, desired)

	assert.Len(s.T(), filtered, 2)
	assert.Equal(s.T(), "value1", filtered["direct_key"])

	section1, ok := filtered["section1"].(map[string]interface{})
	require.True(s.T(), ok, "section1 should be a map")
	assert.Equal(s.T(), "value3", section1["key1"])
}

func (s *StateManagerTestSuite) TestFilterManagedKeys_NestedStructures() {
	current := map[string]interface{}{
		"section": map[string]interface{}{
			"managed":   "value1",
			"unmanaged": "value2",
		},
	}
	desired := map[string]interface{}{
		"section": map[string]interface{}{
			"managed": "new_value",
		},
	}

	filtered := s.manager.filterManagedKeys(current, desired)

	section, ok := filtered["section"].(map[string]interface{})
	require.True(s.T(), ok, "section should be a map")
	assert.Len(s.T(), section, 1)
	assert.Equal(s.T(), "value1", section["managed"])
	assert.NotContains(s.T(), section, "unmanaged")
}

func (s *StateManagerTestSuite) TestFilterManagedKeys_EdgeCases() {
	s.Run("empty current state", func() {
		current := map[string]interface{}{}
		desired := map[string]interface{}{
			"key": "value",
		}

		filtered := s.manager.filterManagedKeys(current, desired)
		assert.Empty(s.T(), filtered)
	})

	s.Run("empty desired state", func() {
		current := map[string]interface{}{
			"key": "value",
		}
		desired := map[string]interface{}{}

		filtered := s.manager.filterManagedKeys(current, desired)
		assert.Empty(s.T(), filtered)
	})

	s.Run("nil states", func() {
		// Should not panic
		filtered := s.manager.filterManagedKeys(nil, nil)
		assert.Empty(s.T(), filtered)
	})
}

func (s *StateManagerTestSuite) TestFilterManagedKeys_TypeMismatch() {
	current := map[string]interface{}{
		"key": "string_value",
	}
	desired := map[string]interface{}{
		"key": map[string]interface{}{
			"nested": "value",
		},
	}

	filtered := s.manager.filterManagedKeys(current, desired)
	assert.Equal(s.T(), "string_value", filtered["key"])
}

func (s *StateManagerTestSuite) TestComputeDiffWithCurrent_BasicDiff() {
	current := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	desired := map[string]interface{}{
		"key1": "new_value1",
		"key3": "value3",
	}

	diff, err := s.manager.ComputeDiffWithCurrent("test-target", desired, current)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test-target", diff.Target)
	assert.False(s.T(), diff.IsEmpty())
}

func (s *StateManagerTestSuite) TestComputeDiffWithCurrent_NilCurrentState() {
	desired := map[string]interface{}{
		"key": "value",
	}

	diff, err := s.manager.ComputeDiffWithCurrent("test-target", desired, nil)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test-target", diff.Target)
	assert.False(s.T(), diff.IsEmpty())
}

func (s *StateManagerTestSuite) TestComputeDiffWithCurrent_EmptyDesiredState() {
	current := map[string]interface{}{
		"key": "value",
	}
	desired := map[string]interface{}{}

	diff, err := s.manager.ComputeDiffWithCurrent("test-target", desired, current)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test-target", diff.Target)
	assert.True(s.T(), diff.IsEmpty())
}

func (s *StateManagerTestSuite) TestComputeDiffWithCurrent_NoChanges() {
	current := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	desired := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	diff, err := s.manager.ComputeDiffWithCurrent("test-target", desired, current)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test-target", diff.Target)
	assert.True(s.T(), diff.IsEmpty())
}

func (s *StateManagerTestSuite) TestExtractRootSection() {
	s.Run("with root section", func() {
		current := map[string]interface{}{
			"": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		}

		rootSection := s.manager.extractRootSection(current)
		require.NotNil(s.T(), rootSection)
		assert.Equal(s.T(), "value1", rootSection["key1"])
		assert.Equal(s.T(), "value2", rootSection["key2"])
	})

	s.Run("without root section", func() {
		current := map[string]interface{}{
			"section": map[string]interface{}{
				"key": "value",
			},
		}

		rootSection := s.manager.extractRootSection(current)
		assert.Nil(s.T(), rootSection)
	})

	s.Run("invalid root section type", func() {
		current := map[string]interface{}{
			"": "not a map",
		}

		rootSection := s.manager.extractRootSection(current)
		assert.Nil(s.T(), rootSection)
	})
}

func (s *StateManagerTestSuite) TestFindKeyInCurrent() {
	current := map[string]interface{}{
		"direct_key": "direct_value",
		"": map[string]interface{}{
			"root_key": "root_value",
		},
	}
	rootSection := s.manager.extractRootSection(current)

	s.Run("find direct key", func() {
		value, exists := s.manager.findKeyInCurrent("direct_key", current, rootSection)
		assert.True(s.T(), exists)
		assert.Equal(s.T(), "direct_value", value)
	})

	s.Run("find root section key", func() {
		value, exists := s.manager.findKeyInCurrent("root_key", current, rootSection)
		assert.True(s.T(), exists)
		assert.Equal(s.T(), "root_value", value)
	})

	s.Run("key not found", func() {
		value, exists := s.manager.findKeyInCurrent("nonexistent_key", current, rootSection)
		assert.False(s.T(), exists)
		assert.Nil(s.T(), value)
	})

	s.Run("nil root section", func() {
		value, exists := s.manager.findKeyInCurrent("root_key", current, nil)
		assert.False(s.T(), exists)
		assert.Nil(s.T(), value)
	})
}

func (s *StateManagerTestSuite) TestFilterKeyValue() {
	s.Run("nested maps", func() {
		current := map[string]interface{}{
			"managed":   "value1",
			"unmanaged": "value2",
		}
		desired := map[string]interface{}{
			"managed": "new_value",
		}

		filtered := s.manager.filterKeyValue(current, desired)
		filteredMap, ok := filtered.(map[string]interface{})
		require.True(s.T(), ok)
		assert.Equal(s.T(), "value1", filteredMap["managed"])
		assert.NotContains(s.T(), filteredMap, "unmanaged")
	})

	s.Run("type mismatch", func() {
		current := "string_value"
		desired := map[string]interface{}{
			"nested": "value",
		}

		filtered := s.manager.filterKeyValue(current, desired)
		assert.Equal(s.T(), "string_value", filtered)
	})

	s.Run("leaf values", func() {
		current := "current_value"
		desired := "desired_value"

		filtered := s.manager.filterKeyValue(current, desired)
		assert.Equal(s.T(), "current_value", filtered)
	})
}
