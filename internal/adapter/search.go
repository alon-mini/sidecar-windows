// Package adapter provides interfaces and types for AI session data sources.
// This file defines the optional MessageSearcher interface for adapters with
// cross-conversation search capability.

package adapter

import "time"

// MessageSearcher is an optional interface for adapters that support searching
// message content across sessions. Adapters implement this to enable the
// cross-conversation search feature.
type MessageSearcher interface {
	// SearchMessages searches message content within a session.
	// Returns matches with context, limited by opts.MaxResults.
	// The query is interpreted as regex if opts.UseRegex is true,
	// otherwise as a literal substring.
	SearchMessages(sessionID, query string, opts SearchOptions) ([]MessageMatch, error)
}

// SearchOptions configures how message search is performed.
type SearchOptions struct {
	// UseRegex treats the query as a regular expression if true,
	// otherwise as a literal substring match.
	UseRegex bool

	// CaseSensitive enables case-sensitive matching if true.
	// Defaults to case-insensitive when false.
	CaseSensitive bool

	// MaxResults limits matches per session. Zero uses default (50).
	MaxResults int
}

// DefaultMaxResults is the default per-session match limit.
const DefaultMaxResults = 50

// DefaultSearchOptions returns sensible defaults for search.
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		UseRegex:      false,
		CaseSensitive: false,
		MaxResults:    DefaultMaxResults,
	}
}

// MessageMatch represents a message containing one or more search matches.
type MessageMatch struct {
	// MessageID is the unique identifier of the matched message.
	MessageID string

	// MessageIdx is the zero-based index of the message in the session.
	MessageIdx int

	// Role is the message author (e.g., "user", "assistant").
	Role string

	// Timestamp is when the message was created.
	Timestamp time.Time

	// Model is the AI model used for this message (may be empty for user messages).
	Model string

	// Matches contains the specific content matches within this message.
	Matches []ContentMatch
}

// ContentMatch represents a single match location within message content.
type ContentMatch struct {
	// BlockType identifies the content block type:
	// "text", "tool_use", "tool_result", or "thinking".
	BlockType string

	// LineNo is the 1-indexed line number within the block.
	LineNo int

	// LineText is the full line containing the match.
	LineText string

	// ColStart is the 0-indexed byte position where the match begins in LineText.
	ColStart int

	// ColEnd is the 0-indexed byte position where the match ends in LineText.
	ColEnd int
}

// TotalMatches returns the total number of content matches across all messages.
func TotalMatches(matches []MessageMatch) int {
	count := 0
	for _, m := range matches {
		count += len(m.Matches)
	}
	return count
}
