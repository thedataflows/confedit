package state

type Manager struct{}

func NewManager(stateDir string) *Manager {
	// stateDir parameter kept for API compatibility but not used
	return &Manager{}
}

func (m *Manager) ComputeDiffWithCurrent(target string, desired map[string]interface{}, currentSystemState map[string]interface{}) (*ConfigDiff, error) {
	if currentSystemState == nil {
		currentSystemState = make(map[string]interface{})
	}

	// Filter current state to only include keys that are managed (present in desired state)
	// This prevents unmanaged keys (like extra commented keys in INI files) from causing false diffs
	filteredCurrent := m.filterManagedKeys(currentSystemState, desired)

	// Compute diff between filtered current system state and desired state
	// Use flattened diff for better display of nested structures
	diff := ComputeFlatDiff(filteredCurrent, desired)
	diff.Target = target

	return diff, nil
}

// filterManagedKeys filters the current state to only include keys that are managed (present in desired state)
// This prevents unmanaged keys (like extra commented keys in INI files) from causing false diffs
func (m *Manager) filterManagedKeys(current, desired map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})
	rootSection := m.extractRootSection(current)

	// Only include keys from current that are also present in desired (managed keys)
	for key, value := range desired {
		if currentValue, exists := m.findKeyInCurrent(key, current, rootSection); exists {
			filtered[key] = m.filterKeyValue(currentValue, value)
		}
		// If key doesn't exist in current, it will be detected as "added" by ComputeDiff
	}

	return filtered
}

// extractRootSection extracts the root section from current state for INI format support
func (m *Manager) extractRootSection(current map[string]interface{}) map[string]interface{} {
	if rootSectionValue, hasRootSection := current[""]; hasRootSection {
		if rootMap, ok := rootSectionValue.(map[string]interface{}); ok {
			return rootMap
		}
	}
	return nil
}

// findKeyInCurrent tries to find a key in current state, checking both direct and root section
func (m *Manager) findKeyInCurrent(key string, current map[string]interface{}, rootSection map[string]interface{}) (interface{}, bool) {
	// First, try to find the key directly in current
	if currentValue, exists := current[key]; exists {
		return currentValue, true
	}

	// If not found directly and we have a root section, check there
	if rootSection != nil {
		if currentValue, exists := rootSection[key]; exists {
			return currentValue, true
		}
	}

	return nil, false
}

// filterKeyValue handles filtering of individual key-value pairs, with support for nested structures
func (m *Manager) filterKeyValue(currentValue, desiredValue interface{}) interface{} {
	// For nested structures (like INI sections), recursively filter
	if currentMap, ok := currentValue.(map[string]interface{}); ok {
		if desiredMap, ok := desiredValue.(map[string]interface{}); ok {
			// Recursively filter the nested structure
			return m.filterManagedKeys(currentMap, desiredMap)
		}
		// Type mismatch - include the current value as-is for diff detection
		return currentValue
	}

	// Leaf value - include as-is
	return currentValue
}
