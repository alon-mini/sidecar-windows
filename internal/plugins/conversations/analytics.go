package conversations

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/marcus/sidecar/internal/adapter/claudecode"
	"github.com/marcus/sidecar/internal/styles"
)

// renderAnalytics renders the global analytics view with scrolling support.
func (p *Plugin) renderAnalytics() string {
	// Build all content lines first
	var lines []string

	// Load stats
	stats, err := claudecode.LoadStatsCache()
	if err != nil {
		lines = append(lines, styles.Title.Render(" Usage Analytics"))
		lines = append(lines, styles.Muted.Render(strings.Repeat("━", p.width-2)))
		lines = append(lines, styles.StatusDeleted.Render(" Unable to load stats: "+err.Error()))
		p.analyticsLines = lines
		return strings.Join(lines, "\n")
	}

	// Header
	lines = append(lines, styles.Title.Render(" Usage Analytics"))
	lines = append(lines, styles.Muted.Render(strings.Repeat("━", p.width-2)))

	// Summary line
	firstDate := stats.FirstSessionDate.Format("Jan 2")
	summary := fmt.Sprintf(" Since %s  │  %d sessions  │  %s messages",
		firstDate,
		stats.TotalSessions,
		formatLargeNumber(stats.TotalMessages))
	lines = append(lines, styles.Body.Render(summary))
	lines = append(lines, "")

	// Weekly activity chart
	lines = append(lines, styles.Title.Render(" This Week's Activity"))
	lines = append(lines, styles.Muted.Render(strings.Repeat("─", p.width-2)))

	recentActivity := stats.GetRecentActivity(7)
	maxMsgs := 0
	for _, day := range recentActivity {
		if day.MessageCount > maxMsgs {
			maxMsgs = day.MessageCount
		}
	}

	for _, day := range recentActivity {
		date, _ := time.Parse("2006-01-02", day.Date)
		dayName := date.Format("Mon")
		bar := renderColoredBar(day.MessageCount, maxMsgs, 16)
		dayLabel := styles.Body.Render(fmt.Sprintf(" %s │ ", dayName))
		statsLabel := styles.Subtitle.Render(fmt.Sprintf(" │ %5d msgs │ %2d sessions", day.MessageCount, day.SessionCount))
		lines = append(lines, dayLabel+bar+statsLabel)
	}
	lines = append(lines, "")

	// Model usage
	lines = append(lines, styles.Title.Render(" Model Usage"))
	lines = append(lines, styles.Muted.Render(strings.Repeat("─", p.width-2)))

	// Sort models by total tokens descending for stable ordering
	type modelEntry struct {
		name        string
		usage       claudecode.ModelUsage
		totalTokens int64
	}
	var models []modelEntry
	var maxTokens int64
	for model, usage := range stats.ModelUsage {
		total := int64(usage.InputTokens) + int64(usage.OutputTokens)
		if total > maxTokens {
			maxTokens = total
		}
		models = append(models, modelEntry{model, usage, total})
	}
	sort.Slice(models, func(i, j int) bool {
		if models[i].totalTokens != models[j].totalTokens {
			return models[i].totalTokens > models[j].totalTokens // descending by tokens
		}
		return models[i].name < models[j].name // ascending by name for ties
	})

	for _, m := range models {
		shortName := modelShortName(m.name)
		if shortName == "" {
			continue
		}

		bar := renderColoredBar64(m.totalTokens, maxTokens, 12)
		cost := claudecode.CalculateModelCost(m.name, m.usage)

		modelLabel := styles.Body.Render(fmt.Sprintf(" %-6s │ ", shortName))
		tokensLabel := styles.Subtitle.Render(fmt.Sprintf(" │ %s in  %s out │ ",
			formatLargeNumber64(int64(m.usage.InputTokens)),
			formatLargeNumber64(int64(m.usage.OutputTokens))))
		costLabel := lipgloss.NewStyle().Foreground(styles.Accent).Render(fmt.Sprintf("~$%.0f", cost))
		lines = append(lines, modelLabel+bar+tokensLabel+costLabel)
	}
	lines = append(lines, "")

	// Stats footer
	cacheEff := stats.CacheEfficiency()
	cacheLabel := styles.Subtitle.Render(" Cache Efficiency: ")
	cacheValue := lipgloss.NewStyle().Foreground(styles.Success).Render(fmt.Sprintf("%.0f%%", cacheEff))
	lines = append(lines, cacheLabel+cacheValue)

	// Peak hours
	peakHours := stats.GetPeakHours(3)
	if len(peakHours) > 0 {
		peakLabel := styles.Subtitle.Render(" Peak Hours:")
		peakValues := ""
		for i, ph := range peakHours {
			if i > 0 {
				peakValues += ","
			}
			peakValues += fmt.Sprintf(" %s:00", ph.Hour)
		}
		lines = append(lines, peakLabel+styles.Body.Render(peakValues))
	}

	// Longest session
	if stats.LongestSession.Duration > 0 {
		dur := time.Duration(stats.LongestSession.Duration) * time.Millisecond
		sessionLabel := styles.Subtitle.Render(" Longest Session: ")
		sessionValue := styles.Body.Render(formatSessionDuration(dur))
		lines = append(lines, sessionLabel+sessionValue)
	}

	// Total cost
	totalCost := stats.TotalCost()
	costLabel := styles.Subtitle.Render(" Total Estimated Cost: ")
	costValue := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true).Render(fmt.Sprintf("~$%.0f", totalCost))
	lines = append(lines, costLabel+costValue)

	// Store lines for scroll calculation
	p.analyticsLines = lines

	// Apply scroll offset and height constraint
	contentHeight := p.height - 2 // leave room for potential padding
	if contentHeight < 1 {
		contentHeight = 1
	}

	start := p.analyticsScrollOff
	if start >= len(lines) {
		start = len(lines) - 1
		if start < 0 {
			start = 0
		}
	}
	end := start + contentHeight
	if end > len(lines) {
		end = len(lines)
	}

	visibleLines := lines[start:end]
	return strings.Join(visibleLines, "\n")
}

// renderColoredBar renders a colored ASCII bar chart segment.
func renderColoredBar(value, max, width int) string {
	if max == 0 {
		return styles.Muted.Render(strings.Repeat("░", width))
	}
	filled := (value * width) / max
	if filled > width {
		filled = width
	}
	filledBar := lipgloss.NewStyle().Foreground(styles.Primary).Render(strings.Repeat("█", filled))
	emptyBar := styles.Muted.Render(strings.Repeat("░", width-filled))
	return filledBar + emptyBar
}

// renderColoredBar64 renders a colored ASCII bar chart segment for int64 values.
func renderColoredBar64(value, max int64, width int) string {
	if max == 0 {
		return styles.Muted.Render(strings.Repeat("░", width))
	}
	filled := int((value * int64(width)) / max)
	if filled > width {
		filled = width
	}
	filledBar := lipgloss.NewStyle().Foreground(styles.Secondary).Render(strings.Repeat("█", filled))
	emptyBar := styles.Muted.Render(strings.Repeat("░", width-filled))
	return filledBar + emptyBar
}

// formatLargeNumber formats a number with K/M suffix.
func formatLargeNumber(n int) string {
	return formatLargeNumber64(int64(n))
}

// formatLargeNumber64 formats an int64 with K/M/B suffix.
func formatLargeNumber64(n int64) string {
	if n >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
