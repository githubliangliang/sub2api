package claude

import "strings"

// RejectsTemperature identifies the existing Claude families that reject sampling parameters.
func RejectsTemperature(model string) bool {
	id := normalizeEffortModelID(model)
	for _, family := range []string{"claude-opus-4-7", "claude-opus-4-8", "claude-opus-5-5"} {
		if id == family || strings.HasPrefix(id, family+"-") {
			return true
		}
	}
	return false
}
