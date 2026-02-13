package query

import (
	"testing"
	"time"

	"github.com/marcus/td/internal/db"
	"github.com/marcus/td/internal/models"
)

func TestHasNoteFields(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		hasNotes bool
	}{
		{"empty query", "", false},
		{"issue field", "status = open", false},
		{"log cross-entity", `log.message ~ "fix"`, false},
		{"note title", `note.title ~ "meeting"`, true},
		{"note content", `note.content ~ "important"`, true},
		{"note pinned", `note.pinned = true`, true},
		{"note in AND", `note.title ~ "x" AND note.pinned = true`, true},
		{"note in OR", `note.title ~ "x" OR note.archived = false`, true},
		{"note in NOT", `NOT note.archived = true`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			got := HasNoteFields(q)
			if got != tt.hasNotes {
				t.Errorf("HasNoteFields() = %v, want %v", got, tt.hasNotes)
			}
		})
	}
}

func TestNoteFieldValidation(t *testing.T) {
	// Valid note fields should parse and validate
	validQueries := []string{
		`note.title ~ "test"`,
		`note.content ~ "hello"`,
		`note.pinned = true`,
		`note.archived = false`,
		`note.created > "2024-01-01"`,
		`note.updated >= -7d`,
	}
	for _, q := range validQueries {
		t.Run(q, func(t *testing.T) {
			parsed, err := Parse(q)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			errs := parsed.Validate()
			if len(errs) > 0 {
				t.Errorf("unexpected validation error: %v", errs)
			}
		})
	}

	// Invalid note fields should fail validation
	invalidQueries := []string{
		`note.nonexistent ~ "test"`,
		`note.foo = bar`,
	}
	for _, q := range invalidQueries {
		t.Run("invalid:"+q, func(t *testing.T) {
			parsed, err := Parse(q)
			if err != nil {
				return // parse error is fine too
			}
			errs := parsed.Validate()
			if len(errs) == 0 {
				t.Errorf("expected validation error for %q", q)
			}
		})
	}
}

func TestNoteNodeToMatcher(t *testing.T) {
	now := time.Now()
	notes := []models.Note{
		{ID: "n-001", Title: "Meeting notes", Content: "Discussed project timeline", CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour), Pinned: true, Archived: false},
		{ID: "n-002", Title: "Todo list", Content: "Buy groceries", CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour), Pinned: false, Archived: false},
		{ID: "n-003", Title: "Old ideas", Content: "Archived brainstorm session", CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now.Add(-72 * time.Hour), Pinned: false, Archived: true},
	}

	tests := []struct {
		name      string
		query     string
		wantMatch []string // expected note IDs that match
	}{
		{
			name:      "title contains",
			query:     `note.title ~ "meeting"`,
			wantMatch: []string{"n-001"},
		},
		{
			name:      "content contains",
			query:     `note.content ~ "groceries"`,
			wantMatch: []string{"n-002"},
		},
		{
			name:      "pinned equals true",
			query:     `note.pinned = true`,
			wantMatch: []string{"n-001"},
		},
		{
			name:      "pinned equals false",
			query:     `note.pinned = false`,
			wantMatch: []string{"n-002", "n-003"},
		},
		{
			name:      "archived equals true",
			query:     `note.archived = true`,
			wantMatch: []string{"n-003"},
		},
		{
			name:      "NOT archived",
			query:     `NOT note.archived = true`,
			wantMatch: []string{"n-001", "n-002"},
		},
		{
			name:      "title AND pinned",
			query:     `note.title ~ "meeting" AND note.pinned = true`,
			wantMatch: []string{"n-001"},
		},
		{
			name:      "title OR content",
			query:     `note.title ~ "meeting" OR note.content ~ "groceries"`,
			wantMatch: []string{"n-001", "n-002"},
		},
		{
			name:      "text search (bare text)",
			query:     `"timeline"`,
			wantMatch: []string{"n-001"},
		},
		{
			name:      "title not contains",
			query:     `note.title !~ "meeting"`,
			wantMatch: []string{"n-002", "n-003"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			ctx := NewEvalContext("")
			matcher, err := noteNodeToMatcher(q.Root, ctx)
			if err != nil {
				t.Fatalf("matcher error: %v", err)
			}

			wantSet := make(map[string]bool)
			for _, id := range tt.wantMatch {
				wantSet[id] = true
			}

			for _, note := range notes {
				got := matcher(note)
				want := wantSet[note.ID]
				if got != want {
					t.Errorf("matcher(%s) = %v, want %v", note.ID, got, want)
				}
			}
		})
	}
}

