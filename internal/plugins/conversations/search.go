package conversations

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/marcus/sidecar/internal/adapter"
)

// SearchFilters holds multi-dimensional filter criteria.
type SearchFilters struct {
	Query      string    // Text search
	Adapters   []string  // ["claude-code", "codex"]
	Models     []string  // ["opus", "sonnet", "haiku"]
	Categories []string  // ["interactive", "cron", "system"]
	DateRange  DateRange // today, week, custom
	MinTokens  int       // Sessions with > N tokens
	MaxTokens  int       // Sessions with < N tokens
	ActiveOnly bool      // Only currently active
	HasFiles   []string  // Sessions that touched these files
}

// DateRange represents a date range filter.
type DateRange struct {
	Preset string    // "today", "yesterday", "week", "month", "all"
	Start  time.Time // For custom range
	End    time.Time
}

// IsActive returns true if any filter is active.
func (f *SearchFilters) IsActive() bool {
	return f.Query != "" ||
		len(f.Adapters) > 0 ||
		len(f.Models) > 0 ||
		len(f.Categories) > 0 ||
		f.DateRange.Preset != "" ||
		f.MinTokens > 0 ||
		f.MaxTokens > 0 ||
		f.ActiveOnly ||
		len(f.HasFiles) > 0
}

// ToggleAdapter toggles an adapter in the filter list.
func (f *SearchFilters) ToggleAdapter(adapterID string) {
	for i, id := range f.Adapters {
		if id == adapterID {
			f.Adapters = append(f.Adapters[:i], f.Adapters[i+1:]...)
			return
		}
	}
	f.Adapters = append(f.Adapters, adapterID)
}

// HasAdapter returns true if the adapter is in the filter list.
func (f *SearchFilters) HasAdapter(adapterID string) bool {
	for _, id := range f.Adapters {
		if id == adapterID {
			return true
		}
	}
	return false
}

// ToggleModel toggles a model in the filter list.
func (f *SearchFilters) ToggleModel(model string) {
	for i, m := range f.Models {
		if m == model {
			f.Models = append(f.Models[:i], f.Models[i+1:]...)
			return
		}
	}
	f.Models = append(f.Models, model)
}

// HasModel returns true if the model is in the filter list.
func (f *SearchFilters) HasModel(model string) bool {
	for _, m := range f.Models {
		if m == model {
			return true
		}
	}
	return false
}

// ToggleCategory toggles a category in the filter list.
func (f *SearchFilters) ToggleCategory(cat string) {
	for i, c := range f.Categories {
		if c == cat {
			f.Categories = append(f.Categories[:i], f.Categories[i+1:]...)
			return
		}
	}
	f.Categories = append(f.Categories, cat)
}

// HasCategory returns true if the category is in the filter list.
func (f *SearchFilters) HasCategory(cat string) bool {
	return slices.Contains(f.Categories, cat)
}

// SetDateRange sets the date range preset.
func (f *SearchFilters) SetDateRange(preset string) {
	if f.DateRange.Preset == preset {
		f.DateRange.Preset = "" // Toggle off
		return
	}
	f.DateRange.Preset = preset

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch preset {
	case "today":
		f.DateRange.Start = today
		f.DateRange.End = now
	case "yesterday":
		f.DateRange.Start = today.AddDate(0, 0, -1)
		f.DateRange.End = today
	case "week":
		f.DateRange.Start = today.AddDate(0, 0, -7)
		f.DateRange.End = now
	case "month":
		f.DateRange.Start = today.AddDate(0, -1, 0)
		f.DateRange.End = now
	default:
		f.DateRange.Start = time.Time{}
		f.DateRange.End = time.Time{}
	}
}

// Matches checks if a session matches all filter criteria.
func (f *SearchFilters) Matches(session adapter.Session) bool {
	// Text search
	if f.Query != "" {
		query := strings.ToLower(f.Query)
		if !strings.Contains(strings.ToLower(session.Name), query) &&
			!strings.Contains(strings.ToLower(session.Slug), query) &&
			!strings.Contains(session.ID, query) &&
			!strings.Contains(strings.ToLower(session.AdapterName), query) {
			return false
		}
	}

	// Adapter filter
	if len(f.Adapters) > 0 && !f.HasAdapter(session.AdapterID) {
		return false
	}

	// Category filter (only applies to sessions with a category set;
	// non-Pi adapters leave SessionCategory empty and always pass through)
	if len(f.Categories) > 0 && session.SessionCategory != "" && !f.HasCategory(session.SessionCategory) {
		return false
	}

	// Model filter - Would need session.PrimaryModel field to check this
	// For now, skip model filtering at session level

	// Date range filter
	if f.DateRange.Preset != "" {
		if session.UpdatedAt.Before(f.DateRange.Start) || session.UpdatedAt.After(f.DateRange.End) {
			return false
		}
	}

	// Token filters
	if f.MinTokens > 0 && session.TotalTokens < f.MinTokens {
		return false
	}
	if f.MaxTokens > 0 && session.TotalTokens > f.MaxTokens {
		return false
	}

	// Active only filter
	if f.ActiveOnly && !session.IsActive {
		return false
	}

	return true
}

// String formats active filters for display.
func (f *SearchFilters) String() string {
	var parts []string

	if len(f.Models) > 0 {
		parts = append(parts, "[model:"+strings.Join(f.Models, ",")+"]")
	}
	if len(f.Adapters) > 0 {
		parts = append(parts, "[adapter:"+strings.Join(f.Adapters, ",")+"]")
	}
	if len(f.Categories) > 0 {
		parts = append(parts, "[category:"+strings.Join(f.Categories, ",")+"]")
	}
	if f.DateRange.Preset != "" {
		parts = append(parts, "["+f.DateRange.Preset+"]")
	}
	if f.MinTokens > 0 {
		parts = append(parts, "[tokens:>"+formatTokenCount(f.MinTokens)+"]")
	}
	if f.MaxTokens > 0 {
		parts = append(parts, "[tokens:<"+formatTokenCount(f.MaxTokens)+"]")
	}
	if f.ActiveOnly {
		parts = append(parts, "[active]")
	}

	return strings.Join(parts, " ")
}

// formatTokenCount formats a token count for display.
func formatTokenCount(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.0fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.0fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}
