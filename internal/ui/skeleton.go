package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcus/sidecar/internal/styles"
)

// SkeletonTickMsg is sent to update the shimmer animation frame.
type SkeletonTickMsg time.Time

// SkeletonTickInterval is the animation frame rate (60fps feel).
const SkeletonTickInterval = 80 * time.Millisecond

// Skeleton renders animated placeholder rows with a shimmer effect.
type Skeleton struct {
	// Configuration
	Rows      int   // Number of skeleton rows to display
	RowWidths []int // Width pattern for each row (cycles if fewer than Rows)

	// Animation state
	frame    int  // Current animation frame
	active   bool // Whether animation is running
	shimmerW int  // Width of the shimmer highlight
}

// NewSkeleton creates a skeleton loader with the given row count.
// RowWidths defines the relative width of each row (e.g., []int{80, 60, 75} for varied lengths).
// If rowWidths is nil, uses a default varied pattern.
func NewSkeleton(rows int, rowWidths []int) Skeleton {
	if rowWidths == nil {
		// Default pattern: varied row lengths for realistic look
		rowWidths = []int{85, 60, 75, 55, 80, 65, 70, 50}
	}
	return Skeleton{
		Rows:      rows,
		RowWidths: rowWidths,
		active:    true,
		shimmerW:  6, // Width of the bright shimmer band
	}
}

// Start begins the shimmer animation.
func (s *Skeleton) Start() tea.Cmd {
	s.active = true
	return s.tick()
}

// Stop halts the shimmer animation.
func (s *Skeleton) Stop() {
	s.active = false
}

// IsActive returns whether the animation is running.
func (s Skeleton) IsActive() bool {
	return s.active
}

// Update handles tick messages for animation.
func (s *Skeleton) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case SkeletonTickMsg:
		if !s.active {
			return nil
		}
		s.frame++
		return s.tick()
	}
	return nil
}

// tick schedules the next animation frame.
func (s *Skeleton) tick() tea.Cmd {
	return tea.Tick(SkeletonTickInterval, func(t time.Time) tea.Msg {
		return SkeletonTickMsg(t)
	})
}

// View renders the skeleton rows with shimmer effect.
// width is the available content width.
func (s Skeleton) View(width int) string {
	if width < 10 {
		width = 10
	}

	var sb strings.Builder

	// Shimmer position cycles across the width
	// Use frame to create a wave effect moving left-to-right
	cycleLen := width + s.shimmerW*2
	shimmerStart := s.frame % cycleLen

	// Styles for dim and bright states
	dimStyle := lipgloss.NewStyle().Foreground(styles.TextSubtle)
	brightStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)

	for row := range s.Rows {
		// Get row width from pattern (cycle if needed)
		widthPct := s.RowWidths[row%len(s.RowWidths)]
		rowWidth := min(max((width*widthPct)/100, 5), width)

		// Render line with shimmer effect
		line := s.renderShimmerLine(rowWidth, shimmerStart+row*2, cycleLen, dimStyle, brightStyle)
		sb.WriteString(line)
		if row < s.Rows-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// renderShimmerLine renders a single line with shimmer effect.
func (s Skeleton) renderShimmerLine(width, shimmerPos, cycleLen int, dimStyle, brightStyle lipgloss.Style) string {
	const (
		charDim    = "░"
		charBright = "▒"
	)

	shimmerPos = shimmerPos % cycleLen

	var parts []string
	inShimmer := false
	segmentStart := 0

	for col := 0; col <= width; col++ {
		distFromShimmer := col - (shimmerPos - s.shimmerW)
		nowInShimmer := distFromShimmer >= 0 && distFromShimmer < s.shimmerW && col < width

		if col == width || nowInShimmer != inShimmer {
			// Emit segment
			segmentLen := col - segmentStart
			if segmentLen > 0 {
				segment := strings.Repeat(charDim, segmentLen)
				if inShimmer {
					segment = strings.Repeat(charBright, segmentLen)
					parts = append(parts, brightStyle.Render(segment))
				} else {
					parts = append(parts, dimStyle.Render(segment))
				}
			}
			segmentStart = col
			inShimmer = nowInShimmer
		}
	}

	return strings.Join(parts, "")
}

// SkeletonTick returns a command to start/continue the skeleton animation.
// Call this from your plugin's Start() or when enabling the skeleton.
func SkeletonTick() tea.Cmd {
	return tea.Tick(SkeletonTickInterval, func(t time.Time) tea.Msg {
		return SkeletonTickMsg(t)
	})
}
