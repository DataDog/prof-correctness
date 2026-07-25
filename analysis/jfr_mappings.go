package analysis

import (
	"strings"
	"unicode"

	"github.com/google/pprof/profile"
)

// jfrProfileMappings keeps the JFR event/profile naming policy in one place.
//
// The raw JFR event IDs are metadata-defined and differ between JDK built-ins
// and Datadog/java-profiler events.  github.com/grafana/jfr-parser normalises
// the supported event variants into pprof profiles; this table maps those
// normalised pprof sample-type sets to the file suffix used by prof-correctness
// expectations.
type jfrProfileMapping struct {
	metric      string
	sampleTypes []string
	name        string
}

var jfrProfileMappings = []jfrProfileMapping{
	{metric: "process_cpu", sampleTypes: []string{"cpu"}, name: "cpu"},
	{metric: "wall", sampleTypes: []string{"wall"}, name: "wall"},
	{metric: "memory", sampleTypes: []string{"alloc_in_new_tlab_objects", "alloc_in_new_tlab_bytes"}, name: "alloc_in_tlab"},
	{metric: "memory", sampleTypes: []string{"alloc_outside_tlab_objects", "alloc_outside_tlab_bytes"}, name: "alloc_outside_tlab"},
	{metric: "memory", sampleTypes: []string{"alloc_sample_objects", "alloc_sample_bytes"}, name: "alloc_sample"},
	{metric: "memory", sampleTypes: []string{"live"}, name: "live"},
	{metric: "memory", sampleTypes: []string{"malloc_objects", "malloc_bytes"}, name: "malloc"},
	{metric: "mutex", sampleTypes: []string{"contentions", "delay"}, name: "lock"},
	{metric: "block", sampleTypes: []string{"contentions", "delay"}, name: "park"},
}

func jfrProfileName(metric string, prof *profile.Profile) string {
	sampleTypes := make([]string, 0, len(prof.SampleType))
	for _, sampleType := range prof.SampleType {
		sampleTypes = append(sampleTypes, sampleType.Type)
	}

	for _, mapping := range jfrProfileMappings {
		if mapping.metric == metric && sameStrings(mapping.sampleTypes, sampleTypes) {
			return mapping.name
		}
	}

	parts := append([]string{metric}, sampleTypes...)
	return sanitizeJFRProfileName(strings.Join(parts, "_"))
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sanitizeJFRProfileName(name string) string {
	var b strings.Builder
	lastUnderscore := false
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}
