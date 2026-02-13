package workflow

import (
	"fmt"

	"github.com/marcus/td/internal/models"
)

// TransitionError represents an error when a transition is not allowed
type TransitionError struct {
	From    models.Status
	To      models.Status
	Reason  string
	IssueID string
}

func (e *TransitionError) Error() string {
	if e.IssueID != "" {
		return fmt.Sprintf("cannot transition %s from %s to %s: %s", e.IssueID, e.From, e.To, e.Reason)
	}
	return fmt.Sprintf("cannot transition from %s to %s: %s", e.From, e.To, e.Reason)
}

// GuardError represents an error when a guard check fails
type GuardError struct {
	GuardName string
	Reason    string
	IssueID   string
}

func (e *GuardError) Error() string {
	if e.IssueID != "" {
		return fmt.Sprintf("guard %s failed for %s: %s", e.GuardName, e.IssueID, e.Reason)
	}
	return fmt.Sprintf("guard %s failed: %s", e.GuardName, e.Reason)
}

// ValidationError wraps multiple guard failures
type ValidationError struct {
	Errors []error
}

func (e *ValidationError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%d validation errors", len(e.Errors))
}

// Add adds an error to the validation error
func (e *ValidationError) Add(err error) {
	e.Errors = append(e.Errors, err)
}

// HasErrors returns true if there are validation errors
func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}
