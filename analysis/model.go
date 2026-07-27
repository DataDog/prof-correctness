// Neutral profile model and per-format adapters.
//
// The analyzer asserts on a small, format-independent representation: a set of
// typed samples, each a folded stack (`a;b;c`), a value, and a canonical label
// set. This is what every assertion in analysis.go consumes.
//
// Rather than convert every wire format into google/pprof and lose whatever
// pprof cannot express, each input format has its own adapter that maps its
// native encoding into this neutral model:
//
//	FromPprof  - google/pprof (labels come from Sample.Label / NumLabel)
//	FromOTLP   - OpenTelemetry profiles (trace/span come from the LinkTable,
//	             other attributes from the per-sample / resource attribute
//	             tables) - no pprof round-trip
//	FromJFR    - (future) Java Flight Recorder
//
// The point of the neutral model is that a single semantic expectation in
// expected_profile.json - e.g. label "span id" matches X - is verified
// identically regardless of how each format happens to encode span linkage.
// The canonical* constants and canonKey() below are the seam where each
// format's key names are normalized to that shared vocabulary.
package analysis

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/google/pprof/profile"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.opentelemetry.io/collector/pdata/pprofile/pprofileotlp"
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
// alloc_space + alloc_objects, or a pprof profile with multiple sample types).
type ProfileSet struct {
	DurationSecs float64
	order        []string // sample-type names, in first-seen order
	typed        map[string][]StackSample
}

func newProfileSet() *ProfileSet {
	return &ProfileSet{typed: map[string][]StackSample{}}
}

