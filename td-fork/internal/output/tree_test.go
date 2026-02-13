package output

import (
	"strings"
	"testing"

	"github.com/marcus/td/internal/models"
)

func TestRenderTreeLines_Empty(t *testing.T) {
	lines := RenderTreeLines(nil, TreeRenderOptions{})
	if len(lines) != 0 {
		t.Errorf("expected empty lines, got %d", len(lines))
	}
}

func TestRenderTreeLines_SingleNode(t *testing.T) {
	nodes := []TreeNode{
		{ID: "td-abc", Title: "Test issue", Type: models.TypeTask, Status: models.StatusOpen},
	}
	lines := RenderTreeLines(nodes, TreeRenderOptions{ShowType: true, ShowStatus: true})

	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	line := lines[0]
	if !strings.Contains(line, "└──") {
		t.Errorf("expected last-item connector, got: %s", line)
	}
	if !strings.Contains(line, "td-abc:") {
		t.Errorf("expected ID in output, got: %s", line)
	}
	if !strings.Contains(line, "Test issue") {
		t.Errorf("expected title in output, got: %s", line)
	}
	if !strings.Contains(line, "task") {
		t.Errorf("expected type in output, got: %s", line)
	}
	if !strings.Contains(line, "[open]") {
		t.Errorf("expected status in output, got: %s", line)
	}
}

func TestRenderTreeLines_MultipleNodes(t *testing.T) {
	nodes := []TreeNode{
		{ID: "td-001", Title: "First", Type: models.TypeTask, Status: models.StatusOpen},
		{ID: "td-002", Title: "Second", Type: models.TypeTask, Status: models.StatusClosed},
	}
	lines := RenderTreeLines(nodes, TreeRenderOptions{ShowType: true, ShowStatus: true})

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// First node should have ├──
	if !strings.Contains(lines[0], "├──") {
		t.Errorf("expected non-last connector for first node, got: %s", lines[0])
	}

	// Last node should have └──
	if !strings.Contains(lines[1], "└──") {
		t.Errorf("expected last connector for second node, got: %s", lines[1])
	}

	// Closed status should have checkmark
	if !strings.Contains(lines[1], "✓") {
		t.Errorf("expected checkmark for closed status, got: %s", lines[1])
	}
}

func TestRenderTreeLines_WithChildren(t *testing.T) {
	nodes := []TreeNode{
		{
			ID:     "td-parent",
			Title:  "Parent",
			Type:   models.TypeEpic,
			Status: models.StatusInProgress,
			Children: []TreeNode{
				{ID: "td-child1", Title: "Child 1", Type: models.TypeTask, Status: models.StatusOpen},
				{ID: "td-child2", Title: "Child 2", Type: models.TypeTask, Status: models.StatusClosed},
			},
		},
	}
	lines := RenderTreeLines(nodes, TreeRenderOptions{ShowType: true, ShowStatus: true})

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (parent + 2 children), got %d: %v", len(lines), lines)
	}

	// Children should be indented
	if !strings.HasPrefix(lines[1], "    ") {
		t.Errorf("expected indentation for child, got: %s", lines[1])
	}
}

func TestRenderTreeLines_MaxDepth(t *testing.T) {
	nodes := []TreeNode{
		{
			ID:     "td-level0",
			Title:  "Level 0",
			Type:   models.TypeEpic,
			Status: models.StatusOpen,
			Children: []TreeNode{
				{
					ID:     "td-level1",
					Title:  "Level 1",
					Type:   models.TypeTask,
					Status: models.StatusOpen,
					Children: []TreeNode{
						{ID: "td-level2", Title: "Level 2", Type: models.TypeTask, Status: models.StatusOpen},
					},
				},
			},
		},
	}

	// MaxDepth 1 should only show level 0 (roots) and level 1
	lines := RenderTreeLines(nodes, TreeRenderOptions{MaxDepth: 1})

	if len(lines) != 1 {
		t.Errorf("expected 1 line with MaxDepth=1, got %d: %v", len(lines), lines)
	}
}

func TestRenderTree(t *testing.T) {
	root := TreeNode{
		ID:     "td-root",
		Title:  "Root",
		Type:   models.TypeEpic,
		Status: models.StatusOpen,
		Children: []TreeNode{
			{ID: "td-child", Title: "Child", Type: models.TypeTask, Status: models.StatusOpen},
		},
	}

	result := RenderTree(root, TreeRenderOptions{ShowType: true})

	// Should contain the child but format as a string
	if !strings.Contains(result, "td-child:") {
		t.Errorf("expected child in output, got: %s", result)
	}
}

