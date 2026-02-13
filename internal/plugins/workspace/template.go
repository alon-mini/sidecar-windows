package workspace

import (
	"regexp"
)

// ticketPattern matches {{ticket}} or {{ticket || 'fallback text'}}
var ticketPattern = regexp.MustCompile(`\{\{ticket(?:\s*\|\|\s*'([^']*)')?\}\}`)

// ExpandPromptTemplate expands template variables in a prompt body.
// - {{ticket}} expands to taskID (returns empty if taskID is empty)
// - {{ticket || 'default'}} expands to taskID, or 'default' if taskID is empty
func ExpandPromptTemplate(body, taskID string) string {
	return ticketPattern.ReplaceAllStringFunc(body, func(match string) string {
		submatch := ticketPattern.FindStringSubmatch(match)

		if taskID != "" {
			return taskID
		}

		// Check for fallback value
		if len(submatch) > 1 && submatch[1] != "" {
			return submatch[1]
		}

		// No fallback, return empty
		return ""
	})
}
