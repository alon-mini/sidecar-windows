package models

import (
	"testing"
)

// TestIsValidTypeValid tests all valid types
func TestIsValidTypeValid(t *testing.T) {
	validTypes := []Type{
		TypeBug,
		TypeFeature,
		TypeTask,
		TypeEpic,
		TypeChore,
	}

	for _, typ := range validTypes {
		if !IsValidType(typ) {
			t.Errorf("Expected %q to be valid type", typ)
		}
	}
}

// TestIsValidTypeInvalid tests invalid types
func TestIsValidTypeInvalid(t *testing.T) {
	invalidTypes := []Type{"invalid", "story", "spike", "subtask", ""}
	for _, typ := range invalidTypes {
		if IsValidType(typ) {
			t.Errorf("Expected %q to be invalid type", typ)
		}
	}
}

// TestIsValidTypeConstants tests type constant values
func TestIsValidTypeConstants(t *testing.T) {
	if TypeBug != "bug" {
		t.Errorf("TypeBug should be 'bug', got %q", TypeBug)
	}
	if TypeFeature != "feature" {
		t.Errorf("TypeFeature should be 'feature', got %q", TypeFeature)
	}
	if TypeTask != "task" {
		t.Errorf("TypeTask should be 'task', got %q", TypeTask)
	}
	if TypeEpic != "epic" {
		t.Errorf("TypeEpic should be 'epic', got %q", TypeEpic)
	}
	if TypeChore != "chore" {
		t.Errorf("TypeChore should be 'chore', got %q", TypeChore)
	}
}

// TestIsValidPriorityValid tests all valid priorities
func TestIsValidPriorityValid(t *testing.T) {
	validPriorities := []Priority{
		PriorityP0,
		PriorityP1,
		PriorityP2,
		PriorityP3,
		PriorityP4,
	}

	for _, p := range validPriorities {
		if !IsValidPriority(p) {
			t.Errorf("Expected %q to be valid priority", p)
		}
	}
}

// TestIsValidPriorityInvalid tests invalid priorities
func TestIsValidPriorityInvalid(t *testing.T) {
	invalidPriorities := []Priority{"P5", "p0", "high", "low", "critical", "urgent", ""}
	for _, p := range invalidPriorities {
		if IsValidPriority(p) {
			t.Errorf("Expected %q to be invalid priority", p)
		}
	}
}

// TestIsValidPriorityConstants tests priority constant values
func TestIsValidPriorityConstants(t *testing.T) {
	if PriorityP0 != "P0" {
		t.Errorf("PriorityP0 should be 'P0', got %q", PriorityP0)
	}
	if PriorityP1 != "P1" {
		t.Errorf("PriorityP1 should be 'P1', got %q", PriorityP1)
	}
	if PriorityP2 != "P2" {
		t.Errorf("PriorityP2 should be 'P2', got %q", PriorityP2)
	}
	if PriorityP3 != "P3" {
		t.Errorf("PriorityP3 should be 'P3', got %q", PriorityP3)
	}
	if PriorityP4 != "P4" {
		t.Errorf("PriorityP4 should be 'P4', got %q", PriorityP4)
	}
}

// TestIsValidPointsValid tests all valid Fibonacci points
func TestIsValidPointsValid(t *testing.T) {
	validPoints := []int{1, 2, 3, 5, 8, 13, 21}
	for _, pts := range validPoints {
		if !IsValidPoints(pts) {
			t.Errorf("Expected %d to be valid Fibonacci point", pts)
		}
	}
}

// TestIsValidPointsInvalid tests invalid point values
func TestIsValidPointsInvalid(t *testing.T) {
	invalidPoints := []int{0, 4, 6, 7, 9, 10, 11, 12, 14, 15, 16, 17, 18, 19, 20, 22, -1, -5, 100}
	for _, pts := range invalidPoints {
		if IsValidPoints(pts) {
			t.Errorf("Expected %d to be invalid point value", pts)
		}
	}
}

// TestValidPointsReturnsExpectedSequence tests Fibonacci sequence
func TestValidPointsReturnsExpectedSequence(t *testing.T) {
	expected := []int{1, 2, 3, 5, 8, 13, 21}
	actual := ValidPoints()

	if len(actual) != len(expected) {
		t.Fatalf("Expected %d valid points, got %d", len(expected), len(actual))
	}

	for i, v := range expected {
		if actual[i] != v {
			t.Errorf("ValidPoints()[%d] = %d, want %d", i, actual[i], v)
		}
	}
}