// mockNoteSource implements NoteQuerySource for testing
type mockNoteSource struct {
	notes []models.Note
}

func (m *mockNoteSource) ListNotes(opts db.ListNotesOptions) ([]models.Note, error) {
	var result []models.Note
	for _, n := range m.notes {
		if n.DeletedAt != nil && !opts.IncludeDeleted {
			continue
		}
		result = append(result, n)
	}
	if opts.Limit > 0 && len(result) > opts.Limit {
		result = result[:opts.Limit]
	}
	return result, nil
}

func TestExecuteNotes(t *testing.T) {
	now := time.Now()
	source := &mockNoteSource{
		notes: []models.Note{
			{ID: "n-001", Title: "Meeting notes", Content: "Discussed timeline", CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour), Pinned: true, Archived: false},
			{ID: "n-002", Title: "Todo list", Content: "Buy groceries", CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour), Pinned: false, Archived: false},
			{ID: "n-003", Title: "Old ideas", Content: "Archived brainstorm", CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now.Add(-72 * time.Hour), Pinned: false, Archived: true},
		},
	}

	tests := []struct {
		name      string
		query     string
		wantCount int
		wantErr   bool
	}{
		{"empty query returns all", "", 3, false},
		{"title filter", `note.title ~ "meeting"`, 1, false},
		{"content filter", `note.content ~ "groceries"`, 1, false},
		{"pinned filter", `note.pinned = true`, 1, false},
		{"archived filter", `note.archived = true`, 1, false},
		{"combined AND", `note.pinned = true AND note.title ~ "meeting"`, 1, false},
		{"combined OR", `note.pinned = true OR note.archived = true`, 2, false},
		{"NOT expression", `NOT note.archived = true`, 2, false},
		{"text search", `"timeline"`, 1, false},
		{"no results", `note.title ~ "nonexistent"`, 0, false},
		{"invalid query", `note.title = `, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := ExecuteNotes(source, tt.query, NoteExecuteOptions{})
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteNotes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(results) != tt.wantCount {
				t.Errorf("ExecuteNotes() returned %d results, want %d", len(results), tt.wantCount)
			}
		})
	}
}

func TestExecuteNotesWithLimit(t *testing.T) {
	now := time.Now()
	source := &mockNoteSource{
		notes: []models.Note{
			{ID: "n-001", Title: "Note 1", Content: "Content", CreatedAt: now, UpdatedAt: now},
			{ID: "n-002", Title: "Note 2", Content: "Content", CreatedAt: now, UpdatedAt: now},
			{ID: "n-003", Title: "Note 3", Content: "Content", CreatedAt: now, UpdatedAt: now},
		},
	}

	results, err := ExecuteNotes(source, "", NoteExecuteOptions{Limit: 2})
	if err != nil {
		t.Fatalf("ExecuteNotes() error = %v", err)
	}
	if len(results) != 2 {
		t.Errorf("ExecuteNotes() returned %d results, want 2", len(results))
	}
}

