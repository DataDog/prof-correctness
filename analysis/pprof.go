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
	durSecs := float64(prof.DurationNanos) / 1e9

	// Merge identical locations/samples so counts are consistent.
	_ = prof.Aggregate(true, true, false, false, false, false)
	prof = prof.Compact()

	typeTotals := make([]int64, len(prof.SampleType))
	for _, sample := range prof.Sample {
		stack := foldPprofStack(sample)
		labels := pprofLabels(sample)
		for i, st := range prof.SampleType {
			ps.add(st.Type, StackSample{Stack: stack, Val: sample.Value[i], Labels: labels})
			typeTotals[i] += sample.Value[i]
		}
	}
	// All sample types in a pprof profile share its single duration.
	for i, st := range prof.SampleType {
		ps.addProfileDuration(st.Type, typeTotals[i], durSecs)
	}
	return ps.finalize()
}

// pprofLabels flattens a sample's string and numeric labels into the canonical
// label map. Values are appended (not overwritten) so distinct raw keys that
// canonicalize to the same key - e.g. a string "thread id" and a numeric
// "thread.id" - both survive; each key's values are sorted for stable
// comparison.
func pprofLabels(sample *profile.Sample) map[string][]string {
	labels := map[string][]string{}
	for k, v := range sample.Label {
		key := canonKey(k)
		labels[key] = append(labels[key], v...)
	}
	for k, v := range sample.NumLabel {
		key := canonKey(k)
		for _, n := range v {
			labels[key] = append(labels[key], strconv.FormatInt(n, 10))
		}
	}
	for k := range labels {
		sort.Strings(labels[k])
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
