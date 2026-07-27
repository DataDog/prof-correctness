// OTLP (OpenTelemetry) profiles input support.
//
// NOTE: this is a converter, not a native OTLP assertion path. The analyzer's
// internal model is github.com/google/pprof; OTLP payloads are read here and
// converted into that pprof model so the existing assertion, capture and
// rate-scaling logic is reused unchanged. Consequently anything OTLP carries
// that does not map onto pprof (resource/sample attribute tables, span/trace
// links, per-sample timestamps beyond their count) is only preserved insofar
// as this converter translates it -- attribute->label mapping is TODO, so
// label-based assertions currently work for pprof inputs only.
//
// The OpenTelemetry profiles signal is the wire format used by the eBPF
// full-host / datadog-agent host profiler, so accepting it lets scenarios feed
// that profiler's output straight to the analyzer (via a dumb OTLP file dump)
// without every scenario reimplementing a conversion.
//
// One OTLP export request can carry many Profile messages (e.g. the eBPF
// profiler emits alloc_space, alloc_objects, inuse_space, inuse_objects and
// cpu "samples" as separate profiles, one per resource/PID). We merge them
// into a single pprof profile whose SampleType list is the union of the
// per-profile types; each sample carries its value in its type's column and
// zero elsewhere. getProfileType() then selects a column by type name exactly
// as it does for a multi-sample-type pprof profile.
package analysis

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/pprof/profile"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.opentelemetry.io/collector/pdata/pprofile/pprofileotlp"
)

// otlpFileSuffixes are filename suffixes that unambiguously denote an OTLP
// profiles payload. Proto is the default; ".json" variants are parsed as OTLP
// JSON. Files without these suffixes are treated as pprof (with an OTLP
// fallback attempted if pprof parsing fails).
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

// parseOTLP unmarshals an OTLP ExportProfilesServiceRequest (proto or JSON)
// and converts it to a merged pprof profile.
func parseOTLP(content []byte, asJSON bool) (*profile.Profile, error) {
	req := pprofileotlp.NewExportRequest()
	var err error
	if asJSON {
		err = req.UnmarshalJSON(content)
	} else {
		err = req.UnmarshalProto(content)
	}
	if err != nil {
		return nil, fmt.Errorf("otlp unmarshal: %w", err)
	}
	return otlpProfilesToPprof(req.Profiles())
}

