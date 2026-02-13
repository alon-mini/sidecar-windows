---
name: create-modal
description: Create declarative modals using the modal library API. Covers modal types (confirm, input, select, form), sections (Text, Buttons, Input, Textarea, Checkbox, List, When, Custom), rendering with OverlayModal, and keyboard/mouse handling. Use when adding modals or dialogs to the application.
---

# Creating Declarative Modals

Use the `internal/modal` package. The library handles keyboard navigation, mouse hit regions, hover states, and scrolling automatically.

## Quick Start

```go
import "github.com/marcus/sidecar/internal/modal"

// 1. Create the modal
m := modal.New("Delete Worktree?",
    modal.WithWidth(58),
    modal.WithVariant(modal.VariantDanger),
    modal.WithPrimaryAction("delete"),
).
    AddSection(modal.Text("Name: " + wt.Name)).
    AddSection(modal.Spacer()).
    AddSection(modal.Buttons(
        modal.Btn(" Delete ", "delete", modal.BtnDanger()),
        modal.Btn(" Cancel ", "cancel"),
    ))

// 2. Render in View
func (p *Plugin) View(width, height int) string {
    background := p.renderListView(width, height)
    rendered := p.myModal.Render(width, height, p.mouseHandler)
    return ui.OverlayModal(background, rendered, width, height)
}

// 3. Handle input in Update
case tea.KeyMsg:
    action, cmd := p.myModal.HandleKey(msg)
    if action != "" {
        return p.handleAction(action) // "delete", "cancel", etc.
    }
    return p, cmd

case tea.MouseMsg:
    action := p.myModal.HandleMouse(msg, p.mouseHandler)
    if action != "" {
        return p.handleAction(action)
    }
    return p, nil
```

## Critical: Modal Initialization Pattern

The modal must exist before input handling. Create an `ensure` function called in **both** View and Update:

```go
func (p *Plugin) ensureMyModal() {
    if p.targetItem == nil {
        return // Required state missing
    }

    modalW := 50
    if modalW > p.width-4 {
        modalW = p.width - 4
    }
    if modalW < 20 {
        modalW = 20
    }

    // Only rebuild if needed
    if p.myModal != nil && p.myModalWidthCache == modalW {
        return
    }
    p.myModalWidthCache = modalW

    p.myModal = modal.New("Title", modal.WithWidth(modalW), ...).
        AddSection(...)
}
```

**Call `ensureModal()` before the nil check in key handlers:**

```go
func (p *Plugin) handleMyModalKeys(msg tea.KeyMsg) tea.Cmd {
    p.ensureMyModal()  // CRITICAL: Before nil check
    if p.myModal == nil {
        return nil
    }
    action, cmd := p.myModal.HandleKey(msg)
    // ...
}
```

Without this, the first keypress after opening drops because View runs after Update in bubbletea.

## Constructor and Options

```go
m := modal.New(title string, opts ...Option)
```

| Option | Description | Default |
|--------|-------------|---------|
| `WithWidth(int)` | Modal width in characters | 50 |
| `WithVariant(Variant)` | Visual style | `VariantDefault` |
| `WithPrimaryAction(string)` | Action ID for Enter on inputs | "" |
| `WithHints(bool)` | Show "Tab to switch..." hint | true |
| `WithCloseOnBackdropClick(bool)` | Backdrop click returns "cancel" | true |

**Variants:** `VariantDefault`, `VariantDanger` (red), `VariantWarning` (yellow), `VariantInfo` (blue)

## Built-in Sections

### Text and Spacer
```go
modal.Text("Static text with auto line wrapping")
modal.Spacer()  // Single blank line
```

### Buttons
```go
modal.Buttons(
    modal.Btn(" Save ", "save"),              // Standard button
    modal.Btn(" Delete ", "delete", modal.BtnDanger()),  // Red
    modal.Btn(" Submit ", "submit", modal.BtnPrimary()), // Primary
    modal.Btn(" Cancel ", "cancel"),
)
```
- Include padding in labels: `" Save "` not `"Save"`
- Button IDs are returned as actions
- Tab/Shift+Tab cycles focus

