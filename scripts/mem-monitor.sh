#!/bin/bash
# mem-monitor.sh - Monitor sidecar memory over time
# Usage: ./scripts/mem-monitor.sh [port] [interval_seconds]
# Requires SIDECAR_PPROF=1 sidecar to be running

set -e

PORT="${1:-6060}"
INTERVAL="${2:-60}"
LOG="mem-$(date +%Y%m%d-%H%M%S).log"
PPROF_URL="http://localhost:${PORT}"

echo "Monitoring sidecar memory (port=$PORT, interval=${INTERVAL}s)"
echo "Logging to $LOG"
echo "Press Ctrl+C to stop"
echo ""

# Header
echo "time,heap_alloc_bytes,heap_inuse_bytes,goroutines,rss_mb" | tee "$LOG"

while true; do
    # Get heap stats from pprof debug endpoint
    HEAP_STATS=$(curl -s "${PPROF_URL}/debug/pprof/heap?debug=1" 2>/dev/null | head -30)
    HEAP_ALLOC=$(echo "$HEAP_STATS" | grep "HeapAlloc = " | awk '{print $3}' || echo "N/A")
    HEAP_INUSE=$(echo "$HEAP_STATS" | grep "HeapInuse = " | awk '{print $3}' || echo "N/A")

    # Get goroutine count (first line of goroutine profile shows count)
    GOROUTINES=$(curl -s "${PPROF_URL}/debug/pprof/goroutine?debug=1" 2>/dev/null | head -1 | grep -o '[0-9]*' | head -1 || echo "N/A")

    # Get RSS from ps (in MB)
    PID=$(pgrep -x sidecar 2>/dev/null || echo "")
    if [ -n "$PID" ]; then
        RSS=$(ps -o rss= -p "$PID" 2>/dev/null | awk '{printf "%.1f", $1/1024}' || echo "N/A")
    else
        RSS="N/A"
    fi

    TIMESTAMP=$(date +%H:%M:%S)
    echo "${TIMESTAMP},${HEAP_ALLOC},${HEAP_INUSE},${GOROUTINES},${RSS}" | tee -a "$LOG"

    sleep "$INTERVAL"
done
