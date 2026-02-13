#!/bin/bash
# tmux-screenshot.sh - Capture sidecar screenshots via tmux
#
# Usage:
#   ./scripts/tmux-screenshot.sh start      - Start sidecar in tmux session
#   ./scripts/tmux-screenshot.sh attach     - Attach to the session (navigate, then detach with Ctrl+A D)
#   ./scripts/tmux-screenshot.sh capture NAME - Capture current view to docs/screenshots/NAME.html and NAME.png
#   ./scripts/tmux-screenshot.sh stop       - Quit sidecar and kill session
#
# Example workflow for LLM:
#   1. Run: ./scripts/tmux-screenshot.sh start
#   2. Run: tmux attach -t sidecar-screenshot  (in interact mode, press keys to navigate, then Ctrl+A D)
#   3. Run: ./scripts/tmux-screenshot.sh capture sidecar-td
#   4. Repeat 2-3 for other views
#   5. Run: ./scripts/tmux-screenshot.sh stop

set -e

SESSION_NAME="sidecar-screenshot"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/../docs/screenshots"

start_session() {
    # Kill any existing session
    tmux kill-session -t "$SESSION_NAME" 2>/dev/null || true
    
    # Get current terminal dimensions
    COLS=$(tput cols)
    LINES=$(tput lines)
    
    # Create output directory
    mkdir -p "$OUTPUT_DIR"
    
    # Start sidecar in tmux with current terminal size
    tmux new-session -d -s "$SESSION_NAME" -x "$COLS" -y "$LINES" "TERM=xterm-256color sidecar"
    
    echo "Started sidecar in tmux session '$SESSION_NAME' (${COLS}x${LINES})"
    echo "Next: attach with 'tmux attach -t $SESSION_NAME' or './scripts/tmux-screenshot.sh attach'"
    echo "Navigate to desired view, then detach with Ctrl+A D (or Ctrl+B D)"
    
    # Wait for sidecar to initialize
    sleep 2
}

attach_session() {
    if ! tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
        echo "Error: No session '$SESSION_NAME'. Run './scripts/tmux-screenshot.sh start' first."
        exit 1
    fi
    exec tmux attach -t "$SESSION_NAME"
}

capture_screenshot() {
    local name="$1"
    if [ -z "$name" ]; then
        name="sidecar-$(date +%Y%m%d-%H%M%S)"
    fi
    
    if ! tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
        echo "Error: No session '$SESSION_NAME'. Run './scripts/tmux-screenshot.sh start' first."
        exit 1
    fi
    
    local txt_file="$OUTPUT_DIR/$name.txt"
    local html_file="$OUTPUT_DIR/$name.html"
    local png_file="$OUTPUT_DIR/$name.png"
    
    # Get terminal dimensions from tmux
    local cols=$(tmux display-message -t "$SESSION_NAME" -p '#{pane_width}')
    local lines=$(tmux display-message -t "$SESSION_NAME" -p '#{pane_height}')
    
    # Capture with ANSI codes
    tmux capture-pane -t "$SESSION_NAME" -e -p > "$txt_file"
    
    # Convert to PNG using termshot (preferred) or fall back to aha+wkhtmltoimage
    if command -v termshot &>/dev/null; then
        termshot --raw-read "$txt_file" --columns "$cols" --filename "$png_file" 2>/dev/null
        echo "Captured: $png_file (${cols}x${lines})"
        
        # Also generate HTML if aha is available
        if command -v aha &>/dev/null; then
            cat "$txt_file" | aha --black > "$html_file"
            echo "Also saved: $html_file"
        fi
        rm -f "$txt_file"
    elif command -v aha &>/dev/null; then
        # Fallback: use aha for HTML
        cat "$txt_file" | aha --black > "$html_file"
        rm -f "$txt_file"
        echo "Captured: $html_file (${cols}x${lines})"
        echo "Tip: Install termshot (brew install homeport/tap/termshot) for better PNG output"
    else
        echo "Captured: $txt_file"
        echo "Install termshot (brew install homeport/tap/termshot) for PNG screenshots"
    fi
}

stop_session() {
    if ! tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
        echo "No session to stop."
        return
    fi
    
    # Send quit keys to sidecar
    tmux send-keys -t "$SESSION_NAME" q
    sleep 0.3
    tmux send-keys -t "$SESSION_NAME" y
    sleep 0.5
    
    # Kill the session
    tmux kill-session -t "$SESSION_NAME" 2>/dev/null || true
    echo "Stopped sidecar session."
}

list_screenshots() {
    echo "Screenshots in $OUTPUT_DIR:"
    
    # Group by base name to show HTML and PNG together
    for html in "$OUTPUT_DIR"/*.html; do
        [ -e "$html" ] || continue
        local base=$(basename "$html" .html)
        local png="$OUTPUT_DIR/$base.png"
        
        echo "  $base:"
        echo "    - $base.html"
        if [ -f "$png" ]; then
            echo "    - $base.png"
        fi
    done
}

show_usage() {
    echo "Usage: $0 <command> [args]"
    echo ""
    echo "Commands:"
    echo "  start         Start sidecar in a tmux session"
    echo "  attach        Attach to the tmux session (navigate, then Ctrl+A/B D to detach)"
    echo "  capture NAME  Capture current view to docs/screenshots/NAME.html and NAME.png"
    echo "  list          List existing screenshots"
    echo "  stop          Quit sidecar and kill the session"
    echo ""
    echo "Workflow:"
    echo "  1. $0 start"
    echo "  2. $0 attach  (navigate in sidecar, then Ctrl+A D to detach)"
    echo "  3. $0 capture sidecar-td"
    echo "  4. Repeat 2-3 for other views"
    echo "  5. $0 stop"
}

case "${1:-}" in
    start)
        start_session
        ;;
    attach)
        attach_session
        ;;
    capture)
        capture_screenshot "$2"
        ;;
    list)
        list_screenshots
        ;;
    stop)
        stop_session
        ;;
    *)
        show_usage
        exit 1
        ;;
esac
