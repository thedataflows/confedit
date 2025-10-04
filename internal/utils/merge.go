package utils

// DeepMerge recursively merges newContent into existing, modifying existing in-place.
// This performs a deep merge down to all nested levels, preserving all keys in existing
// while adding/updating keys from newContent.
func DeepMerge(existing map[string]interface{}, newContent map[string]interface{}) error {
	for key, newValue := range newContent {
		if existingValue, exists := existing[key]; exists {
			// Both values exist, need to merge them
			existingMap, existingIsMap := existingValue.(map[string]interface{})
			newMap, newIsMap := newValue.(map[string]interface{})

			if existingIsMap && newIsMap {
				// Both are maps, merge recursively down to all levels
				if err := DeepMerge(existingMap, newMap); err != nil {
					return err
				}
			} else {
				// One or both are not maps (leaf values), new value overrides existing
				existing[key] = newValue
			}
		} else {
			// Key doesn't exist in existing, add it
			existing[key] = newValue
		}
	}
	return nil
}
