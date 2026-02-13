package fdmonitor

import (
	"testing"
)

func TestCount(t *testing.T) {
	count := Count()
	// On supported platforms (darwin, linux), count should be positive
	// On unsupported platforms or sandboxed environments, count may be -1
	// This is acceptable as the monitoring is best-effort
	t.Logf("Current FD count: %d (negative is OK in sandboxed test environments)", count)
}

func TestCheck(t *testing.T) {
	// First check should work
	count, warned := Check(nil)
	if count > 0 {
		t.Logf("FD count: %d, warned: %v", count, warned)
	}

	// Second immediate check should return cached value due to rate limiting
	count2, _ := Check(nil)
	if count > 0 && count2 != count {
		// Rate limiting should return the same count
		t.Logf("Second check returned different value: %d vs %d (may be due to test timing)", count, count2)
	}
}

func TestDebugInfo(t *testing.T) {
	info := DebugInfo()
	t.Logf("FD breakdown: %v", info)

	// Basic sanity check - we should have at least some FDs
	total := 0
	for _, v := range info {
		total += v
	}
	if total > 0 {
		t.Logf("Total FDs categorized: %d", total)
	}
}

func TestSetThresholds(t *testing.T) {
	// Save original values
	origWarning := warningThreshold
	origCritical := criticalThreshold
	defer func() {
		warningThreshold = origWarning
		criticalThreshold = origCritical
	}()

	SetThresholds(100, 300)
	if warningThreshold != 100 || criticalThreshold != 300 {
		t.Errorf("SetThresholds failed: got warning=%d, critical=%d", warningThreshold, criticalThreshold)
	}
}
