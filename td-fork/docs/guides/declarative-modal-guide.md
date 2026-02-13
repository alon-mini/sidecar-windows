# Declarative Modal Library Guide

This guide documents the declarative modal library in `pkg/monitor/modal/` and `pkg/monitor/mouse/`. This library provides render-then-measure hit regions that eliminate off-by-one bugs, automatic keyboard/mouse navigation, and a clean declarative API.

## Overview

### Purpose

The declarative modal library solves a common problem in TUI development: accurately tracking hit regions for mouse support. Traditional approaches calculate positions inline while rendering, which leads to:

- Off-by-one errors when positions drift
- Duplicated position calculations (render vs. click handler)
- Fragile code when layouts change

This library uses a **render-then-measure** pattern:

1. Sections render content and report focusable elements with relative offsets
2. Layout phase measures rendered heights
3. Hit regions are registered at absolute positions based on measurements

### Key Benefits

- **No off-by-one bugs**: Hit regions are calculated from actual rendered content
- **Automatic keyboard navigation**: Tab/Shift+Tab cycles focus, Enter activates, Esc cancels
- **Automatic hover management**: Mouse motion updates hover state automatically
- **Declarative API**: Build modals from composable sections
- **Type-safe actions**: Section Update() returns action strings for clean handling

## Quick Start

### Simple Confirmation Modal

```go
import (
    "github.com/marcus/td/pkg/monitor/modal"
    "github.com/marcus/td/pkg/monitor/mouse"
)

// Create modal
m := modal.New("Confirm Delete", modal.WithVariant(modal.VariantDanger)).
    AddSection(modal.Text("Are you sure you want to delete this item?")).
    AddSection(modal.Spacer()).
    AddSection(modal.Buttons(
        modal.Btn(" Delete ", "delete", modal.BtnDanger()),
        modal.Btn(" Cancel ", "cancel"),
    ))

// Create mouse handler
handler := mouse.NewHandler()

// In View():
content := m.Render(screenW, screenH, handler)

// In Update() for keyboard:
if keyMsg, ok := msg.(tea.KeyMsg); ok {
    action, cmd := m.HandleKey(keyMsg)
    if action != "" {
        switch action {
        case "delete":
            return performDelete()
        case "cancel":
            return closeModal()
        }
    }
    return m, cmd
}

// In Update() for mouse:
if mouseMsg, ok := msg.(tea.MouseMsg); ok {
    action := m.HandleMouse(mouseMsg, handler)
    if action != "" {
        switch action {
        case "delete":
            return performDelete()
        case "cancel":
            return closeModal()
        }
    }
}
```

### Form Modal with Input

```go
ti := textinput.New()
ti.Placeholder = "Enter title..."
ti.Focus()

m := modal.New("Create Issue", modal.WithPrimaryAction("submit")).
    AddSection(modal.InputWithLabel("title", "Title:", &ti)).
    AddSection(modal.Spacer()).
    AddSection(modal.Buttons(
        modal.Btn(" Create ", "submit"),
        modal.Btn(" Cancel ", "cancel"),
    ))

// Handle actions
switch action {
case "submit":
    title := ti.Value()
    return createIssue(title)
case "cancel":
    return closeModal()
}
```

### Pointer Pattern for Model-Stored Inputs

The quick-start example above works when the input is a local variable. However, when inputs are stored as fields on a bubbletea `Model` struct, you must use **pointers** (`*textinput.Model` / `*textarea.Model`) to avoid stale data.

**Why:** Bubbletea copies the Model (value receiver) on every `Update` call. If you store `textinput.Model` as a value field and pass `&m.SomeInput` to the modal, the modal captures a pointer to *that specific copy's* field. On the next Update, bubbletea creates a new copy with its own independent field — the modal's pointer now references stale data.

**Solution:** Store pointers on the Model struct. When bubbletea copies the Model, it copies the pointer (not the pointed-to data), so the modal and all Model copies share the same underlying instance.

