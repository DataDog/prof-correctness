package analysis

import (
	"io"
	"regexp"
	"testing"
)

// Folded alloc-space stacks from CI run 32003823187 (ddprof_live_heap flake at
// 94% vs expected 100%). Values are byte counts from the captured profile.
const (
	ddprofLiveHeapAllocSpaceRegex = `^.*;main;(allocate_memory\(unsigned long\);operator new\(unsigned long\)|leak_function\(int\))$`
	ddprofLiveHeapOldAllocSpaceRegex = `^.*;main;allocate_memory\(unsigned long\);operator new\(unsigned long\)$`
)

var ddprofLiveHeapCIAllocSpaceSamples = []StackSample{
	{Stack: "test;_start;__libc_start_main;main;allocate_memory(unsigned long);operator new(unsigned long)", Val: 14799788},
	{Stack: "test;_start;__libc_start_main;main;leak_function(int)", Val: 928039},
}

func matchingPercent(prof []StackSample, regex string) int64 {
	rx := regexp.MustCompile(regex)
	var total, matching int64
	for _, ss := range prof {
		total += ss.Val
		if rx.MatchString(ss.Stack) {
			matching += ss.Val
		}
	}
	if total == 0 {
		return 0
	}
	return matching * 100 / total
}

func assertPercentAssertion(t *testing.T, prof []StackSample, regex string, wantPct, errorMargin int64, wantFail bool) {
	t.Helper()
	r := NewStdReporter(io.Discard, io.Discard)
	Run(r, func() {
		var hasFailures bool
		assertStackWithFailureHandling(r, prof, regex, Optional[float64]{}, NewOptionalFrom(wantPct), errorMargin, nil, false, &hasFailures)
	})
	if r.Failed() != wantFail {
		t.Errorf("Failed() = %v, want %v", r.Failed(), wantFail)
	}
}

// TestDDProfLiveHeap_AllocSpaceRegex locks the alloc-space assertion in
// scenarios/ddprof_live_heap/expected_profile.json: the pre-fix regex only
// matched operator-new bytes (~94%) and missed leak_function malloc (~5%).
func TestDDProfLiveHeap_AllocSpaceRegex(t *testing.T) {
	const wantPct int64 = 100
	const errorMargin int64 = 5

	if got := matchingPercent(ddprofLiveHeapCIAllocSpaceSamples, ddprofLiveHeapOldAllocSpaceRegex); got != 94 {
		t.Fatalf("old regex matching = %d%%, want 94%%", got)
	}
	if got := matchingPercent(ddprofLiveHeapCIAllocSpaceSamples, ddprofLiveHeapAllocSpaceRegex); got != 100 {
		t.Fatalf("new regex matching = %d%%, want 100%%", got)
	}

	cases := []struct {
		name     string
		regex    string
		wantFail bool
	}{
		{
			name:     "old_regex_misses_leak_function",
			regex:    ddprofLiveHeapOldAllocSpaceRegex,
			wantFail: true,
		},
		{
			name:     "new_regex_covers_both_paths",
			regex:    ddprofLiveHeapAllocSpaceRegex,
			wantFail: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertPercentAssertion(t, ddprofLiveHeapCIAllocSpaceSamples, tc.regex, wantPct, errorMargin, tc.wantFail)
		})
	}
}
