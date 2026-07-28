// OpenTelemetry (OTLP) profiles adapter: maps OTLP profiles into the neutral
// ProfileSet directly, without a pprof round-trip. Trace/span linkage comes
// from the LinkTable and other attributes from the per-sample / resource
// attribute tables, so information a pprof intermediate cannot represent is
// preserved and normalized to canonical label keys.
package analysis

import (
	"path/filepath"
	"sort"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.opentelemetry.io/collector/pdata/pprofile/pprofileotlp"
)

// --- format detection & parsing --------------------------------------------

// Explicit OTLP-proto suffixes. Note ".pb" is intentionally absent: it is a
// common google/pprof suffix too, so plain ".pb" is left to the content-based
// fallback in LoadProfileSet rather than being forced to OTLP.
var otlpProtoSuffixes = []string{".otlp", ".otlppb", ".otlp.pb"}

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

// loadOTLP unmarshals an OTLP export request (proto or JSON) and converts it.
func loadOTLP(content []byte, asJSON bool) (*ProfileSet, error) {
	req := pprofileotlp.NewExportRequest()
	var err error
	if asJSON {
		err = req.UnmarshalJSON(content)
	} else {
		err = req.UnmarshalProto(content)
	}
	if err != nil {
		return nil, err
	}
	return FromOTLP(req.Profiles()), nil
}

// --- conversion ------------------------------------------------------------

// FromOTLP builds a ProfileSet from OTLP profiles. An export request may carry
// several profiles (per type, per resource/PID); each is added under its
// profile-type name.
func FromOTLP(profiles pprofile.Profiles) *ProfileSet {
	ps := newProfileSet()
	d := newOTLPDict(profiles.Dictionary())

	rps := profiles.ResourceProfiles()
	for i := 0; i < rps.Len(); i++ {
		rp := rps.At(i)
		resLabels := d.resourceLabels(rp.Resource().Attributes())

		sps := rp.ScopeProfiles()
		for j := 0; j < sps.Len(); j++ {
			pl := sps.At(j).Profiles()
			for k := 0; k < pl.Len(); k++ {
				op := pl.At(k)
				profileType := d.profileType(op)
				ps.setDuration(profileType, float64(op.DurationNano())/1e9)

				samples := op.Samples()
				for si := 0; si < samples.Len(); si++ {
					smp := samples.At(si)
					ps.add(profileType, StackSample{
						Stack:  d.foldStack(smp.StackIndex()),
						Val:    sampleValue(smp),
						Labels: d.sampleLabels(smp, resLabels),
					})
				}
			}
		}
	}
	return ps.finalize()
}

// sampleValue returns the sample's value for its (singular) profile type. All
// entries in the repeated Values field belong to that type, so they are summed;
// for sampling profiles that omit values the count is the number of timestamps
// (each timestamp is one occurrence).
func sampleValue(smp pprofile.Sample) int64 {
	vs := smp.Values()
	if vs.Len() == 0 {
		return int64(smp.TimestampsUnixNano().Len())
	}
	var sum int64
	for i := 0; i < vs.Len(); i++ {
		sum += vs.At(i)
	}
	return sum
}

// otlpDict resolves the index-based OTLP ProfilesDictionary into strings,
// folded stacks and canonical labels. All lookups are bounds-checked so a
// malformed dictionary yields empty results rather than a panic.
type otlpDict struct {
	strs   pcommon.StringSlice
	stacks pprofile.StackSlice
	locs   pprofile.LocationSlice
	funcs  pprofile.FunctionSlice
	maps   pprofile.MappingSlice
	links  pprofile.LinkSlice
	attrs  pprofile.KeyValueAndUnitSlice
}

func newOTLPDict(dict pprofile.ProfilesDictionary) *otlpDict {
	return &otlpDict{
		strs:   dict.StringTable(),
		stacks: dict.StackTable(),
		locs:   dict.LocationTable(),
		funcs:  dict.FunctionTable(),
		maps:   dict.MappingTable(),
		links:  dict.LinkTable(),
		attrs:  dict.AttributeTable(),
	}
}