```go
type Model struct {
    nameInput  *textinput.Model   // pointer, not value
    queryInput *textarea.Model    // pointer, not value
    myModal    *modal.Modal
}

func (m Model) openModal() Model {
    // Allocate on heap, store pointer
    ti := textinput.New()
    ti.Placeholder = "Name..."
    ti.Focus()
    m.nameInput = &ti

    ta := textarea.New()
    ta.SetWidth(40)
    ta.SetHeight(3)
    m.queryInput = &ta

    // Pass pointer directly (no & needed — already a pointer)
    m.myModal = modal.New("Edit",
        modal.WithPrimaryAction("save"),
    )
    m.myModal.AddSection(modal.InputWithLabel("name", "Name:", m.nameInput))
    m.myModal.AddSection(modal.TextareaWithLabel("query", "Query:", m.queryInput, 3))
    m.myModal.Reset()
    return m
}

// In Update: just delegate to modal.HandleKey() — no manual forwarding needed.
// The modal updates the shared input through the pointer automatically.
action, cmd := m.myModal.HandleKey(msg)
```

**Key points:**
- All key routing goes through `modal.HandleKey()` — no manual `Update()` forwarding needed
- No manual `Focus()`/`Blur()` sync needed — the modal manages focus on the shared instances
- Set `textarea.SetWidth()` before any `Update`/`View` calls to avoid zero-width panics
- Set pointers to `nil` when closing the modal to avoid dangling references

## API Reference

### Modal

```go
// Create a new modal
func New(title string, opts ...Option) *Modal

// Add a section (chainable)
func (m *Modal) AddSection(s Section) *Modal

// Render the modal and register hit regions
func (m *Modal) Render(screenW, screenH int, handler *mouse.Handler) string

// Handle keyboard input - returns (action, cmd)
func (m *Modal) HandleKey(msg tea.KeyMsg) (string, tea.Cmd)

// Handle mouse input - returns action
func (m *Modal) HandleMouse(msg tea.MouseMsg, handler *mouse.Handler) string

// Set focus to specific element by ID
func (m *Modal) SetFocus(id string)

// Get currently focused element ID
func (m *Modal) FocusedID() string

// Get currently hovered element ID
func (m *Modal) HoveredID() string

// Reset modal state (focus, hover, scroll)
func (m *Modal) Reset()
```

### Options

```go
// Set modal width (default: 50, min: 30, max: 120)
WithWidth(w int)

// Set visual variant (affects border color)
WithVariant(v Variant)
// Variants: VariantDefault (purple), VariantDanger (red),
//           VariantWarning (yellow), VariantInfo (cyan)

// Show/hide keyboard hints at bottom (default: true)
WithHints(show bool)

// Set action ID for implicit Enter submission (e.g., from input field)
WithPrimaryAction(actionID string)

// Control backdrop click behavior (default: true = closes modal)
WithCloseOnBackdropClick(close bool)
```

### Built-in Sections

#### Text

Static text content with automatic word wrapping.

```go
modal.Text("This is the modal content.\n\nIt can have multiple lines.")
```

#### Spacer

Renders a blank line for visual separation.

```go
modal.Spacer()
```

#### Buttons

Row of interactive buttons with focus/hover styling.

```go
modal.Buttons(
    modal.Btn(" Confirm ", "confirm"),           // Normal button
    modal.Btn(" Delete ", "delete", modal.BtnDanger()), // Danger button (red when focused)
    modal.Btn(" Cancel ", "cancel"),
)
```

Button options:
- `BtnDanger()` - Use danger styling (red background when focused)
- `BtnPrimary()` - No-op for compatibility

#### Checkbox

Toggleable checkbox bound to a bool pointer.

```go
agree := false
modal.Checkbox("agree", "I agree to the terms", &agree)
```

Space or Enter toggles the checkbox.

#### Input

Text input wrapping bubbles `textinput.Model`.

```go
ti := textinput.New()
ti.Placeholder = "Enter name..."

// Without label
modal.Input("name", &ti)

// With label
modal.InputWithLabel("name", "Name:", &ti)
```

