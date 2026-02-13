---
name: profile-memory
description: >
  Profile memory usage in sidecar using Go pprof, system tools, and heap analysis.
  Covers identifying memory leaks, goroutine leaks, file descriptor accumulation,
  and CPU profiling. Use when investigating memory issues, profiling performance,
  debugging memory leaks, or diagnosing unresponsive plugins.
disable-model-invocation: true
---

# Memory Profiling for Sidecar

## Quick Triage

| Symptom | Tool | Action |
|---------|------|--------|
| High RSS / memory growth | vmmap, pprof heap | Check system memory, then heap profile |
| Too many open files | lsof | Check FD count and breakdown |
| High CPU | pprof cpu, ps | Capture CPU profile |
| Goroutine leak | pprof goroutines | Check goroutine count and stacks |
| Plugin unresponsive | lsof + goroutines | Check SQLite locks, blocked goroutines |

```
Triage flow: Is RSS high? -> Check FD count -> Check vmmap -> Check heap profile
```

## System-Level Profiling (No pprof Required)

### Find the Process

```bash
pgrep -f sidecar
ps aux | grep sidecar
```

### Basic Stats

```bash
# RSS, VSZ, CPU%, thread count
ps -o pid,rss,vsz,%cpu,nlwp -p <PID>

# Human-readable RSS
ps -o pid,rss -p <PID> | awk 'NR>1{printf "%d MB\n", $2/1024}'

# Watch over time
while true; do ps -o rss,%cpu -p <PID> | tail -1; sleep 5; done
```

### System Memory

**macOS (vmmap):**
```bash
vmmap --summary <PID>
```
Key sections: VM_ALLOCATE (Go heap), MALLOC (C heap/SQLite), Physical footprint, Swapped.

Red flags:
- RESIDENT SIZE >500MB for idle sidecar
- REGION COUNT in thousands for VM_ALLOCATE = mmap leak
- SWAPPED SIZE growing = memory pressure

**Linux:**
```bash
cat /proc/<PID>/status | grep -E 'VmRSS|VmSize|Threads'
pmap -x <PID> | tail -5
ls /proc/<PID>/fd | wc -l
```

### File Descriptor Analysis

```bash
# Count and breakdown
lsof -p <PID> | wc -l
lsof -p <PID> | awk '{print $5}' | sort | uniq -c | sort -rn

# Find leaked files
lsof -p <PID> | grep REG | awk '{print $9}' | sort | uniq -c | sort -rn | head -20

# Check session file leaks
lsof -p <PID> | grep -c '\.claude/projects'
lsof -p <PID> | grep -c '\.codex/sessions'

# Watch FD count
while true; do echo "$(date): $(lsof -p <PID> 2>/dev/null | wc -l) FDs"; sleep 30; done
```

Healthy baselines: Total FDs 50-150, REG files 10-30, PIPEs 10-30, DIRs 5-15.

Red flags: 1000+ total FDs, same file opened 4+ times, growing count over time.

### Thread Count

```bash
# macOS
ps -M -p <PID> | wc -l
# Linux
ls /proc/<PID>/task | wc -l
# Expected: 20-60 threads. 100+ = goroutine leak likely
```

## Go pprof (Runtime Profiling)

### Enable pprof

```bash
SIDECAR_PPROF=1 sidecar        # Default port 6060
SIDECAR_PPROF=6061 sidecar     # Custom port
```

### Heap Profile (Current Allocations)

```bash
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof -top heap.prof
go tool pprof heap.prof   # Interactive: top20, list <func>, web
```

### Allocs Profile (All Allocations Since Start)

```bash
curl http://localhost:6060/debug/pprof/allocs > allocs.prof
go tool pprof -top allocs.prof
```

### Goroutine Profile

```bash
# Count
curl -s http://localhost:6060/debug/pprof/goroutine?debug=1 | head -1

# Full stacks
curl http://localhost:6060/debug/pprof/goroutine?debug=2 > goroutines.txt

# Find stuck goroutines
grep -A5 'runtime.chanrecv' goroutines.txt
```

### Memory Stats

```bash
curl http://localhost:6060/debug/pprof/heap?debug=1 | head -30
```

### Compare Snapshots (Best for Finding Leaks)