func (d *otlpDict) str(i int32) string {
	if i < 0 || int(i) >= d.strs.Len() {
		return ""
	}
	return d.strs.At(int(i))
}

func (d *otlpDict) funcName(i int32) string {
	if i < 0 || int(i) >= d.funcs.Len() {
		return ""
	}
	return d.str(d.funcs.At(int(i)).NameStrindex())
}

func (d *otlpDict) mappingBase(i int32) string {
	if i < 0 || int(i) >= d.maps.Len() {
		return ""
	}
	return filepath.Base(d.str(d.maps.At(int(i)).FilenameStrindex()))
}

// profileType is the sample-type name, defaulting to "samples" when unset.
func (d *otlpDict) profileType(op pprofile.Profile) string {
	if t := d.str(op.SampleType().TypeStrindex()); t != "" {
		return t
	}
	return "samples"
}

// foldStack renders a dictionary stack as a root-first folded stack. Frames
// with no line info (unsymbolized native frames) are named after their mapping
// basename so binary/library-level assertions still match.
func (d *otlpDict) foldStack(stackIdx int32) string {
	if stackIdx < 0 || int(stackIdx) >= d.stacks.Len() {
		return ""
	}
	li := d.stacks.At(int(stackIdx)).LocationIndices()
	frames := make([]string, 0, li.Len()) // leaf-first, reversed below
	for x := 0; x < li.Len(); x++ {
		locIdx := li.At(x)
		if locIdx < 0 || int(locIdx) >= d.locs.Len() {
			continue
		}
		loc := d.locs.At(int(locIdx))
		if lines := loc.Lines(); lines.Len() > 0 {
			for y := 0; y < lines.Len(); y++ {
				frames = append(frames, d.funcName(lines.At(y).FunctionIndex()))
			}
		} else {
			frames = append(frames, d.mappingBase(loc.MappingIndex()))
		}
	}
	reverse(frames)
	return strings.Join(frames, ";")
}

// resourceLabels are the canonical labels shared by every sample under a
// resource (e.g. service.name).
func (d *otlpDict) resourceLabels(attrs pcommon.Map) map[string][]string {
	labels := map[string][]string{}
	attrs.Range(func(k string, v pcommon.Value) bool {
		labels[canonKey(k)] = append(labels[canonKey(k)], v.AsString())
		return true
	})
	return labels
}

// sampleLabels merges the resource labels with the sample's own attributes and
// its trace/span link, returning a fresh sorted map.
func (d *otlpDict) sampleLabels(smp pprofile.Sample, resLabels map[string][]string) map[string][]string {
	labels := map[string][]string{}
	for k, v := range resLabels {
		labels[k] = append([]string(nil), v...)
	}

	ai := smp.AttributeIndices()
	for a := 0; a < ai.Len(); a++ {
		d.addAttr(labels, ai.At(a))
	}
	d.addLink(labels, smp.LinkIndex())

	for k := range labels {
		sort.Strings(labels[k])
	}
	return labels
}

// addAttr adds one attribute-table entry (canonicalized) to labels, skipping
// out-of-range indices and empty keys/values.
func (d *otlpDict) addAttr(labels map[string][]string, idx int32) {
	if idx < 0 || int(idx) >= d.attrs.Len() {
		return
	}
	kv := d.attrs.At(int(idx))
	key := canonKey(d.str(kv.KeyStrindex()))
	val := kv.Value().AsString()
	if key == "" || val == "" {
		return
	}
	labels[key] = append(labels[key], val)
}

// addLink adds trace/span ids from the LinkTable (the OTLP-native encoding of
// span association) as canonical labels.
func (d *otlpDict) addLink(labels map[string][]string, linkIdx int32) {
	if linkIdx < 0 || int(linkIdx) >= d.links.Len() {
		return
	}
	link := d.links.At(int(linkIdx))
	if tid := link.TraceID(); !tid.IsEmpty() {
		labels[LabelTraceID] = append(labels[LabelTraceID], tid.String())
	}
	if sid := link.SpanID(); !sid.IsEmpty() {
		labels[LabelSpanID] = append(labels[LabelSpanID], sid.String())
	}
}

func reverse(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