Options:
- `WithSubmitOnEnter(bool)` - Enable/disable Enter to submit (default: true)
- `WithSubmitAction(actionID string)` - Action ID to return on Enter

#### Textarea

Multiline text input wrapping bubbles `textarea.Model`.

```go
ta := textarea.New()
ta.SetHeight(5)

modal.Textarea("desc", &ta, 5)  // id, model, height in lines
modal.TextareaWithLabel("desc", "Description:", &ta, 5)
```

#### List

Scrollable list of selectable items.

```go
selectedIdx := 0
items := []modal.ListItem{
    {ID: "opt1", Label: "Option 1", Data: "custom-data-1"},
    {ID: "opt2", Label: "Option 2", Data: "custom-data-2"},
    {ID: "opt3", Label: "Option 3", Data: "custom-data-3"},
}

modal.List("options", items, &selectedIdx, modal.WithMaxVisible(5))
```

Options:
- `WithMaxVisible(n int)` - Maximum visible items (default: 5)

Navigation: Up/Down/j/k to move, Enter to select, Home/End to jump.

#### When

Conditional section that renders only when condition is true.

```go
showAdvanced := false

modal.When(func() bool { return showAdvanced },
    modal.Text("Advanced options would go here"))
```

When the condition is false, the section renders to zero height and contributes no focusables.

#### Custom

Escape hatch for complex custom content.

```go
modal.Custom(
    // Render function
    func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
        // Render your content
        return modal.RenderedSection{
            Content: "Custom content here",
            Focusables: []modal.FocusableInfo{
                {ID: "custom-btn", OffsetX: 0, OffsetY: 0, Width: 10, Height: 1},
            },
        }
    },
    // Update function (can be nil)
    func(msg tea.Msg, focusID string) (string, tea.Cmd) {
        if focusID == "custom-btn" {
            if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
                return "custom-action", nil
            }
        }
        return "", nil
    },
)
```

## Mouse Package

The `pkg/monitor/mouse/` package provides hit region management and mouse state tracking.

### HitMap

Tracks rectangular hit regions with priority (later = higher).

```go
hm := mouse.NewHitMap()

// Add regions (later regions have higher priority)
hm.AddRect("backdrop", 0, 0, screenW, screenH, nil)
hm.AddRect("modal", modalX, modalY, modalW, modalH, nil)
hm.AddRect("button", btnX, btnY, btnW, btnH, "btn-data")

// Test hit at point
region := hm.Test(clickX, clickY)
if region != nil {
    fmt.Println("Hit:", region.ID, region.Data)
}

// Clear for next frame
hm.Clear()
```

### Handler

Combines HitMap with mouse state tracking (clicks, double-clicks, drags, hover).

```go
handler := mouse.NewHandler()

// In Update():
action := handler.HandleMouse(mouseMsg)

switch action.Type {
case mouse.ActionClick:
    fmt.Println("Clicked:", action.Region.ID)
case mouse.ActionDoubleClick:
    fmt.Println("Double-clicked:", action.Region.ID)
case mouse.ActionHover:
    if action.Region != nil {
        fmt.Println("Hovering:", action.Region.ID)
    }
case mouse.ActionScrollUp, mouse.ActionScrollDown:
    fmt.Println("Scroll delta:", action.Delta)
case mouse.ActionDrag:
    fmt.Println("Drag delta:", action.DragDX, action.DragDY)
}
```

### Action Types

- `ActionNone` - No action
- `ActionClick` - Left mouse button click
- `ActionDoubleClick` - Double-click (same region within 400ms)
- `ActionScrollUp/Down` - Scroll wheel
- `ActionScrollLeft/Right` - Shift+scroll or horizontal scroll
- `ActionDrag` - Mouse motion while dragging
- `ActionDragEnd` - Mouse release after drag
- `ActionHover` - Mouse motion (not dragging)

## Integration Guide

### Using with Existing Monitor Model