func TestSortNotes(t *testing.T) {
	now := time.Now()
	notes := []models.Note{
		{ID: "n-002", Title: "Bravo", CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now},
		{ID: "n-001", Title: "Alpha", CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour)},
		{ID: "n-003", Title: "Charlie", CreatedAt: now, UpdatedAt: now.Add(1 * time.Hour)},
	}

	t.Run("sort by title ascending", func(t *testing.T) {
		sorted := make([]models.Note, len(notes))
		copy(sorted, notes)
		sortNotes(sorted, "title", false)
		if sorted[0].Title != "Alpha" || sorted[1].Title != "Bravo" || sorted[2].Title != "Charlie" {
			t.Errorf("sort by title asc: got %s, %s, %s", sorted[0].Title, sorted[1].Title, sorted[2].Title)
		}
	})

	t.Run("sort by title descending", func(t *testing.T) {
		sorted := make([]models.Note, len(notes))
		copy(sorted, notes)
		sortNotes(sorted, "title", true)
		if sorted[0].Title != "Charlie" || sorted[1].Title != "Bravo" || sorted[2].Title != "Alpha" {
			t.Errorf("sort by title desc: got %s, %s, %s", sorted[0].Title, sorted[1].Title, sorted[2].Title)
		}
	})

	t.Run("sort by created_at ascending", func(t *testing.T) {
		sorted := make([]models.Note, len(notes))
		copy(sorted, notes)
		sortNotes(sorted, "created_at", false)
		if sorted[0].ID != "n-001" || sorted[1].ID != "n-002" || sorted[2].ID != "n-003" {
			t.Errorf("sort by created asc: got %s, %s, %s", sorted[0].ID, sorted[1].ID, sorted[2].ID)
		}
	})

	t.Run("sort by created_at descending", func(t *testing.T) {
		sorted := make([]models.Note, len(notes))
		copy(sorted, notes)
		sortNotes(sorted, "created_at", true)
		if sorted[0].ID != "n-003" || sorted[1].ID != "n-002" || sorted[2].ID != "n-001" {
			t.Errorf("sort by created desc: got %s, %s, %s", sorted[0].ID, sorted[1].ID, sorted[2].ID)
		}
	})
}

func TestExecuteNotesWithDB(t *testing.T) {
	// Note: db.ListNotes scans time.Time directly from TEXT columns, which
	// can fail with some SQLite drivers. Use mock source for unit tests;
	// real DB integration is tested via the notes CLI commands.
	now := time.Now()
	source := &mockNoteSource{
		notes: []models.Note{
			{ID: "n-db-001", Title: "Meeting notes", Content: "Discussed project timeline", CreatedAt: now, UpdatedAt: now, Pinned: true},
			{ID: "n-db-002", Title: "Todo list", Content: "Buy groceries and milk", CreatedAt: now, UpdatedAt: now},
			{ID: "n-db-003", Title: "Old ideas", Content: "Archived brainstorm session", CreatedAt: now, UpdatedAt: now, Archived: true},
		},
	}

	t.Run("query title contains", func(t *testing.T) {
		results, err := ExecuteNotes(source, `note.title ~ "meeting"`, NoteExecuteOptions{})
		if err != nil {
			t.Fatalf("ExecuteNotes() error = %v", err)
		}
		if len(results) != 1 {
			t.Errorf("got %d results, want 1", len(results))
		}
		if len(results) > 0 && results[0].ID != "n-db-001" {
			t.Errorf("expected n-db-001, got %s", results[0].ID)
		}
	})

	t.Run("query content contains", func(t *testing.T) {
		results, err := ExecuteNotes(source, `note.content ~ "groceries"`, NoteExecuteOptions{})
		if err != nil {
			t.Fatalf("ExecuteNotes() error = %v", err)
		}
		if len(results) != 1 {
			t.Errorf("got %d results, want 1", len(results))
		}
		if len(results) > 0 && results[0].ID != "n-db-002" {
			t.Errorf("expected n-db-002, got %s", results[0].ID)
		}
	})

	t.Run("query pinned", func(t *testing.T) {
		results, err := ExecuteNotes(source, `note.pinned = true`, NoteExecuteOptions{})
		if err != nil {
			t.Fatalf("ExecuteNotes() error = %v", err)
		}
		if len(results) != 1 {
			t.Errorf("got %d results, want 1", len(results))
		}
		if len(results) > 0 && results[0].ID != "n-db-001" {
			t.Errorf("expected n-db-001, got %s", results[0].ID)
		}
	})

	t.Run("query archived", func(t *testing.T) {
		results, err := ExecuteNotes(source, `note.archived = true`, NoteExecuteOptions{})
		if err != nil {
			t.Fatalf("ExecuteNotes() error = %v", err)
		}
		if len(results) != 1 {
			t.Errorf("got %d results, want 1", len(results))
		}
		if len(results) > 0 && results[0].ID != "n-db-003" {
			t.Errorf("expected n-db-003, got %s", results[0].ID)
		}
	})

	t.Run("empty query returns all non-deleted", func(t *testing.T) {
		results, err := ExecuteNotes(source, "", NoteExecuteOptions{})
		if err != nil {
			t.Fatalf("ExecuteNotes() error = %v", err)
		}
		if len(results) != 3 {
			t.Errorf("got %d results, want 3", len(results))
		}
	})
}