### Input
```go
var nameInput textinput.Model
modal.Input("name-input", &nameInput)
modal.InputWithLabel("name-input", "Name:", &nameInput)
modal.Input("name-input", &nameInput,
    modal.WithSubmitOnEnter(true),       // Default: true
    modal.WithSubmitAction("submit"),    // Override primary action
)
```

### Textarea
```go
var msgArea textarea.Model
modal.Textarea("message", &msgArea, 5)          // height in lines
modal.TextareaWithLabel("message", "Label:", &msgArea, 5)
```
- Enter inserts newlines (never submits)

### Checkbox
```go
var includeFiles bool
modal.Checkbox("include-files", "Include untracked files", &includeFiles)
```
- Enter or Space toggles

### List
```go
items := []modal.ListItem{
    {ID: "item-1", Label: "First item", Data: someValue},
    {ID: "item-2", Label: "Second item"},
}
var selectedIdx int
modal.List("my-list", items, &selectedIdx, modal.WithMaxVisible(5))
```
- j/k or up/down moves selection; Enter returns selected item's ID

### When (Conditional)
```go
modal.When(func() bool { return showWarning },
    modal.Text("Warning: This action is irreversible!"),
)
```

### Custom
```go
modal.Custom(
    func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
        return modal.RenderedSection{
            Content: content,
            Focusables: []modal.FocusableInfo{
                {ID: "custom-btn", OffsetX: 0, OffsetY: 2, Width: 10, Height: 1},
            },
        }
    },
    func(msg tea.Msg, focusID string) (string, tea.Cmd) {
        return "", nil  // can be nil if no custom input handling
    },
)
```

## Handling Input

### Keyboard

```go
action, cmd := m.HandleKey(msg)
```

| Key | Behavior |
|-----|----------|
| Tab | Focus next element |
| Shift+Tab | Focus previous element |
| Enter | Return focused element's ID (or primaryAction for inputs) |
| Esc | Return "cancel" |
| Other | Forwarded to focused section |

### Mouse

```go
action := m.HandleMouse(msg, p.mouseHandler)
```

| Event | Behavior |
|-------|----------|
| Click backdrop | Return "cancel" (if enabled) |
| Click button/checkbox | Return element ID |
| Hover element | Update hover state |
| Scroll on modal | Scroll content |

## Modal Methods

```go
m.FocusedID() string   // Currently focused element ID
m.HoveredID() string   // Currently hovered element ID
m.SetFocus(id string)  // Focus specific element
m.Reset()              // Reset focus, hover, scroll to initial state
```

## Rendering Rules

Always use `ui.OverlayModal` for dimmed background:
```go
func (p *Plugin) View(width, height int) string {
    background := p.renderNormalView(width, height)
    rendered := p.myModal.Render(width, height, p.mouseHandler)
    return ui.OverlayModal(background, rendered, width, height)
}
```

**Do not:**
- Pre-center modal content with `lipgloss.Place` (OverlayModal handles centering)
- Render footers or hint lines in plugin View (app renders unified footer)

## State Management

- Focus state persists across renders
- Call `Reset()` when closing and reopening modals
- Width caching should include state-dependent changes

## Troubleshooting

| Issue | Solution |
|-------|----------|
| First keypress dropped | Call `ensureModal()` before nil check in Update |
| Modal too wide/narrow | Use width clamping: `modalW > p.width-4` |
| Hover not updating | Pass `mouseHandler` to both `Render` and `HandleMouse` |
| Input not receiving keys | Check `FocusedID()` |
| Modal rebuilds every frame | Cache by width |
| Modal shows with wrong focus | Call `m.Reset()` when showing modal |

See `references/complete-example.md` for a full plugin implementation with delete confirmation modal.
