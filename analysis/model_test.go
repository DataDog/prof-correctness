package analysis

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/google/pprof/profile"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.opentelemetry.io/collector/pdata/pprofile/pprofileotlp"
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

// buildOTLPWithLink constructs a minimal OTLP payload with one cpu-time sample
// whose stack leaf is symbolized "myHandler" in /app/server, that is linked to
// a trace/span via the LinkTable, carries a per-sample "thread.id" attribute,
// and inherits a resource-level "service.name". This exercises exactly the
// encodings a pprof round-trip cannot represent without re-encoding.
func buildOTLPWithLink(t *testing.T) []byte {
	t.Helper()
	profiles := pprofile.NewProfiles()
	dict := profiles.Dictionary()

	st := dict.StringTable()
	add := func(s string) int32 { st.Append(s); return int32(st.Len() - 1) }
	_ = add("") // 0
	cpuType := add("cpu-time")
	nanos := add("nanoseconds")
	fnName := add("myHandler")
	binFile := add("/app/server")
	threadKey := add("thread.id")

	fn := dict.FunctionTable().AppendEmpty()
	fn.SetNameStrindex(fnName)
	m := dict.MappingTable().AppendEmpty()
	m.SetFilenameStrindex(binFile)
	loc := dict.LocationTable().AppendEmpty()
	loc.SetMappingIndex(0)
	loc.Lines().AppendEmpty().SetFunctionIndex(0)
	stk := dict.StackTable().AppendEmpty()
	stk.LocationIndices().Append(0)

	// LinkTable entry with a concrete trace/span id.
	link := dict.LinkTable().AppendEmpty()
	link.SetTraceID(pcommon.TraceID{0x0a, 0x0b, 0x0c, 0x0d, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	link.SetSpanID(pcommon.SpanID{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88})

	// Per-sample attribute: thread.id = 4242.
	kv := dict.AttributeTable().AppendEmpty()
	kv.SetKeyStrindex(threadKey)
	kv.Value().SetInt(4242)

	rp := profiles.ResourceProfiles().AppendEmpty()
	rp.Resource().Attributes().PutStr("service.name", "checkout")
	sp := rp.ScopeProfiles().AppendEmpty()
	p := sp.Profiles().AppendEmpty()
	p.SampleType().SetTypeStrindex(cpuType)
	p.SampleType().SetUnitStrindex(nanos)
	s := p.Samples().AppendEmpty()
	s.SetStackIndex(0)
	s.SetLinkIndex(0)
	s.AttributeIndices().Append(0)
	s.Values().Append(1000)

	data, err := pprofileotlp.NewExportRequestFromProfiles(profiles).MarshalProto()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return data
}

// TestFromOTLP_TraceSpanAndAttributesSurvive is the crux of native OTLP
// support: trace/span linkage (LinkTable) and attributes (per-sample +
// resource) are mapped into canonical labels, so a semantic label assertion
// works on OTLP input. A pprof-as-hub converter drops these unless it
// re-encodes them, which is exactly what this design avoids.
func TestFromOTLP_TraceSpanAndAttributesSurvive(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/svc.otlp"
	if err := os.WriteFile(path, buildOTLPWithLink(t), 0o644); err != nil {
		t.Fatal(err)
	}

	ps, err := LoadProfileSet(path)
	if err != nil {
		t.Fatalf("LoadProfileSet: %v", err)
	}
	samples, ok := ps.Samples("cpu-time")
	if !ok || len(samples) != 1 {
		t.Fatalf("expected 1 cpu-time sample, got ok=%v n=%d", ok, len(samples))
	}
	s := samples[0]

	if !strings.Contains(s.Stack, "myHandler") {
		t.Errorf("stack %q missing symbolized frame myHandler", s.Stack)
	}
	if s.Val != 1000 {
		t.Errorf("value = %d, want 1000", s.Val)
	}

	// Canonical labels sourced from OTLP-specific encodings:
	checkLabel := func(key, want string) {
		vals, ok := s.Labels[key]
		if !ok || len(vals) == 0 {
			t.Errorf("label %q missing (labels=%v)", key, s.Labels)
			return
		}
		if vals[0] != want {
			t.Errorf("label %q = %q, want %q", key, vals[0], want)
		}
	}
	checkLabel(LabelTraceID, "0a0b0c0d000000000000000000000000") // from LinkTable
	checkLabel(LabelSpanID, "1122334455667788")                  // from LinkTable
	checkLabel(LabelThreadID, "4242")                            // per-sample attr, canonicalized from "thread.id"
	checkLabel(LabelService, "checkout")                         // resource attr, canonicalized from "service.name"
}

// TestFromOTLP_ExpectedJSONLabelAssertion drives the real assertion path: an
// expected_profile.json label expectation on span id is verified against OTLP
// input, end to end (the same JSON would work against pprof input carrying the
// same canonical label).
func TestFromOTLP_ExpectedJSONLabelAssertion(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/svc.otlp", buildOTLPWithLink(t), 0o644); err != nil {
		t.Fatal(err)
	}
	jsonPath := dir + "/expected_profile.json"
	if err := os.WriteFile(jsonPath, []byte(`{
      "test_name": "otlp_labels",
      "pprof-regex": ".*\\.otlp$",
      "scale_by_duration": false,
      "stacks": [{
        "profile-type": "cpu-time",
        "stack-content": [{
          "regular_expression": ".*myHandler.*",
          "value": 1000,
          "error_margin": 0,
          "labels": [
            { "key": "span id",  "values_regex": "^1122334455667788$" },
            { "key": "service",  "values": ["checkout"] }
          ]
        }]
      }]
    }`), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewStdReporter(os.Stdout, os.Stderr)
	Run(r, func() { AnalyzeResults(r, jsonPath, dir) })
	if r.Failed() {
		t.Fatalf("label assertion against OTLP input failed")
	}
}

// --- broader FromOTLP / LoadProfileSet coverage ----------------------------

// otlpBuilder is a tiny helper to assemble OTLP payloads with shared tables.
type otlpBuilder struct {
	t    *testing.T
	p    pprofile.Profiles
	strs map[string]int32
}

func newOTLPBuilder(t *testing.T) *otlpBuilder {
	b := &otlpBuilder{t: t, p: pprofile.NewProfiles(), strs: map[string]int32{}}
	b.str("") // index 0 must be empty
	return b
}
func (b *otlpBuilder) str(s string) int32 {
	if i, ok := b.strs[s]; ok {
		return i
	}
	st := b.p.Dictionary().StringTable()
	st.Append(s)
	i := int32(st.Len() - 1)
	b.strs[s] = i
	return i
}
func (b *otlpBuilder) fn(name string) int32 {
	f := b.p.Dictionary().FunctionTable().AppendEmpty()
	f.SetNameStrindex(b.str(name))
	return int32(b.p.Dictionary().FunctionTable().Len() - 1)
}
func (b *otlpBuilder) mapping(file string) int32 {
	m := b.p.Dictionary().MappingTable().AppendEmpty()
	m.SetFilenameStrindex(b.str(file))
	return int32(b.p.Dictionary().MappingTable().Len() - 1)
}
func (b *otlpBuilder) symLoc(fnIdx int32) int32 {
	l := b.p.Dictionary().LocationTable().AppendEmpty()
	l.Lines().AppendEmpty().SetFunctionIndex(fnIdx)
	return int32(b.p.Dictionary().LocationTable().Len() - 1)
}
func (b *otlpBuilder) unsymLoc(mapIdx int32) int32 {
	l := b.p.Dictionary().LocationTable().AppendEmpty()
	l.SetMappingIndex(mapIdx)
	return int32(b.p.Dictionary().LocationTable().Len() - 1)
}
func (b *otlpBuilder) stack(locsLeafFirst ...int32) int32 {
	s := b.p.Dictionary().StackTable().AppendEmpty()
	for _, l := range locsLeafFirst {
		s.LocationIndices().Append(l)
	}
	return int32(b.p.Dictionary().StackTable().Len() - 1)
}
func (b *otlpBuilder) marshal() []byte {
	data, err := pprofileotlp.NewExportRequestFromProfiles(b.p).MarshalProto()
	if err != nil {
		b.t.Fatalf("marshal: %v", err)
	}
	return data
}

// TestFromOTLP_StackFoldingAndValues covers multi-frame root-first ordering, an
// unsymbolized frame rendered as its mapping basename, explicit values, and the
// timestamp-count fallback across two merged profile types + duration adoption.
func TestFromOTLP_StackFoldingAndValues(t *testing.T) {
	b := newOTLPBuilder(t)
	root := b.symLoc(b.fn("root"))
	leaf := b.symLoc(b.fn("leaf"))
	lib := b.unsymLoc(b.mapping("/usr/lib/libfoo.so.1")) // unsymbolized

	// leaf-first: leaf, lib, root  -> folded root-first: root;libfoo.so.1;leaf
	stk := b.stack(leaf, lib, root)

	rp := b.p.ResourceProfiles().AppendEmpty()
	sp := rp.ScopeProfiles().AppendEmpty()

	// alloc_space with explicit value + a duration.
	p1 := sp.Profiles().AppendEmpty()
	p1.SampleType().SetTypeStrindex(b.str("alloc_space"))
	p1.SampleType().SetUnitStrindex(b.str("bytes"))
	p1.SetDurationNano(7_000_000_000)
	s1 := p1.Samples().AppendEmpty()
	s1.SetStackIndex(stk)
	s1.Values().Append(500)

	// samples profile with NO values but 3 timestamps -> count 3.
	p2 := sp.Profiles().AppendEmpty()
	p2.SampleType().SetTypeStrindex(b.str("samples"))
	p2.SampleType().SetUnitStrindex(b.str("count"))
	s2 := p2.Samples().AppendEmpty()
	s2.SetStackIndex(stk)
	s2.TimestampsUnixNano().Append(1, 2, 3)

	path := t.TempDir() + "/x.otlp"
	if err := os.WriteFile(path, b.marshal(), 0o644); err != nil {
		t.Fatal(err)
	}
	ps, err := LoadProfileSet(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if ps.DurationSecs != 7 {
		t.Errorf("duration = %v, want 7", ps.DurationSecs)
	}
	alloc, _ := ps.Samples("alloc_space")
	if len(alloc) != 1 || alloc[0].Val != 500 {
		t.Fatalf("alloc_space = %+v", alloc)
	}
	if alloc[0].Stack != "root;libfoo.so.1;leaf" {
		t.Errorf("folded stack = %q, want root;libfoo.so.1;leaf", alloc[0].Stack)
	}
	samp, _ := ps.Samples("samples")
	if len(samp) != 1 || samp[0].Val != 3 {
		t.Errorf("samples (timestamp count) = %+v, want val 3", samp)
	}
}

// TestFromOTLP_EmptyTypeDefaultsToSamples covers the empty sample-type default.
func TestFromOTLP_EmptyTypeDefaultsToSamples(t *testing.T) {
	b := newOTLPBuilder(t)
	stk := b.stack(b.symLoc(b.fn("f")))
	rp := b.p.ResourceProfiles().AppendEmpty()
	p := rp.ScopeProfiles().AppendEmpty().Profiles().AppendEmpty()
	// leave SampleType type strindex at 0 (empty string)
	p.Samples().AppendEmpty().SetStackIndex(stk)

	ps := FromOTLP(b.p)
	if _, ok := ps.Samples("samples"); !ok {
		t.Fatalf("empty sample type should default to 'samples', got types=%v", ps.SampleTypes())
	}
}

// TestFromOTLP_IgnoresBadRefsAndEmptyLabels covers the defensive guards: an
// out-of-range attribute index, an empty-valued attribute, a zero (unset) link,
// and an out-of-range stack index must not panic or create bogus labels.
func TestFromOTLP_IgnoresBadRefsAndEmptyLabels(t *testing.T) {
	b := newOTLPBuilder(t)
	stk := b.stack(b.symLoc(b.fn("f")))

	// attribute with an empty string value -> skipped.
	kv := b.p.Dictionary().AttributeTable().AppendEmpty()
	kv.SetKeyStrindex(b.str("empty.attr"))
	kv.Value().SetStr("")
	// zero/unset link -> skipped.
	b.p.Dictionary().LinkTable().AppendEmpty()

	rp := b.p.ResourceProfiles().AppendEmpty()
	p := rp.ScopeProfiles().AppendEmpty().Profiles().AppendEmpty()
	p.SampleType().SetTypeStrindex(b.str("samples"))
	s := p.Samples().AppendEmpty()
	s.SetStackIndex(stk)
	s.AttributeIndices().Append(0)   // empty value
	s.AttributeIndices().Append(999) // out of range
	s.SetLinkIndex(0)                // zero link
	s.Values().Append(1)

	// A second sample with an out-of-range stack index -> empty stack, no panic.
	s2 := p.Samples().AppendEmpty()
	s2.SetStackIndex(999)
	s2.Values().Append(1)

	ps := FromOTLP(b.p)
	samples, _ := ps.Samples("samples")
	if len(samples) != 2 {
		t.Fatalf("expected 2 samples, got %d", len(samples))
	}
	for _, ss := range samples {
		if len(ss.Labels) != 0 {
			t.Errorf("expected no labels from bad/empty refs, got %v", ss.Labels)
		}
	}
}

// TestFromOTLP_MultipleResources covers per-resource attribute scoping.
func TestFromOTLP_MultipleResources(t *testing.T) {
	b := newOTLPBuilder(t)
	stk := b.stack(b.symLoc(b.fn("f")))
	for _, svc := range []string{"svc-a", "svc-b"} {
		rp := b.p.ResourceProfiles().AppendEmpty()
		rp.Resource().Attributes().PutStr("service.name", svc)
		p := rp.ScopeProfiles().AppendEmpty().Profiles().AppendEmpty()
		p.SampleType().SetTypeStrindex(b.str("samples"))
		s := p.Samples().AppendEmpty()
		s.SetStackIndex(stk)
		s.Values().Append(1)
	}
	ps := FromOTLP(b.p)
	samples, _ := ps.Samples("samples")
	seen := map[string]bool{}
	for _, ss := range samples {
		for _, v := range ss.Labels[LabelService] {
			seen[v] = true
		}
	}
	if !seen["svc-a"] || !seen["svc-b"] {
		t.Errorf("expected both services in per-resource labels, got %v", seen)
	}
}

// TestLoadProfileSet_JSONAndPprof covers the OTLP-JSON suffix and the pprof
// path (and thus FromPprof) through the same loader.
func TestLoadProfileSet_JSONAndPprof(t *testing.T) {
	dir := t.TempDir()

	// OTLP JSON.
	b := newOTLPBuilder(t)
	stk := b.stack(b.symLoc(b.fn("jsonFn")))
	rp := b.p.ResourceProfiles().AppendEmpty()
	p := rp.ScopeProfiles().AppendEmpty().Profiles().AppendEmpty()
	p.SampleType().SetTypeStrindex(b.str("samples"))
	s := p.Samples().AppendEmpty()
	s.SetStackIndex(stk)
	s.Values().Append(5)
	jsonData, err := pprofileotlp.NewExportRequestFromProfiles(b.p).MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	jp := dir + "/a.otlp.json"
	if err := os.WriteFile(jp, jsonData, 0o644); err != nil {
		t.Fatal(err)
	}
	ps, err := LoadProfileSet(jp)
	if err != nil {
		t.Fatalf("load json: %v", err)
	}
	if samp, _ := ps.Samples("samples"); len(samp) != 1 || !strings.Contains(samp[0].Stack, "jsonFn") {
		t.Errorf("OTLP JSON not parsed as expected: %+v", samp)
	}

	// pprof via the same loader.
	pp := dir + "/b.pprof"
	if err := os.WriteFile(pp, buildPprof(t), 0o644); err != nil {
		t.Fatal(err)
	}
	ps2, err := LoadProfileSet(pp)
	if err != nil {
		t.Fatalf("load pprof: %v", err)
	}
	cpu, ok := ps2.Samples("cpu")
	if !ok || len(cpu) != 1 || cpu[0].Val != 42 || !strings.Contains(cpu[0].Stack, "pprofFn") {
		t.Errorf("pprof via loader not parsed as expected: %+v", cpu)
	}
}

// TestCanonKey pins the cross-format label vocabulary: each format's native key
// names must normalize to the same canonical key, and unknown keys pass through.
func TestCanonKey(t *testing.T) {
	cases := map[string]string{
		"thread.id":          LabelThreadID,
		"thread.name":        LabelThreadName,
		"process.pid":        LabelProcessID,
		"process.id":         LabelProcessID,
		"service.name":       LabelService,
		"trace.id":           LabelTraceID,
		"trace_id":           LabelTraceID,
		"span.id":            LabelSpanID,
		"span_id":            LabelSpanID,
		"local_root_span_id": LabelLocalRootSID,
		"local.root.span.id": LabelLocalRootSID,
		"some.custom.key":    "some.custom.key", // passthrough
		"span id":            "span id",         // already canonical
	}
	for in, want := range cases {
		if got := canonKey(in); got != want {
			t.Errorf("canonKey(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestLoadProfileSet_FallbackAndError covers a suffixless file whose bytes are
// OTLP proto (parsed via the fallback) and an unparseable file (error).
func TestLoadProfileSet_FallbackAndError(t *testing.T) {
	dir := t.TempDir()

	// OTLP proto bytes under a non-OTLP suffix -> parsed via fallback.
	b := newOTLPBuilder(t)
	stk := b.stack(b.symLoc(b.fn("fallbackFn")))
	rp := b.p.ResourceProfiles().AppendEmpty()
	p := rp.ScopeProfiles().AppendEmpty().Profiles().AppendEmpty()
	p.SampleType().SetTypeStrindex(b.str("samples"))
	s := p.Samples().AppendEmpty()
	s.SetStackIndex(stk)
	s.Values().Append(1)
	fb := dir + "/mystery.bin"
	if err := os.WriteFile(fb, b.marshal(), 0o644); err != nil {
		t.Fatal(err)
	}
	if ps, err := LoadProfileSet(fb); err != nil {
		t.Errorf("fallback parse failed: %v", err)
	} else if samp, _ := ps.Samples("samples"); len(samp) != 1 {
		t.Errorf("fallback: expected 1 sample, got %d", len(samp))
	}

	// Unparseable bytes -> error (not a panic).
	bad := dir + "/garbage.bin"
	if err := os.WriteFile(bad, []byte("this is not a profile at all"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProfileSet(bad); err == nil {
		t.Errorf("expected error for unparseable input, got nil")
	}
}
