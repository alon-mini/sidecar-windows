package query

import (
	"fmt"
	"strings"

	"github.com/marcus/td/internal/db"
	"github.com/marcus/td/internal/models"
)

// NoteExecuteOptions contains options for note query execution
type NoteExecuteOptions struct {
	Limit      int
	SortBy     string
	SortDesc   bool
	MaxResults int
}

// HasNoteFields checks if the query AST contains any note. prefix fields.
// Used to determine whether to route the query to ExecuteNotes instead of Execute.
func HasNoteFields(q *Query) bool {
	if q.Root == nil {
		return false
	}
	return hasNoteFieldsNode(q.Root)
}

func hasNoteFieldsNode(n Node) bool {
	switch node := n.(type) {
	case *BinaryExpr:
		return hasNoteFieldsNode(node.Left) || hasNoteFieldsNode(node.Right)
	case *UnaryExpr:
		return hasNoteFieldsNode(node.Expr)
	case *FieldExpr:
		parts := strings.Split(node.Field, ".")
		return len(parts) > 1 && parts[0] == "note"
	default:
		return false
	}
}

// ExecuteNotes parses and executes a TDQ query against the notes table.
// The query must use note. prefixed fields (e.g., note.title ~ "meeting").
// The note. prefix is stripped before matching against Note struct fields.
func ExecuteNotes(database NoteQuerySource, queryStr string, opts NoteExecuteOptions) ([]models.Note, error) {
	q, err := Parse(queryStr)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	if errs := q.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("validation error: %v", errs[0])
	}

	maxResults := opts.MaxResults
	if maxResults <= 0 {
		maxResults = DefaultMaxResults
	}

	// Determine sort
	sortBy := opts.SortBy
	sortDesc := opts.SortDesc
	if q.Sort != nil {
		sortBy = q.Sort.Field
		sortDesc = q.Sort.Descending
	}
	// Map sort field to note column if needed
	if col, ok := NoteSortFieldToColumn[sortBy]; ok {
		sortBy = col
	}

	// Fetch notes from DB
	fetchOpts := db.ListNotesOptions{
		Limit:          maxResults,
		IncludeDeleted: false,
	}
	notes, err := database.ListNotes(fetchOpts)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Build matcher and filter in memory
	ctx := NewEvalContext("")
	var filtered []models.Note
	if q.Root == nil {
		filtered = notes
	} else {
		matcher, err := noteNodeToMatcher(q.Root, ctx)
		if err != nil {
			return nil, fmt.Errorf("matcher error: %w", err)
		}
		for _, note := range notes {
			if matcher(note) {
				filtered = append(filtered, note)
			}
		}
	}

	// Sort if requested (ListNotes default order is pinned DESC, updated_at DESC)
	if sortBy != "" {
		sortNotes(filtered, sortBy, sortDesc)
	}

	// Apply limit
	if opts.Limit > 0 && len(filtered) > opts.Limit {
		filtered = filtered[:opts.Limit]
	}

	return filtered, nil
}

// noteNodeToMatcher converts a query AST node to a Note matcher function.
func noteNodeToMatcher(n Node, ctx *EvalContext) (func(models.Note) bool, error) {
	switch node := n.(type) {
	case *BinaryExpr:
		leftMatcher, err := noteNodeToMatcher(node.Left, ctx)
		if err != nil {
			return nil, err
		}
		rightMatcher, err := noteNodeToMatcher(node.Right, ctx)
		if err != nil {
			return nil, err
		}
		if node.Op == OpAnd {
			return func(n models.Note) bool {
				return leftMatcher(n) && rightMatcher(n)
			}, nil
		}
		return func(n models.Note) bool {
			return leftMatcher(n) || rightMatcher(n)
		}, nil

	case *UnaryExpr:
		matcher, err := noteNodeToMatcher(node.Expr, ctx)
		if err != nil {
			return nil, err
		}
		return func(n models.Note) bool {
			return !matcher(n)
		}, nil

	case *FieldExpr:
		return noteFieldExprToMatcher(node, ctx)

	case *TextSearch:
		pattern := strings.ToLower(node.Text)
		return func(n models.Note) bool {
			return strings.Contains(strings.ToLower(n.ID), pattern) ||
				strings.Contains(strings.ToLower(n.Title), pattern) ||
				strings.Contains(strings.ToLower(n.Content), pattern)
		}, nil

	default:
		return nil, fmt.Errorf("unsupported node type for note matcher: %T", n)
	}
}

