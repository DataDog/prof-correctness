package analysis

import (
	"bytes"
	"testing"

	"github.com/google/pprof/profile"
)

// buildPprof returns a minimal serialized google/pprof profile with one
// cpu/count sample of value 42 in function pprofFn.
func buildPprof(t *testing.T) []byte {
	t.Helper()
	fn := &profile.Function{ID: 1, Name: "pprofFn"}
	loc := &profile.Location{ID: 1, Line: []profile.Line{{Function: fn}}}
	p := &profile.Profile{
		SampleType: []*profile.ValueType{{Type: "cpu", Unit: "count"}},
		Function:   []*profile.Function{fn},
		Location:   []*profile.Location{loc},
		Sample:     []*profile.Sample{{Location: []*profile.Location{loc}, Value: []int64{42}}},
	}
	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("write pprof: %v", err)
	}
	return buf.Bytes()
}

// TestFromPprof covers the pprof adapter: root-first folded stacks, per-type
// values, and string + numeric labels flattened into the canonical map.
func TestFromPprof(t *testing.T) {
	outer := &profile.Function{ID: 1, Name: "outer"}
	inner := &profile.Function{ID: 2, Name: "inner"}
	// Location slice is leaf-first; folded stack should be root-first.
	leaf := &profile.Location{ID: 1, Line: []profile.Line{{Function: inner}}}
	root := &profile.Location{ID: 2, Line: []profile.Line{{Function: outer}}}
	p := &profile.Profile{
		DurationNanos: 2_000_000_000,
		SampleType: []*profile.ValueType{
			{Type: "samples", Unit: "count"},
			{Type: "cpu", Unit: "nanoseconds"},
		},
		Function: []*profile.Function{outer, inner},
		Location: []*profile.Location{leaf, root},
		Sample: []*profile.Sample{{
			Location: []*profile.Location{leaf, root},
			Value:    []int64{3, 300},
			// "thread id" (string) and "thread.id" (numeric) both canonicalize to
			// the LabelThreadID key: values must be merged, not overwritten.
			Label:    map[string][]string{"span id": {"abc"}, "thread id": {"main"}},
			NumLabel: map[string][]int64{"thread.id": {7}},
		}},
	}

	ps := FromPprof(p)
	// All pprof sample types share the profile's single duration.
	if ps.Duration("samples") != 2 || ps.Duration("cpu") != 2 {
		t.Errorf("durations = samples:%v cpu:%v, want 2/2", ps.Duration("samples"), ps.Duration("cpu"))
	}

	samp, ok := ps.Samples("samples")
	if !ok || len(samp) != 1 {
		t.Fatalf("samples type missing: %+v", ps.SampleTypes())
	}
	if samp[0].Stack != "outer;inner" {
		t.Errorf("folded stack = %q, want outer;inner", samp[0].Stack)
	}
	if samp[0].Val != 3 {
		t.Errorf("samples value = %d, want 3", samp[0].Val)
	}
	// Second sample type shares the stack with its own value.
	if cpu, _ := ps.Samples("cpu"); len(cpu) != 1 || cpu[0].Val != 300 {
		t.Errorf("cpu samples = %+v, want val 300", cpu)
	}
	// String label passes through; numeric label is canonicalized + stringified.
	if got := samp[0].Labels["span id"]; len(got) != 1 || got[0] != "abc" {
		t.Errorf("span id label = %v, want [abc]", got)
	}
	if got := samp[0].Labels[LabelThreadID]; len(got) != 2 || got[0] != "7" || got[1] != "main" {
		t.Errorf("thread id label = %v, want [7 main] (string + numeric merged)", got)
	}
}
