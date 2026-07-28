// google/pprof adapter: maps a pprof profile into the neutral ProfileSet.
package analysis

import (
	"sort"
	"strconv"
	"strings"

	"github.com/google/pprof/profile"
)

// FromPprof builds a ProfileSet from a google/pprof profile. pprof carries its
// labels in Sample.Label / NumLabel; keys are run through canonKey (a no-op for
// the space-form keys Datadog pprof profilers already emit).
func FromPprof(prof *profile.Profile) *ProfileSet {
	ps := newProfileSet()
	ps.DurationSecs = float64(prof.DurationNanos) / 1e9

	// Merge identical locations/samples so counts are consistent.
	_ = prof.Aggregate(true, true, false, false, false, false)
	prof = prof.Compact()

	for _, sample := range prof.Sample {
		stack := foldPprofStack(sample)
		labels := pprofLabels(sample)
		for i, st := range prof.SampleType {
			ps.add(st.Type, StackSample{Stack: stack, Val: sample.Value[i], Labels: labels})
		}
	}
	return ps.finalize()
}

// pprofLabels flattens a sample's string and numeric labels into the canonical
// label map (values sorted for stable comparison).
func pprofLabels(sample *profile.Sample) map[string][]string {
	labels := map[string][]string{}
	for k, v := range sample.Label {
		cp := append([]string(nil), v...)
		sort.Strings(cp)
		labels[canonKey(k)] = cp
	}
	for k, v := range sample.NumLabel {
		vals := make([]string, 0, len(v))
		for _, n := range v {
			vals = append(vals, strconv.FormatInt(n, 10))
		}
		sort.Strings(vals)
		labels[canonKey(k)] = vals
	}
	return labels
}

// foldPprofStack renders a pprof sample's locations as a root-first folded
// stack (outermost frame first), matching the historical analyzer output.
func foldPprofStack(sample *profile.Sample) string {
	var frames []string
	for i := range sample.Location {
		loc := sample.Location[len(sample.Location)-i-1]
		for j := range loc.Line {
			line := loc.Line[len(loc.Line)-j-1]
			frames = append(frames, line.Function.Name)
		}
	}
	return strings.Join(frames, ";")
}
