// Package modal provides a declarative modal dialog library with automatic
// hit region management for mouse support.
//
// The library eliminates off-by-one hit region bugs via a render-then-measure
// pattern and provides automatic keyboard navigation (Tab/Shift+Tab, Enter, Esc)
// and hover state management.
//
// # Quick Start
//
//	m := modal.New("Confirm Delete", modal.WithVariant(modal.VariantDanger)).
//	    AddSection(modal.Text("Are you sure you want to delete this item?")).
//	    AddSection(modal.Spacer()).
//	    AddSection(modal.Buttons(
//	        modal.Btn(" Delete ", "delete", modal.BtnDanger()),
//	        modal.Btn(" Cancel ", "cancel"),
//	    ))
//
//	// In View():
//	content := m.Render(screenW, screenH, mouseHandler)
//
//	// In Update():
//	if action, cmd := m.HandleKey(keyMsg); action != "" {
//	    switch action {
//	    case "delete":
//	        return performDelete()
//	    case "cancel":
//	        return closeModal()
//	    }
//	}
//
// # Built-in Sections
//
//   - Text(s string) - static text, auto-wrapped
//   - Spacer() - blank line
//   - Buttons(btns ...ButtonDef) - button row with focus/hover styling
//   - Checkbox(id, label string, checked *bool) - toggleable checkbox
//   - Input(id string, model *textinput.Model, opts...) - text input
//   - Textarea(id string, model *textarea.Model, height int, opts...) - multiline
//   - List(id string, items []ListItem, selectedIdx *int, opts...) - scrollable list
//   - When(condition func() bool, section) - conditional rendering
//   - Custom(renderFn, updateFn) - escape hatch for complex content
//
// # Options
//
//   - WithWidth(w int) - set modal width (default: 50)
//   - WithVariant(v Variant) - set visual style (Default, Danger, Warning, Info)
//   - WithHints(show bool) - show/hide keyboard hints at bottom
//   - WithPrimaryAction(actionID string) - action for implicit Enter submit
//   - WithCloseOnBackdropClick(close bool) - close on backdrop click
//
// See the package-level documentation for detailed integration guides.
package modal