func (ps *ProfileSet) add(profileType string, s StackSample) {
	if _, ok := ps.typed[profileType]; !ok {
		ps.order = append(ps.order, profileType)
	}
	ps.typed[profileType] = append(ps.typed[profileType], s)
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

// FromPprof builds a ProfileSet from a google/pprof profile. pprof carries its
// labels in Sample.Label / NumLabel, which are copied verbatim (Datadog pprof
// profilers already use the canonical key names).
func FromPprof(prof *profile.Profile) *ProfileSet {
	ps := newProfileSet()
	ps.DurationSecs = float64(prof.DurationNanos) / 1e9

	// Merge identical locations/samples so counts are consistent.
	_ = prof.Aggregate(true, true, false, false, false, false)
	prof = prof.Compact()

	for _, sample := range prof.Sample {
		stack := foldPprofStack(sample)
		labels := map[string][]string{}
		for k, v := range sample.Label {
			cp := append([]string(nil), v...)
			sort.Strings(cp)
			labels[canonKey(k)] = cp
		}
		for k, v := range sample.NumLabel {
			var vals []string
			for _, n := range v {
				vals = append(vals, strconv.FormatInt(n, 10))
			}
			sort.Strings(vals)
			labels[canonKey(k)] = vals
		}
		for i, st := range prof.SampleType {
			ps.add(st.Type, StackSample{Stack: stack, Val: sample.Value[i], Labels: labels})
		}
	}
	return ps.finalize()
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

// FromOTLP builds a ProfileSet directly from OTLP profiles - no pprof
// intermediate. Trace/span linkage is read from the LinkTable and per-sample /
// resource attributes from the attribute table, then normalized to canonical
// label keys. This is the fidelity that a pprof round-trip would drop.
func FromOTLP(profiles pprofile.Profiles) *ProfileSet {
	ps := newProfileSet()
	dict := profiles.Dictionary()
	strs := dict.StringTable()
	getStr := func(i int32) string {
		if i < 0 || int(i) >= strs.Len() {
			return ""
		}
		return strs.At(int(i))
	}

	stackTable := dict.StackTable()
	locTable := dict.LocationTable()
	funcTable := dict.FunctionTable()
	mapTable := dict.MappingTable()
	linkTable := dict.LinkTable()
	attrTable := dict.AttributeTable()

	fnName := func(i int32) string {
		if i < 0 || int(i) >= funcTable.Len() {
			return ""
		}
		return getStr(funcTable.At(int(i)).NameStrindex())
	}
	mapBase := func(i int32) string {
		if i < 0 || int(i) >= mapTable.Len() {
			return ""
		}
		return filepath.Base(getStr(mapTable.At(int(i)).FilenameStrindex()))
	}

	// foldStack renders a dictionary stack as a root-first folded stack.
	foldStack := func(stackIdx int32) string {
		if stackIdx < 0 || int(stackIdx) >= stackTable.Len() {
			return ""
		}
		li := stackTable.At(int(stackIdx)).LocationIndices()
		// li is leaf-first; build leaf-first then reverse to root-first.
		var leafFirst []string
		for x := 0; x < li.Len(); x++ {
			locIdx := li.At(x)
			if locIdx < 0 || int(locIdx) >= locTable.Len() {
				continue
			}
			loc := locTable.At(int(locIdx))
			lines := loc.Lines()
			if lines.Len() == 0 {
				// Unsymbolized native frame: name it after the mapping
				// basename so binary/library-level assertions still match.
				leafFirst = append(leafFirst, mapBase(loc.MappingIndex()))
				continue
			}
			for y := 0; y < lines.Len(); y++ {
				leafFirst = append(leafFirst, fnName(lines.At(y).FunctionIndex()))
			}
		}
		for i, j := 0, len(leafFirst)-1; i < j; i, j = i+1, j-1 {
			leafFirst[i], leafFirst[j] = leafFirst[j], leafFirst[i]
		}
		return strings.Join(leafFirst, ";")
	}

	addAttr := func(labels map[string][]string, idx int32) {
		if idx < 0 || int(idx) >= attrTable.Len() {
			return
		}
		kv := attrTable.At(int(idx))
		key := canonKey(getStr(kv.KeyStrindex()))
		val := kv.Value().AsString()
		if key == "" || val == "" {
			return
		}
		labels[key] = append(labels[key], val)
	}

	rps := profiles.ResourceProfiles()
	for i := 0; i < rps.Len(); i++ {
		rp := rps.At(i)

		// Resource-level attributes (e.g. service.name) apply to every sample
		// under this resource.
		resLabels := map[string][]string{}
		rp.Resource().Attributes().Range(func(k string, v pcommon.Value) bool {
			resLabels[canonKey(k)] = append(resLabels[canonKey(k)], v.AsString())
			return true
		})

		sps := rp.ScopeProfiles()
		for j := 0; j < sps.Len(); j++ {
			pl := sps.At(j).Profiles()
			for k := 0; k < pl.Len(); k++ {
				op := pl.At(k)
				if d := float64(op.DurationNano()) / 1e9; d > ps.DurationSecs {
					ps.DurationSecs = d
				}
				profileType := getStr(op.SampleType().TypeStrindex())
				if profileType == "" {
					profileType = "samples"
				}
				samples := op.Samples()
				for si := 0; si < samples.Len(); si++ {
					smp := samples.At(si)

					labels := map[string][]string{}
					for ck, cv := range resLabels {
						labels[ck] = append([]string(nil), cv...)
					}
					ai := smp.AttributeIndices()
					for a := 0; a < ai.Len(); a++ {
						addAttr(labels, ai.At(a))
					}
					// Trace/span linkage lives in the LinkTable, not in labels.
					if li := smp.LinkIndex(); li >= 0 && int(li) < linkTable.Len() {
						link := linkTable.At(int(li))
						if tid := link.TraceID(); !tid.IsEmpty() {
							labels[LabelTraceID] = append(labels[LabelTraceID], tid.String())
						}
						if sid := link.SpanID(); !sid.IsEmpty() {
							labels[LabelSpanID] = append(labels[LabelSpanID], sid.String())
						}
					}
					for lk := range labels {
						sort.Strings(labels[lk])
					}

					// Value: explicit if present, else the number of
					// timestamps (sampling profiles encode count that way).
					var val int64
					if vs := smp.Values(); vs.Len() > 0 {
						val = vs.At(0)
					} else {
						val = int64(smp.TimestampsUnixNano().Len())
					}

					ps.add(profileType, StackSample{
						Stack:  foldStack(smp.StackIndex()),
						Val:    val,
						Labels: labels,
					})
				}
			}
		}
	}
	return ps.finalize()
}

// --- format detection / loading -------------------------------------------

var otlpProtoSuffixes = []string{".otlp", ".otlppb", ".otlp.pb", ".pb"}

func isOTLPProtoName(name string) bool {
	n := strings.ToLower(name)
	for _, s := range otlpProtoSuffixes {
		if strings.HasSuffix(n, s) {
			return true
		}
	}
	return false
}

func isOTLPJSONName(name string) bool {
	n := strings.ToLower(name)
	return strings.HasSuffix(n, ".otlp.json") || strings.HasSuffix(n, ".otlpjson")
}

func parseOTLPProfiles(content []byte, asJSON bool) (pprofile.Profiles, error) {
	req := pprofileotlp.NewExportRequest()
	var err error
	if asJSON {
		err = req.UnmarshalJSON(content)
	} else {
		err = req.UnmarshalProto(content)
	}
	if err != nil {
		return pprofile.NewProfiles(), err
	}
	return req.Profiles(), nil
}

// LoadProfileSet reads a profile file (pprof or OTLP) and returns the neutral
// ProfileSet. Format is chosen by filename suffix (.otlp/.pb -> OTLP proto,
// .otlp.json -> OTLP JSON, else pprof) with an OTLP fallback if pprof parsing
// fails, so callers need not know the format up front.
func LoadProfileSet(path string) (*ProfileSet, error) {
	content, err := readAndDecompress(path)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(path)
	switch {
	case isOTLPJSONName(name):
		p, err := parseOTLPProfiles(content, true)
		if err != nil {
			return nil, err
		}
		return FromOTLP(p), nil
	case isOTLPProtoName(name):
		p, err := parseOTLPProfiles(content, false)
		if err != nil {
			return nil, err
		}
		return FromOTLP(p), nil
	}
	if prof, perr := profile.ParseData(content); perr == nil {
		return FromPprof(prof), nil
	} else if p, oerr := parseOTLPProfiles(content, false); oerr == nil {
		return FromOTLP(p), nil
	} else if p, jerr := parseOTLPProfiles(content, true); jerr == nil {
		return FromOTLP(p), nil
	} else {
		return nil, perr
	}
}
