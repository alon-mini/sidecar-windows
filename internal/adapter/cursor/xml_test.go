package cursor

import "testing"

func TestExtractUserQuery(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "Empty string",
			content: "",
			want:    "",
		},
		{
			name:    "No tags",
			content: "Just plain text",
			want:    "",
		},
		{
			name:    "Simple user query",
			content: "<user_query>Hello world</user_query>",
			want:    "Hello world",
		},
		{
			name:    "Surrounding content",
			content: "Prefix <user_query>Target</user_query> Suffix",
			want:    "Target",
		},
		{
			name:    "Multiple lines",
			content: "<user_query>\nLine 1\nLine 2\n</user_query>",
			want:    "Line 1\nLine 2",
		},
		{
			name:    "Nested tags in query",
			content: "<user_query>Query with <b>bold</b> text</user_query>",
			want:    "Query with <b>bold</b> text",
		},
		{
			name:    "Incomplete tags",
			content: "<user_query>Missing closing tag",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractUserQuery(tt.content)
			if got != tt.want {
				t.Errorf("extractUserQuery(%q) = %q, want %q", tt.content, got, tt.want)
			}
		})
	}
}

func TestIsSystemContextMessage(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "Empty string",
			content: "",
			want:    false,
		},
		{
			name:    "Plain user message",
			content: "Hello",
			want:    false,
		},
		{
			name:    "User query with system tags",
			content: "<user_info>OS: Mac</user_info><user_query>Hi</user_query>",
			want:    false, // Has user query, so not just system context
		},
		{
			name:    "System context only (user_info)",
			content: "<user_info>OS: Mac</user_info>Rules...",
			want:    true,
		},
		{
			name:    "System context only (project_layout)",
			content: "<project_layout>Files...</project_layout>",
			want:    true,
		},
		{
			name:    "System context only (git_status)",
			content: "<git_status>Clean</git_status>",
			want:    true,
		},
		{
			name:    "Mixed system tags without query",
			content: "<user_info>...</user_info><git_status>...</git_status>",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSystemContextMessage(tt.content)
			if got != tt.want {
				t.Errorf("isSystemContextMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStripXMLTags(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "Extracts user query if present",
			content: "<system>...</system><user_query>Hello</user_query>",
			want:    "Hello",
		},
		{
			name:    "Strips all tags if no user query",
			content: "<tag>Content</tag> with <other>tags</other>",
			want:    "Content with tags",
		},
		{
			name:    "Preserves content without tags",
			content: "Just plain text",
			want:    "Just plain text",
		},
		{
			name:    "Handles nested tags (naive stripping)",
			content: "<outer>Outer <inner>Inner</inner></outer>",
			want:    "Outer Inner",
		},
		{
			name:    "Handles attributes in tags",
			content: `<span class="red">Red</span> text`,
			want:    "Red text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripXMLTags(tt.content)
			if got != tt.want {
				t.Errorf("stripXMLTags(%q) = %q, want %q", tt.content, got, tt.want)
			}
		})
	}
}
