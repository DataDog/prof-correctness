// Package reporter ships per-assertion and per-scenario metrics from
// prof-correctness runs to a Datadog backend. Default is no-op; the Datadog
// implementation is activated via DD_PROF_CORRECTNESS_REPORT=1 plus the
// usual DD_API_KEY_STAGING / DD_SITE env vars.
//
// Tag set is deliberately low-cardinality (scenario, language, profile_type,
// assertion_kind, runner_label) so we stay well under Datadog's custom-metric
// budget. High-cardinality context (git_sha, ci_run_id, individual stack
// regexes) is intentionally not on metric tags — it belongs on CI Visibility
// or in failure logs.
package reporter

import (
	"strings"
	"sync"
)

// AssertionKind is the kind of assertion that was evaluated. Constants are
// chosen to match what `analysis/analysis.go` already distinguishes
// (per-stack value, per-stack percent, per-typed-stacks matching sum).
type AssertionKind string

const (
	AssertionKindValue        AssertionKind = "value"
	AssertionKindPercent      AssertionKind = "percent"
	AssertionKindMatchingSum  AssertionKind = "matching_sum"
)

// AssertionContext is what the recorder needs to know per assertion. All
// fields are required (no zero-value sentinels).
type AssertionContext struct {
	Scenario     string        // e.g. "python_many_threads"
	Language     string        // first folder segment, e.g. "python", "ddprof", "ruby"
	ProfileType  string        // e.g. "wall-time", "cpu-time", "alloc-space"
	Kind         AssertionKind
	ErrorPct     float64 // unsigned: percent deviation from expected value
	ErrorMargin  int64   // configured threshold (percent)
	Passed       bool
	AllowFailure bool // true when the assertion is in the "first profile" tolerance window
}

// MetricsRecorder collects per-assertion and per-scenario events.
// All methods must be safe to call from multiple goroutines (Go tests can
// run scenarios in parallel via t.Parallel()).
type MetricsRecorder interface {
	RecordAssertion(ev AssertionEvent)
	RecordScenarioResult(scenario, language string, failed bool, durationSeconds float64)
	// Flush sends any buffered metrics to the backend and returns any
	// transport error. Implementations must never panic; callers may ignore
	// the error if they choose to (correctness tests should not fail because
	// reporting failed).
	Flush() error
}

// NoopRecorder discards everything. Used when reporting is disabled.
type NoopRecorder struct{}

func (NoopRecorder) RecordAssertion(AssertionEvent)                            {}
func (NoopRecorder) RecordScenarioResult(string, string, bool, float64)        {}
func (NoopRecorder) Flush() error                                              { return nil }

// LanguageFromScenarioFolder extracts the language tag from a scenario folder
// path like "scenarios/python_many_threads" → "python". Convention is that
// each scenario folder is named "<lang>_<topic>"; the first underscore-
// separated segment is the language. Unknown/unparseable folders return
// "unknown" rather than failing — we never want the recorder to break a test.
func LanguageFromScenarioFolder(folder string) string {
	// Strip a "scenarios/" prefix if present so callers can pass either.
	folder = strings.TrimPrefix(folder, "scenarios/")
	folder = strings.TrimPrefix(folder, "./scenarios/")
	if folder == "" {
		return "unknown"
	}
	if idx := strings.IndexByte(folder, '_'); idx > 0 {
		return folder[:idx]
	}
	// No underscore — treat whole folder as language (e.g. "ddprof").
	return folder
}

// ScenarioFromFolder strips the "scenarios/" prefix so the tag is just the
// scenario name (e.g. "python_many_threads"), matching what a human would
// write in a Datadog query.
func ScenarioFromFolder(folder string) string {
	folder = strings.TrimPrefix(folder, "scenarios/")
	folder = strings.TrimPrefix(folder, "./scenarios/")
	return folder
}

// --- Thread-safety helper ----------------------------------------------------

// safeBuffer is a small concurrent-safe slice wrapper used by the Datadog
// recorder to accumulate points between Flush() calls. Kept here so concrete
// implementations can reuse it.
type safeBuffer[T any] struct {
	mu   sync.Mutex
	data []T
}

func (b *safeBuffer[T]) append(v T) {
	b.mu.Lock()
	b.data = append(b.data, v)
	b.mu.Unlock()
}

func (b *safeBuffer[T]) drain() []T {
	b.mu.Lock()
	out := b.data
	b.data = nil
	b.mu.Unlock()
	return out
}
