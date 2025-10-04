package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ComputeDiffTestSuite struct {
	suite.Suite
}

func (s *ComputeDiffTestSuite) TestComputeDiff_EmptyDesired() {
	current := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": map[string]interface{}{
			"nested": "value",
		},
	}
	desired := map[string]interface{}{}

	diff := ComputeDiff(current, desired)

	// When desired is empty, everything should be marked as removed
	assert.Equal(s.T(), 3, len(diff.Removed))
	assert.Contains(s.T(), diff.Removed, "key1")
	assert.Contains(s.T(), diff.Removed, "key2")
	assert.Contains(s.T(), diff.Removed, "key3")

	// Changes should include all removed items for generate command
	assert.Equal(s.T(), 3, len(diff.Changes))
	assert.Equal(s.T(), "value1", diff.Changes["key1"])
	assert.Equal(s.T(), "value2", diff.Changes["key2"])
	assert.Equal(s.T(), current["key3"], diff.Changes["key3"])

	// Added and Modified should be empty
	assert.Empty(s.T(), diff.Added)
	assert.Empty(s.T(), diff.Modified)

	// Should not be empty
	assert.False(s.T(), diff.IsEmpty())
}

func (s *ComputeDiffTestSuite) TestComputeDiff_EmptyCurrent() {
	current := map[string]interface{}{}
	desired := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	diff := ComputeDiff(current, desired)

	// Everything should be marked as added
	assert.Equal(s.T(), 2, len(diff.Added))
	assert.Equal(s.T(), "value1", diff.Added["key1"])
	assert.Equal(s.T(), "value2", diff.Added["key2"])

	// Changes should include all added items
	assert.Equal(s.T(), 2, len(diff.Changes))
	assert.Equal(s.T(), "value1", diff.Changes["key1"])
	assert.Equal(s.T(), "value2", diff.Changes["key2"])

	// Removed and Modified should be empty
	assert.Empty(s.T(), diff.Removed)
	assert.Empty(s.T(), diff.Modified)

	// Should not be empty
	assert.False(s.T(), diff.IsEmpty())
}

func (s *ComputeDiffTestSuite) TestComputeDiff_BothEmpty() {
	current := map[string]interface{}{}
	desired := map[string]interface{}{}

	diff := ComputeDiff(current, desired)

	// Everything should be empty
	assert.Empty(s.T(), diff.Changes)
	assert.Empty(s.T(), diff.Added)
	assert.Empty(s.T(), diff.Removed)
	assert.Empty(s.T(), diff.Modified)

	// Should be empty
	assert.True(s.T(), diff.IsEmpty())
}

func (s *ComputeDiffTestSuite) TestComputeDiff_MixedChanges() {
	current := map[string]interface{}{
		"unchanged": "value",
		"modified":  "old_value",
		"removed":   "to_be_removed",
	}
	desired := map[string]interface{}{
		"unchanged": "value",
		"modified":  "new_value",
		"added":     "new_item",
	}

	diff := ComputeDiff(current, desired)

	// Check added items
	assert.Equal(s.T(), 1, len(diff.Added))
	assert.Equal(s.T(), "new_item", diff.Added["added"])

	// Check modified items
	assert.Equal(s.T(), 1, len(diff.Modified))
	assert.Equal(s.T(), "old_value", diff.Modified["modified"].Old)
	assert.Equal(s.T(), "new_value", diff.Modified["modified"].New)

	// Check removed items
	assert.Equal(s.T(), 1, len(diff.Removed))
	assert.Contains(s.T(), diff.Removed, "removed")

	// Check changes (should include added, modified, and removed)
	assert.Equal(s.T(), 3, len(diff.Changes))
	assert.Equal(s.T(), "new_item", diff.Changes["added"])
	assert.Equal(s.T(), "new_value", diff.Changes["modified"])
	assert.Equal(s.T(), "to_be_removed", diff.Changes["removed"])

	// Should not be empty
	assert.False(s.T(), diff.IsEmpty())
}

func (s *ComputeDiffTestSuite) TestComputeDiff_NoChanges() {
	current := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	desired := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	diff := ComputeDiff(current, desired)

	// Everything should be empty
	assert.Empty(s.T(), diff.Changes)
	assert.Empty(s.T(), diff.Added)
	assert.Empty(s.T(), diff.Removed)
	assert.Empty(s.T(), diff.Modified)

	// Should be empty
	assert.True(s.T(), diff.IsEmpty())
}

func (s *ComputeDiffTestSuite) TestComputeDiff_NestedMaps() {
	current := map[string]interface{}{
		"nested": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}
	desired := map[string]interface{}{
		"nested": map[string]interface{}{
			"key1": "modified_value",
			"key3": "new_value",
		},
	}

	diff := ComputeDiff(current, desired)

	// Should detect the nested map as modified
	assert.Equal(s.T(), 1, len(diff.Modified))
	assert.Contains(s.T(), diff.Modified, "nested")

	// Changes should include the modified nested map
	assert.Equal(s.T(), 1, len(diff.Changes))
	expectedNested := map[string]interface{}{
		"key1": "modified_value",
		"key3": "new_value",
	}
	assert.Equal(s.T(), expectedNested, diff.Changes["nested"])

	// Should not be empty
	assert.False(s.T(), diff.IsEmpty())
}

func (s *ComputeDiffTestSuite) TestComputeFlatDiff_EmptyDesired() {
	current := map[string]interface{}{
		"section1": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		"section2": map[string]interface{}{
			"key3": "value3",
		},
	}
	desired := map[string]interface{}{}

	diff := ComputeFlatDiff(current, desired)

	// When desired is empty, everything should be marked as removed
	assert.True(s.T(), len(diff.Removed) > 0)
	assert.Contains(s.T(), diff.Removed, "section1.key1")
	assert.Contains(s.T(), diff.Removed, "section1.key2")
	assert.Contains(s.T(), diff.Removed, "section2.key3")

	// Changes should include all removed items (flattened)
	assert.True(s.T(), len(diff.Changes) > 0)
	assert.Equal(s.T(), "value1", diff.Changes["section1.key1"])
	assert.Equal(s.T(), "value2", diff.Changes["section1.key2"])
	assert.Equal(s.T(), "value3", diff.Changes["section2.key3"])

	// Should not be empty
	assert.False(s.T(), diff.IsEmpty())
}

func (s *ComputeDiffTestSuite) TestComputeFlatDiff_WithDeletedKeys() {
	current := map[string]interface{}{
		"section1": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}
	desired := map[string]interface{}{
		"section1": map[string]interface{}{
			"key1": map[string]interface{}{
				"deleted": true,
			},
			"key2": "modified_value",
		},
	}

	diff := ComputeFlatDiff(current, desired)

	// key1 should be removed (deleted=true)
	assert.Contains(s.T(), diff.Removed, "section1.key1")
	assert.Equal(s.T(), "value1", diff.Changes["section1.key1"])

	// key2 should be modified
	assert.Contains(s.T(), diff.Modified, "section1.key2")
	assert.Equal(s.T(), "modified_value", diff.Changes["section1.key2"])

	// Should not be empty
	assert.False(s.T(), diff.IsEmpty())
}

func TestComputeDiffTestSuite(t *testing.T) {
	suite.Run(t, new(ComputeDiffTestSuite))
}
