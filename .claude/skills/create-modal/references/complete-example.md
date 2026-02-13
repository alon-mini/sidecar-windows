# Complete Modal Example

Full plugin implementation showing a delete confirmation modal with keyboard and mouse handling.

```go
type Plugin struct {
    // ...
    deleteModal      *modal.Modal
    deleteModalWidth int
    targetWorktree   *Worktree
    mouseHandler     *mouse.Handler
}

func (p *Plugin) ensureDeleteModal() {
    if p.targetWorktree == nil {
        return
    }

    modalW := 58
    if modalW > p.width-4 {
        modalW = p.width - 4
    }
    if modalW < 30 {
        modalW = 30
    }

    if p.deleteModal != nil && p.deleteModalWidth == modalW {
        return
    }
    p.deleteModalWidth = modalW

    wt := p.targetWorktree
    p.deleteModal = modal.New("Delete Worktree?",
        modal.WithWidth(modalW),
        modal.WithVariant(modal.VariantDanger),
        modal.WithPrimaryAction("delete"),
    ).
        AddSection(modal.Text("Name: " + wt.Name)).
        AddSection(modal.Text("Path: " + wt.Path)).
        AddSection(modal.Spacer()).
        AddSection(modal.Buttons(
            modal.Btn(" Delete ", "delete", modal.BtnDanger()),
            modal.Btn(" Cancel ", "cancel"),
        ))
}

func (p *Plugin) handleDeleteModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    p.ensureDeleteModal()
    if p.deleteModal == nil {
        return p, nil
    }

    action, cmd := p.deleteModal.HandleKey(msg)
    switch action {
    case "delete":
        return p.executeDelete()
    case "cancel":
        p.showingDeleteModal = false
        p.deleteModal = nil
        return p, nil
    }
    return p, cmd
}

func (p *Plugin) handleDeleteModalMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
    if p.deleteModal == nil {
        return p, nil
    }

    action := p.deleteModal.HandleMouse(msg, p.mouseHandler)
    switch action {
    case "delete":
        return p.executeDelete()
    case "cancel":
        p.showingDeleteModal = false
        p.deleteModal = nil
        return p, nil
    }
    return p, nil
}

func (p *Plugin) renderDeleteView(width, height int) string {
    p.ensureDeleteModal()
    background := p.renderListView(width, height)
    rendered := p.deleteModal.Render(width, height, p.mouseHandler)
    return ui.OverlayModal(background, rendered, width, height)
}
```