// otlpProfilesToPprof converts an OTLP profiles payload into a single merged
// google/pprof profile.
func otlpProfilesToPprof(profiles pprofile.Profiles) (*profile.Profile, error) {
	dict := profiles.Dictionary()
	strs := dict.StringTable()
	getStr := func(idx int32) string {
		if idx < 0 || int(idx) >= strs.Len() {
			return ""
		}
		return strs.At(int(idx))
	}

	out := &profile.Profile{}

	// Union of sample types across all profiles, in first-seen order.
	type stKey struct{ typ, unit string }
	colOf := map[stKey]int{}
	addType := func(typ, unit string) int {
		if typ == "" {
			typ = "samples"
		}
		if unit == "" {
			unit = "count"
		}
		k := stKey{typ, unit}
		if c, ok := colOf[k]; ok {
			return c
		}
		c := len(out.SampleType)
		colOf[k] = c
		out.SampleType = append(out.SampleType, &profile.ValueType{Type: typ, Unit: unit})
		return c
	}

	// Shared dedup caches keyed by dictionary index (the dictionary is shared
	// across the whole request).
	funcs := map[int32]*profile.Function{}
	locs := map[int32]*profile.Location{}
	maps := map[int32]*profile.Mapping{}
	var funcID, locID, mapID, synthID uint64

	mappingTable := dict.MappingTable()
	locationTable := dict.LocationTable()
	functionTable := dict.FunctionTable()
	stackTable := dict.StackTable()

	var getLocation func(idx int32) *profile.Location
	getMapping := func(idx int32) *profile.Mapping {
		if idx < 0 || int(idx) >= mappingTable.Len() {
			return nil
		}
		if m, ok := maps[idx]; ok {
			return m
		}
		om := mappingTable.At(int(idx))
		mapID++
		m := &profile.Mapping{
			ID:     mapID,
			Start:  om.MemoryStart(),
			Limit:  om.MemoryLimit(),
			Offset: om.FileOffset(),
			File:   getStr(om.FilenameStrindex()),
		}
		out.Mapping = append(out.Mapping, m)
		maps[idx] = m
		return m
	}
	getFunction := func(idx int32) *profile.Function {
		if idx < 0 || int(idx) >= functionTable.Len() {
			return nil
		}
		if f, ok := funcs[idx]; ok {
			return f
		}
		of := functionTable.At(int(idx))
		funcID++
		f := &profile.Function{
			ID:         funcID,
			Name:       getStr(of.NameStrindex()),
			SystemName: getStr(of.SystemNameStrindex()),
			Filename:   getStr(of.FilenameStrindex()),
			StartLine:  of.StartLine(),
		}
		out.Function = append(out.Function, f)
		funcs[idx] = f
		return f
	}
	getLocation = func(idx int32) *profile.Location {
		if idx < 0 || int(idx) >= locationTable.Len() {
			return nil
		}
		if l, ok := locs[idx]; ok {
			return l
		}
		ol := locationTable.At(int(idx))
		locID++
		l := &profile.Location{
			ID:      locID,
			Address: ol.Address(),
			Mapping: getMapping(ol.MappingIndex()),
		}
		lines := ol.Lines()
		for li := 0; li < lines.Len(); li++ {
			ln := lines.At(li)
			if fn := getFunction(ln.FunctionIndex()); fn != nil {
				l.Line = append(l.Line, profile.Line{Function: fn, Line: ln.Line(), Column: ln.Column()})
			}
		}
		// Unsymbolized native frames carry only a Mapping (no Line/Function).
		// The analyzer folds stacks purely on Function.Name, so synthesize a
		// frame named after the mapping's basename, matching how pprof tooling
		// renders unsymbolized frames and letting assertions match the
		// originating binary/library.
		if len(l.Line) == 0 && l.Mapping != nil && l.Mapping.File != "" {
			synthID++
			fn := &profile.Function{
				ID:         1_000_000_000 + synthID,
				Name:       filepath.Base(l.Mapping.File),
				SystemName: l.Mapping.File,
				Filename:   l.Mapping.File,
			}
			out.Function = append(out.Function, fn)
			l.Line = append(l.Line, profile.Line{Function: fn})
		}
		out.Location = append(out.Location, l)
		locs[idx] = l
		return l
	}

	rps := profiles.ResourceProfiles()
	for i := 0; i < rps.Len(); i++ {
		sps := rps.At(i).ScopeProfiles()
		for j := 0; j < sps.Len(); j++ {
			ps := sps.At(j).Profiles()
			for k := 0; k < ps.Len(); k++ {
				op := ps.At(k)
				col := addType(getStr(op.SampleType().TypeStrindex()), getStr(op.SampleType().UnitStrindex()))

				// Period / duration: adopt the first meaningful values seen.
				if out.PeriodType == nil {
					if pt := op.PeriodType(); getStr(pt.TypeStrindex()) != "" {
						out.PeriodType = &profile.ValueType{Type: getStr(pt.TypeStrindex()), Unit: getStr(pt.UnitStrindex())}
						out.Period = op.Period()
					}
				}
				if d := int64(op.DurationNano()); d > out.DurationNanos {
					out.DurationNanos = d
				}
				if out.TimeNanos == 0 {
					out.TimeNanos = int64(op.Time())
				}

				samples := op.Samples()
				for si := 0; si < samples.Len(); si++ {
					smp := samples.At(si)
					stackIdx := smp.StackIndex()
					if stackIdx < 0 || int(stackIdx) >= stackTable.Len() {
						continue
					}
					locIdx := stackTable.At(int(stackIdx)).LocationIndices()
					var stack []*profile.Location
					for li := 0; li < locIdx.Len(); li++ {
						if l := getLocation(locIdx.At(li)); l != nil {
							stack = append(stack, l)
						}
					}
					// The OTLP profiles schema allows a sample to omit an explicit
					// value: the count is then the number of timestamps. The eBPF
					// profiler uses this for the CPU "samples" profile (one
					// timestamp per observed on-CPU sample) while heap/off-CPU
					// profiles carry explicit values. Honor both.
					var v int64
					if vs := smp.Values(); vs.Len() > 0 {
						v = vs.At(0)
					} else {
						v = int64(smp.TimestampsUnixNano().Len())
					}
					val := make([]int64, col+1)
					val[col] = v
					out.Sample = append(out.Sample, &profile.Sample{Location: stack, Value: val})
				}
			}
		}
	}

	if len(out.SampleType) == 0 {
		out.SampleType = []*profile.ValueType{{Type: "samples", Unit: "count"}}
	}
	// Pad every sample's value vector to the full union width (zeros for the
	// types it doesn't belong to).
	width := len(out.SampleType)
	for _, s := range out.Sample {
		if len(s.Value) < width {
			padded := make([]int64, width)
			copy(padded, s.Value)
			s.Value = padded
		}
	}
	return out, nil
}
