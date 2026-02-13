package community

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
)

//go:embed schemes.json
var schemesData []byte

// SchemeIndex is a sorted list of all scheme names.
var SchemeIndex []string

// SchemeMap provides O(1) lookup by name.
var SchemeMap map[string]*CommunityScheme

func init() {
	var schemes []CommunityScheme
	if err := json.Unmarshal(schemesData, &schemes); err != nil {
		panic(fmt.Sprintf("community: failed to parse embedded schemes.json: %v", err))
	}

	SchemeMap = make(map[string]*CommunityScheme, len(schemes))
	SchemeIndex = make([]string, 0, len(schemes))

	for i := range schemes {
		s := &schemes[i]
		SchemeMap[s.Name] = s
		SchemeIndex = append(SchemeIndex, s.Name)
	}
	sort.Strings(SchemeIndex)
}

// ListSchemes returns a sorted list of all available scheme names.
func ListSchemes() []string {
	return SchemeIndex
}

// GetScheme returns a scheme by name, or nil if not found.
func GetScheme(name string) *CommunityScheme {
	return SchemeMap[name]
}

// SchemeCount returns the number of available schemes.
func SchemeCount() int {
	return len(SchemeIndex)
}
