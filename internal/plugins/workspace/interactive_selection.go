package workspace

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	app "github.com/marcus/sidecar/internal/app"
	"github.com/marcus/sidecar/internal/mouse"
	"github.com/marcus/sidecar/internal/ui"
)

// interactiveColAtX maps a viewport X coordinate to a visual column in the given line.
// The returned column is in visual space (post-tab-expansion, accounting for multi-width chars).
func (p *Plugin) interactiveColAtX(x, lineIdx int) (int, bool) {
	relX := x - p.selection.ViewRect.X - panelOverhead/2
	if relX < 0 {
		return 0, false
	}

	buf := p.interactiveOutputBuffer()
	if buf == nil {
		return 0, true
	}
	if lineIdx < 0 || lineIdx >= buf.LineCount() {
		return 0, true
	}

	lines := buf.LinesRange(lineIdx, lineIdx+1)
	if len(lines) == 0 {
		return 0, true
	}
	expanded := ui.ExpandTabs(lines[0], tabStopWidth)

	return ui.VisualColAtRelativeX(expanded, relX), true
}

// interactiveCharAtXY maps viewport coordinates to line index + visual column.
func (p *Plugin) interactiveCharAtXY(x, y int) (int, int, bool) {
	lineIdx, ok := p.interactiveLineIndexAtY(y)
	if !ok {
		return 0, 0, false
	}
	col, ok := p.interactiveColAtX(x, lineIdx)
	return lineIdx, col, ok
}

func (p *Plugin) interactiveLineIndexAtY(y int) (int, bool) {
	if p.interactiveState == nil || !p.interactiveState.Active {
		return 0, false
	}
	if p.selection.ViewRect.W == 0 || p.selection.ViewRect.H == 0 {
		return 0, false
	}
	if p.interactiveState.VisibleEnd <= p.interactiveState.VisibleStart {
		return 0, false
	}

	contentRow := y - p.selection.ViewRect.Y - 1 // account for top border
	if contentRow < 0 {
		return 0, false
	}
	if p.interactiveState.ContentRowOffset <= 0 {
		return 0, false
	}
	outputRow := contentRow - p.interactiveState.ContentRowOffset
	if outputRow < 0 {
		return 0, false
	}
	lineIdx := p.interactiveState.VisibleStart + outputRow
	if lineIdx < p.interactiveState.VisibleStart || lineIdx >= p.interactiveState.VisibleEnd {
		return 0, false
	}
	return lineIdx, true
}

// prepareInteractiveDrag stores the click position and starts drag tracking
// without initializing selection. Selection only activates on actual drag motion.
func (p *Plugin) prepareInteractiveDrag(action mouse.MouseAction) tea.Cmd {
	if action.Region == nil {
		return nil
	}
	// Set ViewRect before charAtXY so interactiveLineIndexAtY can use it
	p.selection.ViewRect = action.Region.Rect

	lineIdx, col, ok := p.interactiveCharAtXY(action.X, action.Y)
	if !ok {
		p.selection.Clear()
		return nil
	}

	p.selection.PrepareDrag(lineIdx, col, action.Region.Rect)
	p.autoScrollOutput = false

	p.mouseHandler.StartDrag(action.X, action.Y, regionPreviewPane, lineIdx)
	return nil
}

func (p *Plugin) handleInteractiveSelectionDrag(action mouse.MouseAction) tea.Cmd {
	lineIdx, col, ok := p.interactiveCharAtXY(action.X, action.Y)
	if !ok {
		return nil
	}

	p.selection.HandleDrag(lineIdx, col)
	return nil
}

func (p *Plugin) finishInteractiveSelection() tea.Cmd {
	p.selection.FinishDrag()
	return nil
}

func (p *Plugin) interactiveOutputBuffer() *OutputBuffer {
	if p.shellSelected {
		shell := p.getSelectedShell()
		if shell != nil && shell.Agent != nil {
			return shell.Agent.OutputBuf
		}
		return nil
	}
	wt := p.selectedWorktree()
	if wt == nil || wt.Agent == nil {
		return nil
	}
	return wt.Agent.OutputBuf
}

func (p *Plugin) interactiveSelectionLines() []string {
	if !p.selection.HasSelection() {
		return nil
	}
	buf := p.interactiveOutputBuffer()
	if buf == nil {
		return nil
	}

	lineCount := buf.LineCount()
	if lineCount == 0 {
		return nil
	}

	startLine := p.selection.Start.Line
	endLine := p.selection.End.Line
	if startLine > endLine {
		startLine, endLine = endLine, startLine
	}
	if startLine < 0 {
		startLine = 0
	}
	if endLine >= lineCount {
		endLine = lineCount - 1
	}
	if endLine < startLine {
		return nil
	}

	lines := buf.LinesRange(startLine, endLine+1)
	if len(lines) == 0 {
		return nil
	}

	return p.selection.SelectedText(lines, startLine, tabStopWidth)
}

func (p *Plugin) interactiveVisibleLines() []string {
	if p.interactiveState == nil || !p.interactiveState.Active {
		return nil
	}
	buf := p.interactiveOutputBuffer()
	if buf == nil {
		return nil
	}
	start := p.interactiveState.VisibleStart
	end := p.interactiveState.VisibleEnd
	if end <= start {
		return nil
	}
	return buf.LinesRange(start, end)
}

func (p *Plugin) copyInteractiveSelectionCmd() tea.Cmd {
	return func() tea.Msg {
		lines := p.interactiveSelectionLines()
		if len(lines) == 0 {
			lines = p.interactiveVisibleLines()
		}
		if len(lines) == 0 {
			return app.ToastMsg{Message: "No output to copy", Duration: 2 * time.Second}
		}

		stripped := make([]string, 0, len(lines))
		for _, line := range lines {
			stripped = append(stripped, ansi.Strip(line))
		}
		text := strings.Join(stripped, "\n")
		if err := clipboard.WriteAll(text); err != nil {
			return app.ToastMsg{Message: "Copy failed: " + err.Error(), Duration: 2 * time.Second, IsError: true}
		}

		return app.ToastMsg{Message: fmt.Sprintf("Copied %d line(s)", len(stripped)), Duration: 2 * time.Second}
	}
}