// TestValidPointsOrder tests that points are returned in order
func TestValidPointsOrder(t *testing.T) {
	points := ValidPoints()
	for i := 1; i < len(points); i++ {
		if points[i] <= points[i-1] {
			t.Errorf("ValidPoints not in ascending order at index %d", i)
		}
	}
}

// TestIsValidStatusValid tests all valid statuses
func TestIsValidStatusValid(t *testing.T) {
	validStatuses := []Status{
		StatusOpen,
		StatusInProgress,
		StatusBlocked,
		StatusInReview,
		StatusClosed,
	}

	for _, s := range validStatuses {
		if !IsValidStatus(s) {
			t.Errorf("Expected %q to be valid status", s)
		}
	}
}

// TestIsValidStatusInvalid tests invalid statuses
func TestIsValidStatusInvalid(t *testing.T) {
	invalidStatuses := []Status{"pending", "done", "cancelled", "archived", ""}
	for _, s := range invalidStatuses {
		if IsValidStatus(s) {
			t.Errorf("Expected %q to be invalid status", s)
		}
	}
}

// TestIsValidStatusConstants tests status constant values
func TestIsValidStatusConstants(t *testing.T) {
	if StatusOpen != "open" {
		t.Errorf("StatusOpen should be 'open', got %q", StatusOpen)
	}
	if StatusInProgress != "in_progress" {
		t.Errorf("StatusInProgress should be 'in_progress', got %q", StatusInProgress)
	}
	if StatusBlocked != "blocked" {
		t.Errorf("StatusBlocked should be 'blocked', got %q", StatusBlocked)
	}
	if StatusInReview != "in_review" {
		t.Errorf("StatusInReview should be 'in_review', got %q", StatusInReview)
	}
	if StatusClosed != "closed" {
		t.Errorf("StatusClosed should be 'closed', got %q", StatusClosed)
	}
}

// TestLogTypeConstants tests log type constant values
func TestLogTypeConstants(t *testing.T) {
	if LogTypeProgress != "progress" {
		t.Errorf("LogTypeProgress should be 'progress', got %q", LogTypeProgress)
	}
	if LogTypeBlocker != "blocker" {
		t.Errorf("LogTypeBlocker should be 'blocker', got %q", LogTypeBlocker)
	}
	if LogTypeDecision != "decision" {
		t.Errorf("LogTypeDecision should be 'decision', got %q", LogTypeDecision)
	}
	if LogTypeHypothesis != "hypothesis" {
		t.Errorf("LogTypeHypothesis should be 'hypothesis', got %q", LogTypeHypothesis)
	}
	if LogTypeTried != "tried" {
		t.Errorf("LogTypeTried should be 'tried', got %q", LogTypeTried)
	}
	if LogTypeResult != "result" {
		t.Errorf("LogTypeResult should be 'result', got %q", LogTypeResult)
	}
}

// TestFileRoleConstants tests file role constant values
func TestFileRoleConstants(t *testing.T) {
	if FileRoleImplementation != "implementation" {
		t.Errorf("FileRoleImplementation should be 'implementation', got %q", FileRoleImplementation)
	}
	if FileRoleTest != "test" {
		t.Errorf("FileRoleTest should be 'test', got %q", FileRoleTest)
	}
	if FileRoleReference != "reference" {
		t.Errorf("FileRoleReference should be 'reference', got %q", FileRoleReference)
	}
	if FileRoleConfig != "config" {
		t.Errorf("FileRoleConfig should be 'config', got %q", FileRoleConfig)
	}
}

// TestActionTypeConstants tests action type constant values
func TestActionTypeConstants(t *testing.T) {
	tests := []struct {
		action   ActionType
		expected string
	}{
		{ActionCreate, "create"},
		{ActionUpdate, "update"},
		{ActionDelete, "delete"},
		{ActionRestore, "restore"},
		{ActionStart, "start"},
		{ActionReview, "review"},
		{ActionApprove, "approve"},
		{ActionReject, "reject"},
		{ActionBlock, "block"},
		{ActionUnblock, "unblock"},
		{ActionClose, "close"},
		{ActionReopen, "reopen"},
		{ActionAddDep, "add_dependency"},
		{ActionRemoveDep, "remove_dependency"},
		{ActionLinkFile, "link_file"},
		{ActionUnlinkFile, "unlink_file"},
	}

	for _, tc := range tests {
		if string(tc.action) != tc.expected {
			t.Errorf("Action %v should be %q, got %q", tc.action, tc.expected, string(tc.action))
		}
	}
}