// noteFieldExprToMatcher converts a field expression to a Note matcher.
// Fields can be prefixed with "note." which is stripped before matching.
func noteFieldExprToMatcher(node *FieldExpr, ctx *EvalContext) (func(models.Note) bool, error) {
	field := node.Field

	// Strip note. prefix
	field, _ = strings.CutPrefix(field, "note.")

	evaluator := &Evaluator{ctx: ctx}
	value := evaluator.resolveValue(node.Value)

	getter := getNoteFieldGetter(field)
	if getter == nil {
		return nil, fmt.Errorf("unknown note field: %s", field)
	}

	switch node.Operator {
	case OpEq:
		return func(n models.Note) bool {
			return evaluator.compareEqual(getter(n), value)
		}, nil
	case OpNeq:
		return func(n models.Note) bool {
			return !evaluator.compareEqual(getter(n), value)
		}, nil
	case OpContains:
		pattern := strings.ToLower(fmt.Sprintf("%v", value))
		return func(n models.Note) bool {
			fieldVal := strings.ToLower(fmt.Sprintf("%v", getter(n)))
			return strings.Contains(fieldVal, pattern)
		}, nil
	case OpNotContains:
		pattern := strings.ToLower(fmt.Sprintf("%v", value))
		return func(n models.Note) bool {
			fieldVal := strings.ToLower(fmt.Sprintf("%v", getter(n)))
			return !strings.Contains(fieldVal, pattern)
		}, nil
	case OpLt, OpGt, OpLte, OpGte:
		return func(n models.Note) bool {
			return evaluator.compareOrder(getter(n), value, node.Operator)
		}, nil
	default:
		return func(models.Note) bool { return true }, nil
	}
}

// getNoteFieldGetter returns a function that extracts a field value from a Note.
func getNoteFieldGetter(field string) func(models.Note) any {
	switch field {
	case "id":
		return func(n models.Note) any { return n.ID }
	case "title":
		return func(n models.Note) any { return n.Title }
	case "content":
		return func(n models.Note) any { return n.Content }
	case "created", "created_at":
		return func(n models.Note) any { return n.CreatedAt }
	case "updated", "updated_at":
		return func(n models.Note) any { return n.UpdatedAt }
	case "pinned":
		return func(n models.Note) any { return n.Pinned }
	case "archived":
		return func(n models.Note) any { return n.Archived }
	default:
		return nil
	}
}

// sortNotes sorts notes in place by the given field.
func sortNotes(notes []models.Note, field string, desc bool) {
	// Use a simple insertion sort (adequate for typical note counts)
	for i := 1; i < len(notes); i++ {
		j := i
		for j > 0 && noteShouldSwap(notes[j-1], notes[j], field, desc) {
			notes[j-1], notes[j] = notes[j], notes[j-1]
			j--
		}
	}
}

// noteShouldSwap returns true if a should come after b in the desired order.
func noteShouldSwap(a, b models.Note, field string, desc bool) bool {
	cmp := noteFieldCompare(a, b, field)
	if desc {
		return cmp < 0 // for descending, swap when a < b
	}
	return cmp > 0 // for ascending, swap when a > b
}

// noteFieldCompare returns -1, 0, or 1 comparing a vs b on the given field.
func noteFieldCompare(a, b models.Note, field string) int {
	switch field {
	case "created_at":
		if a.CreatedAt.Before(b.CreatedAt) {
			return -1
		}
		if a.CreatedAt.After(b.CreatedAt) {
			return 1
		}
		return 0
	case "updated_at":
		if a.UpdatedAt.Before(b.UpdatedAt) {
			return -1
		}
		if a.UpdatedAt.After(b.UpdatedAt) {
			return 1
		}
		return 0
	case "title":
		aLower := strings.ToLower(a.Title)
		bLower := strings.ToLower(b.Title)
		if aLower < bLower {
			return -1
		}
		if aLower > bLower {
			return 1
		}
		return 0
	case "pinned":
		if a.Pinned == b.Pinned {
			return 0
		}
		if a.Pinned {
			return 1 // pinned=true is "greater"
		}
		return -1
	case "archived":
		if a.Archived == b.Archived {
			return 0
		}
		if a.Archived {
			return 1
		}
		return -1
	default:
		return 0
	}
}
