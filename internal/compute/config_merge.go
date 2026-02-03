package compute

import "encoding/json"

// MergeConfigMaps merges override onto base recursively for map values.
// Non-map values in override replace base values.
func MergeConfigMaps(base map[string]interface{}, override map[string]interface{}) map[string]interface{} {
	if base == nil && override == nil {
		return nil
	}

	result := map[string]interface{}{}
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		if overrideMap, ok := v.(map[string]interface{}); ok {
			if baseMap, ok := result[k].(map[string]interface{}); ok {
				result[k] = MergeConfigMaps(baseMap, overrideMap)
				continue
			}
		}
		result[k] = v
	}

	return result
}

// MergeConfigJSON merges a base config map with an override JSON blob.
// Returns the merged configuration as JSON.
func MergeConfigJSON(base map[string]interface{}, override json.RawMessage) (json.RawMessage, error) {
	if len(base) == 0 && len(override) == 0 {
		return nil, nil
	}

	var overrideMap map[string]interface{}
	if len(override) > 0 {
		if err := json.Unmarshal(override, &overrideMap); err != nil {
			return nil, err
		}
	}

	merged := MergeConfigMaps(base, overrideMap)
	if len(merged) == 0 {
		return nil, nil
	}

	return json.Marshal(merged)
}