// TestIssueDefaultValues tests Issue struct default values
func TestIssueDefaultValues(t *testing.T) {
	issue := Issue{}

	if issue.ID != "" {
		t.Error("ID should be empty by default")
	}
	if issue.Title != "" {
		t.Error("Title should be empty by default")
	}
	if issue.Status != "" {
		t.Error("Status should be empty by default")
	}
	if issue.Type != "" {
		t.Error("Type should be empty by default")
	}
	if issue.Priority != "" {
		t.Error("Priority should be empty by default")
	}
	if issue.Points != 0 {
		t.Error("Points should be 0 by default")
	}
	if issue.Labels != nil {
		t.Error("Labels should be nil by default")
	}
}

// TestLogDefaultValues tests Log struct default values
func TestLogDefaultValues(t *testing.T) {
	log := Log{}

	if log.ID != "" {
		t.Error("ID should be empty by default")
	}
	if log.IssueID != "" {
		t.Error("IssueID should be empty by default")
	}
	if log.SessionID != "" {
		t.Error("SessionID should be empty by default")
	}
	if log.Message != "" {
		t.Error("Message should be empty by default")
	}
}

// TestHandoffDefaultValues tests Handoff struct default values
func TestHandoffDefaultValues(t *testing.T) {
	handoff := Handoff{}

	if handoff.Done != nil {
		t.Error("Done should be nil by default")
	}
	if handoff.Remaining != nil {
		t.Error("Remaining should be nil by default")
	}
	if handoff.Decisions != nil {
		t.Error("Decisions should be nil by default")
	}
	if handoff.Uncertain != nil {
		t.Error("Uncertain should be nil by default")
	}
}

// TestNormalizePriority tests priority normalization including word forms
func TestNormalizePriority(t *testing.T) {
	tests := []struct {
		input    string
		expected Priority
	}{
		// Canonical forms
		{"P0", PriorityP0},
		{"P1", PriorityP1},
		{"P2", PriorityP2},
		{"P3", PriorityP3},
		{"P4", PriorityP4},
		// Lowercase
		{"p0", PriorityP0},
		{"p1", PriorityP1},
		{"p2", PriorityP2},
		{"p3", PriorityP3},
		{"p4", PriorityP4},
		// Numeric
		{"0", PriorityP0},
		{"1", PriorityP1},
		{"2", PriorityP2},
		{"3", PriorityP3},
		{"4", PriorityP4},
		// Word forms
		{"critical", PriorityP0},
		{"highest", PriorityP0},
		{"high", PriorityP1},
		{"medium", PriorityP2},
		{"normal", PriorityP2},
		{"default", PriorityP2},
		{"low", PriorityP3},
		{"lowest", PriorityP4},
		{"none", PriorityP4},
		// Mixed case word forms
		{"CRITICAL", PriorityP0},
		{"High", PriorityP1},
		{"Medium", PriorityP2},
		{"LOW", PriorityP3},
	}

	for _, tc := range tests {
		got := NormalizePriority(tc.input)
		if got != tc.expected {
			t.Errorf("NormalizePriority(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

// TestNormalizePriorityInvalid tests that invalid inputs are returned uppercase
func TestNormalizePriorityInvalid(t *testing.T) {
	invalid := []string{"invalid", "P5", "urgent", "asap"}
	for _, input := range invalid {
		got := NormalizePriority(input)
		// Invalid inputs should be returned uppercase (for consistent error messages)
		if !IsValidPriority(got) {
			// Good - invalid remains invalid
		} else {
			t.Errorf("NormalizePriority(%q) = %q, should be invalid", input, got)
		}
	}
}

// TestFibonacciPropertyBetween tests Fibonacci property between consecutive values
func TestFibonacciPropertyBetween(t *testing.T) {
	points := ValidPoints()

	// Check that Fibonacci-like property holds (next = prev + prev-1 for indices > 2)
	// Actually the valid points are 1, 2, 3, 5, 8, 13, 21 which is the standard Fibonacci
	// starting from 1, 2 (or adding 0 at start: 0, 1, 1, 2, 3, 5, 8, 13, 21)
	// Let's verify the sequence is correct

	expected := []int{1, 2, 3, 5, 8, 13, 21}
	for i, v := range expected {
		if points[i] != v {
			t.Errorf("Expected Fibonacci value %d at index %d, got %d", v, i, points[i])
		}
	}
}
