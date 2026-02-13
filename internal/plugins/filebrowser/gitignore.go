package filebrowser

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GitIgnore manages .gitignore patterns for file filtering.
type GitIgnore struct {
	patterns []gitIgnorePattern
	cache    map[string]bool // Path -> isIgnored cache
}

type gitIgnorePattern struct {
	pattern  string
	negate   bool // Starts with !
	dirOnly  bool // Ends with /
	anchored bool // Contains / (not at end)
	regex    *regexp.Regexp
}

// NewGitIgnore creates a new GitIgnore instance.
func NewGitIgnore() *GitIgnore {
	return &GitIgnore{
		cache: make(map[string]bool),
	}
}

// LoadFile loads patterns from a .gitignore file.
func (gi *GitIgnore) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No gitignore is fine
		}
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		gi.addPattern(line)
	}
	return nil
}

// addPattern parses and adds a gitignore pattern.
func (gi *GitIgnore) addPattern(line string) {
	p := gitIgnorePattern{pattern: line}

	// Check for negation
	if strings.HasPrefix(line, "!") {
		p.negate = true
		line = line[1:]
	}

	// Check for directory only
	if strings.HasSuffix(line, "/") {
		p.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}

	// Check if anchored (contains / not at end)
	p.anchored = strings.Contains(line, "/")
	if p.anchored {
		line = strings.TrimPrefix(line, "/")
	}

	// Convert glob to regex
	regex := globToRegex(line, p.anchored)
	compiled, err := regexp.Compile(regex)
	if err != nil {
		return // Skip invalid patterns
	}
	p.regex = compiled

	gi.patterns = append(gi.patterns, p)
}

// globToRegex converts a gitignore glob pattern to a regex.
func globToRegex(glob string, anchored bool) string {
	var sb strings.Builder
	sb.WriteString("^")

	if !anchored {
		sb.WriteString("(.*/)?") // Match any leading path
	}

	i := 0
	for i < len(glob) {
		c := glob[i]
		switch c {
		case '*':
			if i+1 < len(glob) && glob[i+1] == '*' {
				// ** matches everything including /
				sb.WriteString(".*")
				i++ // Skip second *
			} else {
				// * matches everything except /
				sb.WriteString("[^/]*")
			}
		case '?':
			sb.WriteString("[^/]")
		case '.', '+', '^', '$', '(', ')', '[', ']', '{', '}', '|', '\\':
			sb.WriteByte('\\')
			sb.WriteByte(c)
		default:
			sb.WriteByte(c)
		}
		i++
	}

	sb.WriteString("$")
	return sb.String()
}

// IsIgnored checks if a path matches any gitignore pattern.
func (gi *GitIgnore) IsIgnored(path string, isDir bool) bool {
	// Normalize path separators
	path = filepath.ToSlash(path)

	// Check cache (key includes isDir since dir-only patterns differ)
	cacheKey := path
	if isDir {
		cacheKey = path + "/"
	}
	if cached, ok := gi.cache[cacheKey]; ok {
		return cached
	}

	ignored := false
	for _, p := range gi.patterns {
		if p.dirOnly && !isDir {
			continue
		}
		if p.matches(path) {
			ignored = !p.negate
		}
	}

	gi.cache[cacheKey] = ignored
	return ignored
}

// matches checks if a path matches this pattern.
func (p *gitIgnorePattern) matches(path string) bool {
	if p.regex == nil {
		return false
	}

	// Try matching the full path
	if p.regex.MatchString(path) {
		return true
	}

	// Also try matching just the filename for non-anchored patterns
	if !p.anchored {
		base := filepath.Base(path)
		return p.regex.MatchString(base)
	}

	return false
}

// ClearCache clears the path cache.
func (gi *GitIgnore) ClearCache() {
	gi.cache = make(map[string]bool)
}
