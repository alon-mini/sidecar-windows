package pricing

import (
	"math"
	"testing"
)

func TestModelCost_OpusNew(t *testing.T) {
	tests := []struct {
		model string
		desc  string
	}{
		{"claude-opus-4-5-20251101", "opus 4.5"},
		{"claude-opus-4-6-20260101", "opus 4.6"},
		{"claude-opus-4-5", "opus 4.5 short"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// 1M input + 1M output, no cache
			cost := ModelCost(tt.model, Usage{InputTokens: 1_000_000, OutputTokens: 1_000_000})
			// $5 in + $25 out = $30
			assertCost(t, 30.0, cost)
		})
	}
}

func TestModelCost_OpusOld(t *testing.T) {
	tests := []struct {
		model string
		desc  string
	}{
		{"claude-3-opus-20240229", "opus 3"},
		{"claude-opus-4-20250101", "opus 4.0"},
		{"claude-opus-4-1-20250301", "opus 4.1"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			cost := ModelCost(tt.model, Usage{InputTokens: 1_000_000, OutputTokens: 1_000_000})
			// $15 in + $75 out = $90
			assertCost(t, 90.0, cost)
		})
	}
}

func TestModelCost_Sonnet(t *testing.T) {
	tests := []struct {
		model string
		desc  string
	}{
		{"claude-sonnet-4-5-20250929", "sonnet 4.5"},
		{"claude-3-5-sonnet-20241022", "sonnet 3.5"},
		{"claude-sonnet-4-20250301", "sonnet 4"},
		{"claude-3-7-sonnet-20250219", "sonnet 3.7"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			cost := ModelCost(tt.model, Usage{InputTokens: 1_000_000, OutputTokens: 1_000_000})
			// $3 in + $15 out = $18
			assertCost(t, 18.0, cost)
		})
	}
}

func TestModelCost_HaikuNew(t *testing.T) {
	tests := []struct {
		model string
		desc  string
	}{
		{"claude-haiku-4-5-20251001", "haiku 4.5"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			cost := ModelCost(tt.model, Usage{InputTokens: 1_000_000, OutputTokens: 1_000_000})
			// $1 in + $5 out = $6
			assertCost(t, 6.0, cost)
		})
	}
}

func TestModelCost_Haiku35(t *testing.T) {
	tests := []struct {
		model string
		desc  string
	}{
		{"claude-3-5-haiku-20241022", "haiku 3.5"},
		{"claude-3-5-haiku-latest", "haiku 3.5 latest"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			cost := ModelCost(tt.model, Usage{InputTokens: 1_000_000, OutputTokens: 1_000_000})
			// $0.80 in + $4 out = $4.80
			assertCost(t, 4.80, cost)
		})
	}
}

func TestModelCost_HaikuOld(t *testing.T) {
	tests := []struct {
		model string
		desc  string
	}{
		{"claude-3-haiku-20240307", "haiku 3"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			cost := ModelCost(tt.model, Usage{InputTokens: 1_000_000, OutputTokens: 1_000_000})
			// $0.25 in + $1.25 out = $1.50
			assertCost(t, 1.50, cost)
		})
	}
}

func TestModelCost_UnknownDefaultsSonnet(t *testing.T) {
	cost := ModelCost("some-unknown-model", Usage{InputTokens: 1_000_000, OutputTokens: 1_000_000})
	// Defaults to sonnet: $3 in + $15 out = $18
	assertCost(t, 18.0, cost)
}

func TestModelCost_CacheRead(t *testing.T) {
	// Sonnet: $3/M input, cache read at 10% = $0.30/M
	cost := ModelCost("claude-sonnet-4-5-20250929", Usage{
		InputTokens: 200_000,
		CacheRead:   800_000,
	})
	// input: 200k * 3 / 1M = 0.60
	// cache read: 800k * 3 * 0.1 / 1M = 0.24
	// total: 0.84
	assertCost(t, 0.84, cost)
}

func TestModelCost_CacheWrite(t *testing.T) {
	// Sonnet: $3/M input, cache write at 1.25x = $3.75/M
	cost := ModelCost("claude-sonnet-4-5-20250929", Usage{
		InputTokens: 200_000,
		CacheWrite:  800_000,
	})
	// input: 200k * 3 / 1M = 0.60
	// cache write: 800k * 3 * 1.25 / 1M = 3.00
	// total: 3.60
	assertCost(t, 3.60, cost)
}

func TestModelCost_AllComponents(t *testing.T) {
	// Opus 4.5: $5/$25
	cost := ModelCost("claude-opus-4-5-20251101", Usage{
		InputTokens:  100_000,
		OutputTokens: 50_000,
		CacheRead:    400_000,
		CacheWrite:   200_000,
	})
	// input: 100k * 5 / 1M = 0.50
	// output: 50k * 25 / 1M = 1.25
	// cache read: 400k * 5 * 0.1 / 1M = 0.20
	// cache write: 200k * 5 * 1.25 / 1M = 1.25
	// total: 3.20
	assertCost(t, 3.20, cost)
}

func TestModelCost_ZeroTokens(t *testing.T) {
	cost := ModelCost("claude-opus-4-5-20251101", Usage{})
	if cost != 0 {
		t.Errorf("expected 0, got %f", cost)
	}
}

func TestClassifyModel(t *testing.T) {
	tests := []struct {
		model   string
		inRate  float64
		outRate float64
	}{
		{"claude-opus-4-6-20260101", 5.0, 25.0},
		{"claude-opus-4-5-20251101", 5.0, 25.0},
		{"claude-opus-4-1-20250301", 15.0, 75.0},
		{"claude-opus-4-20250101", 15.0, 75.0},
		{"claude-3-opus-20240229", 15.0, 75.0},
		{"claude-sonnet-4-5-20250929", 3.0, 15.0},
		{"claude-3-5-sonnet-20241022", 3.0, 15.0},
		{"claude-haiku-4-5-20251001", 1.0, 5.0},
		{"claude-3-5-haiku-20241022", 0.80, 4.0},
		{"claude-3-haiku-20240307", 0.25, 1.25},
		{"unknown-model", 3.0, 15.0},
	}
	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			tier := classifyModel(tt.model)
			if tier.inRate != tt.inRate {
				t.Errorf("inRate: expected %f, got %f", tt.inRate, tier.inRate)
			}
			if tier.outRate != tt.outRate {
				t.Errorf("outRate: expected %f, got %f", tt.outRate, tier.outRate)
			}
		})
	}
}

func assertCost(t *testing.T, expected, actual float64) {
	t.Helper()
	if math.Abs(expected-actual) > 0.01 {
		t.Errorf("expected cost %.2f, got %.4f", expected, actual)
	}
}