```bash
curl http://localhost:6060/debug/pprof/heap > heap1.prof
# Wait (1 hour or overnight)
curl http://localhost:6060/debug/pprof/heap > heap2.prof
go tool pprof -base heap1.prof heap2.prof
# Then: top20
```

### Continuous Monitoring

```bash
./scripts/mem-monitor.sh         # Default: port 6060, 60s interval
./scripts/mem-monitor.sh 6061 30 # Custom port and interval
```

Output: CSV `time,heap_alloc_bytes,heap_inuse_bytes,goroutines,rss_mb` to `mem-YYYYMMDD-HHMMSS.log`.

### CPU Profile

```bash
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof -top cpu.prof
go tool pprof cpu.prof   # Interactive: top20, list <func>, web
```

### Web UI

```bash
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap
```

Requires Graphviz (`brew install graphviz` / `apt install graphviz`) for flame graphs.

## What to Look For

**Goroutine leaks:** Count should stabilize at 20-50 after startup. Steady growth = leak. Search for `runtime.chanrecv1`, `runtime.chansend1`, `time.Sleep`.

**Heap growth:** `HeapAlloc` should stabilize after loading sessions. Consistent upward trend = leak.

**Common pprof signatures:** `bufio.Scanner` (buffer not returned), `json.Unmarshal` (large objects retained), `append` in loops (slice growing), channel operations (blocked senders/receivers).

## Diagnosing Unresponsive Plugins

1. Check goroutines for blocked operations:
   ```bash
   curl -s http://localhost:6060/debug/pprof/goroutine?debug=2 | grep -A10 'monitor\|td'
   ```
2. Check SQLite locks: `lsof -p <PID> | grep '\.db'`
3. Check stuck tea.Cmd goroutines:
   ```bash
   curl -s http://localhost:6060/debug/pprof/goroutine?debug=2 | grep -B2 'fetchData\|FetchData'
   ```
4. Check for swap pressure: `vmmap --summary <PID> | grep -E 'Physical|Swapped'`

Common causes: SQLite locked by concurrent td CLI, memory pressure causing swap thrashing, goroutine blocked on unread channel, accumulating fetchData() goroutines.

## Known Leak Patterns

### File Descriptor Accumulation (Adapter/Watcher Layer)

Files: `internal/adapter/claudecode/watcher.go` etc. Uses fsnotify to watch directories.

Symptoms: RSS grows to 5-15GB overnight, lsof shows 1000+ session files open.

```bash
lsof -p <PID> | grep '\.claude/projects' | wc -l
lsof -p <PID> | grep '\.codex/sessions' | wc -l
```

### OutputBuffer Substring Retention

`OutputBuffer.Update()` uses `strings.Split()` creating substrings sharing backing array. NOT a leak if `Update()` keeps being called. Only retains if polling stops.

### Poll Chain Duplication

Entering interactive mode without incrementing generation counter creates parallel poll chains. Doubles CPU but no memory leak. Fix: increment `shellPollGeneration` or `pollGeneration` on entry.

### Go MADV_FREE (macOS)

Go runtime uses `MADV_FREE` marking pages reusable without returning to OS. RSS appears high but memory is available. NOT a leak. Use `vmmap --summary` and compare "Physical footprint (peak)" vs current.

## Investigation Workflow

1. `ps -o pid,rss,%cpu -p <PID>` -- establish baseline
2. `lsof -p <PID> | wc -l` -- if >200, focus on FD leak
3. `lsof -p <PID> | grep REG | awk '{print $9}' | sort | uniq -c | sort -rn | head -20`
4. `vmmap --summary <PID>` -- RESIDENT vs VIRTUAL, region count
5. If pprof available: heap snapshot comparison over time
6. If pprof unavailable: restart with `SIDECAR_PPROF=1`, reproduce the leak

## Overnight Test Procedure

1. Start: `SIDECAR_PPROF=1 sidecar`
2. Baseline: `curl .../heap > baseline.prof`
3. Note goroutine count
4. Leave overnight (`caffeinate -i` on macOS)
5. Morning: capture new profiles
6. Compare: `go tool pprof -base baseline.prof morning.prof`
7. Check goroutines for stuck ones

## Notes

- pprof adds ~1-2MB overhead, no TUI performance impact
- HTTP server runs in separate goroutine, localhost only
- On macOS, vmmap/lsof need no special permissions for own processes
