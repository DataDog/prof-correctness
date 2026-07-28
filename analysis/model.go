// Package analysis: neutral profile model and format dispatch.
//
// The analyzer asserts on a small, format-independent representation: a set of
// typed samples, each a folded stack (`a;b;c`), a value, and a canonical label
// set. This is what every assertion in analysis.go consumes.
//
// Rather than convert every wire format into google/pprof and lose whatever
// pprof cannot express, each input format has its own adapter (in its own file)
// that maps its native encoding into this neutral model:
//
//	pprof.go  FromPprof - google/pprof (labels come from Sample.Label / NumLabel)
//	otlp.go   FromOTLP  - OpenTelemetry profiles (trace/span from the LinkTable,
//	                      other attributes from the per-sample / resource
//	                      attribute tables) - no pprof round-trip
//	(future)  FromJFR   - Java Flight Recorder
//
// The point of the neutral model is that a single semantic expectation in
// expected_profile.json - e.g. label "span id" matches X - is verified
// identically regardless of how each format happens to encode span linkage.
// The Label* constants and canonKey() below are the seam where each format's
// key names are normalized to that shared vocabulary.
package analysis

import (
	"path/filepath"
	"sort"

	"github.com/google/pprof/profile"
)

// Canonical label keys. Adapters normalize each format's native key names to
// these so expected_profile.json can assert on one vocabulary across pprof,
// OTLP and (later) JFR. The values intentionally match the keys Datadog pprof
// profilers already emit, so existing scenarios keep working unchanged.
const (
	LabelTraceID      = "trace id"
	LabelSpanID       = "span id"
	LabelLocalRootSID = "local root span id"
	LabelThreadID     = "thread id"
	LabelThreadName   = "thread name"
	LabelProcessID    = "process_id"
	LabelService      = "service"
)

// canonKey maps a source format's attribute key to the canonical vocabulary.
// Unknown keys pass through unchanged. This is where JFR/OTLP/pprof key naming
// differences are reconciled.
func canonKey(k string) string {
	switch k {
	case "thread.id":
		return LabelThreadID
	case "thread.name":
		return LabelThreadName
	case "process.pid", "process.id":
		return LabelProcessID
	case "service.name":
		return LabelService
	case "trace.id", "trace_id":
		return LabelTraceID
	case "span.id", "span_id":
		return LabelSpanID
	case "local_root_span_id", "local.root.span.id":
		return LabelLocalRootSID
	default:
		return k
	}
}

// ProfileSet is the neutral, format-independent view of one profile file. A
// single file may contain several profile types (e.g. an OTLP export carrying
// alloc_space + alloc_objects, or a pprof profile with multiple sample types),
// each with its own duration.
type ProfileSet struct {
	order []string // sample-type names, in first-seen order
	typed map[string][]StackSample
	dur   map[string]*durAgg
}

// durAgg accumulates, per profile type, the total value and total rate
// (Σ valueᵢ/durationᵢ) across every profile of that type in the file. The
// effective duration is valueSum/rateSum, which makes total/duration equal the
// true aggregate rate even when same-type profiles (e.g. concurrent per-PID
// resources) have different durations. Only profiles with a positive duration
// contribute; snapshots (duration 0) leave the type's duration at 0.
type durAgg struct {
	valueSum int64
	rateSum  float64
}

func newProfileSet() *ProfileSet {
	return &ProfileSet{typed: map[string][]StackSample{}, dur: map[string]*durAgg{}}
}

func (ps *ProfileSet) add(profileType string, s StackSample) {
	if _, ok := ps.typed[profileType]; !ok {
		ps.order = append(ps.order, profileType)
	}
	ps.typed[profileType] = append(ps.typed[profileType], s)
}

// addProfileDuration folds one profile's (total value, duration) into the
// per-type duration aggregate. See durAgg. secs<=0 (a snapshot) is ignored.
func (ps *ProfileSet) addProfileDuration(profileType string, totalValue int64, secs float64) {
	if secs <= 0 {
		return
	}
	a := ps.dur[profileType]
	if a == nil {
		a = &durAgg{}
		ps.dur[profileType] = a
	}
	a.valueSum += totalValue
	a.rateSum += float64(totalValue) / secs
}

// Duration returns the effective duration in seconds for a profile type (0 if
// unknown or a snapshot). Kept per-type because one file can mix, e.g., a 10s
// allocation profile and a 60s CPU profile; for multiple same-type profiles it
// is valueSum/rateSum so total/Duration is the correct aggregate rate.
func (ps *ProfileSet) Duration(profileType string) float64 {
	a := ps.dur[profileType]
	if a == nil || a.rateSum == 0 {
		return 0
	}
	return float64(a.valueSum) / a.rateSum
}

// SampleTypes returns the profile-type names present, in first-seen order.
func (ps *ProfileSet) SampleTypes() []string { return ps.order }

// Samples returns the samples for a profile type, and whether that type exists.
func (ps *ProfileSet) Samples(profileType string) ([]StackSample, bool) {
	s, ok := ps.typed[profileType]
	return s, ok
}

// finalize sorts each type's samples by descending value for stable, readable
// capture output (assertions sum, so ordering is cosmetic there).
func (ps *ProfileSet) finalize() *ProfileSet {
	for _, s := range ps.typed {
		sort.SliceStable(s, func(i, j int) bool { return s[i].Val > s[j].Val })
	}
	return ps
}

// LoadProfileSet reads a profile file (pprof or OTLP) and returns the neutral
// ProfileSet. Format is chosen by filename suffix (.otlp/.otlp.pb -> OTLP
// proto, .otlp.json -> OTLP JSON, else pprof) with an OTLP fallback if pprof
// parsing fails. Ambiguous suffixes such as .pb (used by both pprof and OTLP)
// go through the content-based fallback rather than being forced to a format.
// The per-format parsing lives in the respective adapter file (pprof.go /
// otlp.go).
func LoadProfileSet(path string) (*ProfileSet, error) {
	content, err := readAndDecompress(path)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(path)
	switch {
	case isOTLPJSONName(name):
		return loadOTLP(content, true)
	case isOTLPProtoName(name):
		return loadOTLP(content, false)
	}

	// Unknown suffix: try pprof first, then OTLP (proto, then JSON).
	if prof, perr := profile.ParseData(content); perr == nil {
		return FromPprof(prof), nil
	}
	if ps, err := loadOTLP(content, false); err == nil {
		return ps, nil
	}
	if ps, err := loadOTLP(content, true); err == nil {
		return ps, nil
	}
	// Report the pprof error, which is the most informative for the common case.
	_, perr := profile.ParseData(content)
	return nil, perr
}
