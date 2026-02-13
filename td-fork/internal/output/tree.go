package output

import (
	"fmt"
	"strings"

	"github.com/marcus/td/internal/models"
)

// TreeNode represents a node in a tree structure for rendering
type TreeNode struct {
	ID       string
	Title    string
	Type     models.Type
	Status   models.Status
	Children []TreeNode
}

// TreeRenderOptions configures tree rendering behavior
type TreeRenderOptions struct {
	MaxDepth    int  // 0 = unlimited
	ShowStatus  bool // Whether to show status indicator
	ShowType    bool // Whether to show issue type
	Indentation int  // Base indentation level (for nested contexts)
}

// statusMark returns a status indicator symbol
func statusMark(s models.Status) string {
	switch s {
	case models.StatusClosed:
		return " \u2713" // ✓
	case models.StatusInReview:
		return " \u29d7" // ⧗
	case models.StatusInProgress:
		return " \u25cf" // ●
	case models.StatusBlocked:
		return " \u2717" // ✗
	default:
		return ""
	}
}

// RenderTree renders a tree starting from a single root node
// Returns the complete tree as a string (without the root - just children)
func RenderTree(root TreeNode, opts TreeRenderOptions) string {
	lines := renderTreeNodes(root.Children, opts, 0, "")
	return strings.Join(lines, "\n")
}

// RenderTreeLines renders multiple root nodes and returns individual lines
// Useful for embedding trees in other output
func RenderTreeLines(roots []TreeNode, opts TreeRenderOptions) []string {
	return renderTreeNodes(roots, opts, 0, "")
}

// renderTreeNodes recursively renders tree nodes
func renderTreeNodes(nodes []TreeNode, opts TreeRenderOptions, depth int, prefix string) []string {
	if opts.MaxDepth > 0 && depth >= opts.MaxDepth {
		return nil
	}

	var lines []string

	for i, node := range nodes {
		isLast := i == len(nodes)-1

		// Build connector
		connector := "\u251c\u2500\u2500 " // ├──
		if isLast {
			connector = "\u2514\u2500\u2500 " // └──
		}

		// Build the line content
		var parts []string
		if opts.ShowType {
			parts = append(parts, string(node.Type))
		}
		parts = append(parts, node.ID+":")
		parts = append(parts, node.Title)

		if opts.ShowStatus {
			parts = append(parts, FormatStatus(node.Status))
			parts = append(parts, statusMark(node.Status))
		}

		line := prefix + connector + strings.Join(parts, " ")
		lines = append(lines, line)

		// Build prefix for children
		childPrefix := prefix
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "\u2502   " // │
		}

		// Recurse for children
		childLines := renderTreeNodes(node.Children, opts, depth+1, childPrefix)
		lines = append(lines, childLines...)
	}

	return lines
}

// RenderBlockedTree renders a blocked-by tree with "blocks:" header
// This matches the output format of printBlockedTree in dependencies.go
func RenderBlockedTree(nodes []TreeNode, opts TreeRenderOptions, directOnly bool) string {
	if len(nodes) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "\u2514\u2500\u2500 blocks:") // └── blocks:

	nodeLines := renderBlockedNodes(nodes, opts, 0, "    ", directOnly, make(map[string]bool))
	lines = append(lines, nodeLines...)

	return strings.Join(lines, "\n")
}

// renderBlockedNodes renders blocked tree nodes with visited tracking
func renderBlockedNodes(nodes []TreeNode, opts TreeRenderOptions, depth int, prefix string, directOnly bool, visited map[string]bool) []string {
	var lines []string

	for i, node := range nodes {
		if visited[node.ID] {
			continue
		}
		visited[node.ID] = true

		isLast := i == len(nodes)-1

		connector := "\u251c\u2500\u2500 " // ├──
		if isLast {
			connector = "\u2514\u2500\u2500 " // └──
		}

		line := fmt.Sprintf("%s%s%s: %s %s", prefix, connector, node.ID, node.Title, FormatStatus(node.Status))
		lines = append(lines, line)

		if !directOnly && len(node.Children) > 0 {
			childPrefix := prefix + "    "
			childLines := renderBlockedNodes(node.Children, opts, depth+1, childPrefix, directOnly, visited)
			lines = append(lines, childLines...)
		}
	}

	return lines
}

// RenderChildrenList renders children in a simple list format
// This matches the CHILDREN section output in show.go
func RenderChildrenList(nodes []TreeNode) []string {
	var lines []string

	for i, node := range nodes {
		isLast := i == len(nodes)-1

		connector := "\u251c\u2500\u2500 " // ├──
		if isLast {
			connector = "\u2514\u2500\u2500 " // └──
		}

		mark := statusMark(node.Status)

		line := fmt.Sprintf("  %s%s %s: %s [%s]%s",
			connector, node.Type, node.ID, node.Title, node.Status, mark)
		lines = append(lines, line)
	}

	return lines
}