```go
type Model struct {
    // ... existing fields ...

    // Modal state
    myModal       *modal.Modal
    myModalOpen   bool
    mouseHandler  *mouse.Handler
}

func NewModel() Model {
    return Model{
        mouseHandler: mouse.NewHandler(),
    }
}

func (m Model) openMyModal() Model {
    m.myModalOpen = true
    m.myModal = modal.New("My Modal").
        AddSection(modal.Text("Content here")).
        AddSection(modal.Buttons(
            modal.Btn(" OK ", "ok"),
            modal.Btn(" Cancel ", "cancel"),
        ))
    m.myModal.Reset() // Reset focus/scroll state
    return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if m.myModalOpen {
        switch msg := msg.(type) {
        case tea.KeyMsg:
            action, cmd := m.myModal.HandleKey(msg)
            if action != "" {
                return m.handleMyModalAction(action)
            }
            return m, cmd

        case tea.MouseMsg:
            action := m.myModal.HandleMouse(msg, m.mouseHandler)
            if action != "" {
                return m.handleMyModalAction(action)
            }
            return m, nil
        }
    }

    // ... rest of Update ...
}

func (m Model) View() string {
    if m.myModalOpen {
        // Option 1: Modal renders over blank background
        return m.myModal.Render(m.width, m.height, m.mouseHandler)

        // Option 2: Use OverlayModal for dimmed background
        background := m.renderBaseView()
        modalContent := m.myModal.Render(m.width, m.height, m.mouseHandler)
        return OverlayModal(background, modalContent, m.width, m.height)
    }
    return m.renderBaseView()
}

func (m Model) handleMyModalAction(action string) (tea.Model, tea.Cmd) {
    switch action {
    case "ok":
        m.myModalOpen = false
        return m.doSomething()
    case "cancel":
        m.myModalOpen = false
        return m, nil
    }
    return m, nil
}
```

### Mouse Handler Lifecycle

The mouse handler should be cleared and repopulated each frame:

1. **Clear**: Call `handler.HitMap.Clear()` at start of Render (done automatically by modal.Render)
2. **Register**: Hit regions are registered during Render
3. **Test**: HandleMouse tests against registered regions

The modal's Render method handles clearing automatically, but if you have additional hit regions outside the modal, manage them separately.

## Migration Notes

### From Manual Hit Region Calculation

Before (manual):
```go
func (m Model) renderModal() string {
    // Render content
    content := renderButtons()

    // Calculate button positions (error-prone!)
    btn1X := modalX + padding + 0
    btn2X := modalX + padding + btn1Width + spacing

    // Register hit regions manually
    m.hitmap.AddRect("btn1", btn1X, buttonY, btn1Width, 1, nil)
    m.hitmap.AddRect("btn2", btn2X, buttonY, btn2Width, 1, nil)

    return content
}
```

After (declarative):
```go
m := modal.New("Title").
    AddSection(modal.Buttons(
        modal.Btn(" Button 1 ", "btn1"),
        modal.Btn(" Button 2 ", "btn2"),
    ))

// Hit regions are calculated from rendered content automatically
content := m.Render(screenW, screenH, handler)
```

### Modal Inventory Status

See [modal-inventory.md](../modal-inventory.md) for the migration status of each modal in the codebase. Priority migrations:

1. Confirmation dialogs (already use button pattern)
2. Form modals (use huh library, consider hybrid approach)
3. Stats/Handoffs modals (missing button pattern)

## File Locations

- Modal library: `pkg/monitor/modal/`
  - `options.go` - Variant, Option funcs, constants
  - `section.go` - Section interface, Text, Spacer, Buttons, Checkbox, When, Custom
  - `input.go` - Input, Textarea sections
  - `list.go` - List section
  - `modal.go` - Modal struct and methods
  - `layout.go` - buildLayout with render-measure-register pattern
  - `styles.go` - Style mappings to monitor styles

- Mouse library: `pkg/monitor/mouse/`
  - `mouse.go` - Rect, Region, HitMap, Handler, ActionType

- Tests:
  - `pkg/monitor/modal/modal_test.go`
  - `pkg/monitor/mouse/mouse_test.go`
