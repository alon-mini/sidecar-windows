// Package adapter provides shared utilities for message search across adapters.
package adapter

import (
	"fmt"
	"regexp"
	"strings"
)

// CompileSearchPattern compiles a search pattern based on options.
// Returns a compiled regexp for matching, or error if pattern is invalid.
func CompileSearchPattern(query string, opts SearchOptions) (*regexp.Regexp, error) {
	pattern := query
	if !opts.UseRegex {
		pattern = regexp.QuoteMeta(query)
	}
	flags := ""
	if !opts.CaseSensitive {
		flags = "(?i)"
	}
	return regexp.Compile(flags + pattern)
}

// SearchContent searches text content line-by-line and returns matches.
// Returns all matches found in content with line numbers and positions.
func SearchContent(content string, blockType string, re *regexp.Regexp) []ContentMatch {
	if content == "" {
		return nil
	}

	var matches []ContentMatch
	lines := strings.Split(content, "\n")
	for lineNo, line := range lines {
		locs := re.FindAllStringIndex(line, -1)
		for _, loc := range locs {
			matches = append(matches, ContentMatch{
				BlockType: blockType,
				LineNo:    lineNo + 1, // 1-indexed
				LineText:  line,
				ColStart:  loc[0],
				ColEnd:    loc[1],
			})
		}
	}
	return matches
}

// SearchMessage searches all content in a message and returns matches.
// Searches across text content, tool uses, tool results, and thinking blocks.
func SearchMessage(msg *Message, msgIdx int, re *regexp.Regexp, maxResults int, currentTotal int) *MessageMatch {
	var allMatches []ContentMatch

	// Search main content
	if msg.Content != "" {
		allMatches = append(allMatches, SearchContent(msg.Content, "text", re)...)
	}

	// Search content blocks (covers text, tool_use, tool_result, thinking)
	for _, cb := range msg.ContentBlocks {
		switch cb.Type {
		case "text":
			if cb.Text != "" {
				allMatches = append(allMatches, SearchContent(cb.Text, "text", re)...)
			}
		case "thinking":
			if cb.Text != "" {
				allMatches = append(allMatches, SearchContent(cb.Text, "thinking", re)...)
			}
		case "tool_use":
			if cb.ToolName != "" {
				allMatches = append(allMatches, SearchContent(cb.ToolName, "tool_use", re)...)
			}
			if cb.ToolInput != "" {
				allMatches = append(allMatches, SearchContent(cb.ToolInput, "tool_use", re)...)
			}
		case "tool_result":
			if cb.ToolOutput != "" {
				allMatches = append(allMatches, SearchContent(cb.ToolOutput, "tool_result", re)...)
			}
		}
	}

	// Search tool uses directly (for adapters that populate ToolUses)
	for _, tu := range msg.ToolUses {
		if tu.Name != "" {
			allMatches = append(allMatches, SearchContent(tu.Name, "tool_use", re)...)
		}
		if tu.Input != "" {
			allMatches = append(allMatches, SearchContent(tu.Input, "tool_use", re)...)
		}
		if tu.Output != "" {
			allMatches = append(allMatches, SearchContent(tu.Output, "tool_result", re)...)
		}
	}

	// Search thinking blocks directly (for adapters that populate ThinkingBlocks)
	for _, tb := range msg.ThinkingBlocks {
		if tb.Content != "" {
			allMatches = append(allMatches, SearchContent(tb.Content, "thinking", re)...)
		}
	}

	// Deduplicate matches (same line might be searched multiple times)
	allMatches = deduplicateMatches(allMatches)

	if len(allMatches) == 0 {
		return nil
	}

	// Limit matches if we're approaching maxResults
	if maxResults > 0 && currentTotal+len(allMatches) > maxResults {
		remaining := maxResults - currentTotal
		if remaining <= 0 {
			return nil
		}
		allMatches = allMatches[:remaining]
	}

	return &MessageMatch{
		MessageID:  msg.ID,
		MessageIdx: msgIdx,
		Role:       msg.Role,
		Timestamp:  msg.Timestamp,
		Model:      msg.Model,
		Matches:    allMatches,
	}
}

// deduplicateMatches removes duplicate matches (same position, same text).
func deduplicateMatches(matches []ContentMatch) []ContentMatch {
	if len(matches) <= 1 {
		return matches
	}

	seen := make(map[string]struct{})
	result := make([]ContentMatch, 0, len(matches))

	for _, m := range matches {
		// Create a unique key for this match
		key := fmt.Sprintf("%s|%d|%d|%d|%s", m.BlockType, m.LineNo, m.ColStart, m.ColEnd, m.LineText)
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			result = append(result, m)
		}
	}

	return result
}

// SearchMessagesSlice is a helper that searches a slice of messages.
// Returns MessageMatch results up to opts.MaxResults.
func SearchMessagesSlice(messages []Message, query string, opts SearchOptions) ([]MessageMatch, error) {
	maxResults := opts.MaxResults
	if maxResults <= 0 {
		maxResults = DefaultMaxResults
	}

	re, err := CompileSearchPattern(query, opts)
	if err != nil {
		return nil, err
	}

	var results []MessageMatch
	totalMatches := 0

	for idx, msg := range messages {
		if maxResults > 0 && totalMatches >= maxResults {
			break
		}

		match := SearchMessage(&msg, idx, re, maxResults, totalMatches)
		if match != nil {
			results = append(results, *match)
			totalMatches += len(match.Matches)
		}
	}

	return results, nil
}