func TestStatusMark(t *testing.T) {
	tests := []struct {
		status models.Status
		expect string
	}{
		{models.StatusClosed, " ✓"},
		{models.StatusInReview, " ⧗"},
		{models.StatusInProgress, " ●"},
		{models.StatusBlocked, " ✗"},
		{models.StatusOpen, ""},
	}

	for _, tc := range tests {
		result := statusMark(tc.status)
		if result != tc.expect {
			t.Errorf("statusMark(%s) = %q, want %q", tc.status, result, tc.expect)
		}
	}
}

func TestRenderBlockedTree(t *testing.T) {
	nodes := []TreeNode{
		{ID: "td-blocked", Title: "Blocked issue", Status: models.StatusBlocked},
	}

	result := RenderBlockedTree(nodes, TreeRenderOptions{}, false)

	if !strings.Contains(result, "blocks:") {
		t.Errorf("expected 'blocks:' header, got: %s", result)
	}
	if !strings.Contains(result, "td-blocked") {
		t.Errorf("expected blocked issue ID, got: %s", result)
	}
}

func TestRenderBlockedTree_Empty(t *testing.T) {
	result := RenderBlockedTree(nil, TreeRenderOptions{}, false)
	if result != "" {
		t.Errorf("expected empty string for empty nodes, got: %s", result)
	}
}

func TestRenderBlockedTree_SkipsDuplicates(t *testing.T) {
	// Create a scenario where an issue appears multiple times
	nodes := []TreeNode{
		{ID: "td-001", Title: "First", Status: models.StatusOpen},
		{ID: "td-001", Title: "First (duplicate)", Status: models.StatusOpen}, // Same ID
		{ID: "td-002", Title: "Second", Status: models.StatusOpen},
	}

	result := RenderBlockedTree(nodes, TreeRenderOptions{}, false)

	// Count occurrences of td-001
	count := strings.Count(result, "td-001:")
	if count != 1 {
		t.Errorf("expected td-001 to appear once, got %d times in: %s", count, result)
	}
}

func TestRenderChildrenList(t *testing.T) {
	nodes := []TreeNode{
		{ID: "td-001", Title: "First", Type: models.TypeTask, Status: models.StatusOpen},
		{ID: "td-002", Title: "Second", Type: models.TypeBug, Status: models.StatusClosed},
	}

	lines := RenderChildrenList(nodes)

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// Check indentation (2 spaces)
	if !strings.HasPrefix(lines[0], "  ") {
		t.Errorf("expected 2-space indent, got: %s", lines[0])
	}

	// Check format: "  ├── task td-001: First [open]"
	if !strings.Contains(lines[0], "├──") {
		t.Errorf("expected non-last connector, got: %s", lines[0])
	}
	if !strings.Contains(lines[0], "task") {
		t.Errorf("expected type, got: %s", lines[0])
	}
	if !strings.Contains(lines[0], "[open]") {
		t.Errorf("expected status in brackets, got: %s", lines[0])
	}

	// Last item uses └──
	if !strings.Contains(lines[1], "└──") {
		t.Errorf("expected last connector, got: %s", lines[1])
	}

	// Closed status has checkmark
	if !strings.Contains(lines[1], "✓") {
		t.Errorf("expected checkmark for closed, got: %s", lines[1])
	}
}

func TestRenderTreeLines_ShowTypeAndStatus(t *testing.T) {
	nodes := []TreeNode{
		{ID: "td-test", Title: "Test", Type: models.TypeFeature, Status: models.StatusInProgress},
	}

	// Without ShowType
	linesNoType := RenderTreeLines(nodes, TreeRenderOptions{ShowStatus: true})
	if strings.Contains(linesNoType[0], "feature") {
		t.Errorf("should not show type when ShowType=false: %s", linesNoType[0])
	}

	// With ShowType
	linesWithType := RenderTreeLines(nodes, TreeRenderOptions{ShowType: true, ShowStatus: true})
	if !strings.Contains(linesWithType[0], "feature") {
		t.Errorf("should show type when ShowType=true: %s", linesWithType[0])
	}

	// Without ShowStatus
	linesNoStatus := RenderTreeLines(nodes, TreeRenderOptions{ShowType: true})
	if strings.Contains(linesNoStatus[0], "[in_progress]") {
		t.Errorf("should not show status when ShowStatus=false: %s", linesNoStatus[0])
	}
}
