package analysis

import (
	"os"
	"strings"
	"testing"

	"go.opentelemetry.io/collector/pdata/pprofile/pprofileotlp"
)

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

// TestLoadProfileSet_JSONAndPprof covers the format dispatch: the OTLP-JSON
// suffix (-> FromOTLP) and the pprof path (-> FromPprof) through one loader.
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

// TestCaptureJSONPath verifies the capture-output path never collides with the
// input: normal profiles get <base>.json, but an input already ending in .json
// (e.g. *.otlp.json) gets a .capture.json variant so its source is preserved.
func TestCaptureJSONPath(t *testing.T) {
	cases := map[string]string{
		"/d/x.pprof":     "/d/x.json",
		"/d/x.otlp":      "/d/x.json",
		"/d/x.otlp.json": "/d/x.otlp.capture.json", // must not equal the input
	}
	for in, want := range cases {
		if got := captureJSONPath(in); got != want {
			t.Errorf("captureJSONPath(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestCaptureDoesNotClobberOTLPJSON drives the analyzer with capture enabled on
// an .otlp.json input and confirms the source payload is still readable
// afterwards (previously the capture JSON overwrote it).
func TestCaptureDoesNotClobberOTLPJSON(t *testing.T) {
	dir := t.TempDir()
	b := newOTLPBuilder(t)
	stk := b.stack(b.symLoc(b.fn("capFn")))
	rp := b.p.ResourceProfiles().AppendEmpty()
	p := rp.ScopeProfiles().AppendEmpty().Profiles().AppendEmpty()
	p.SampleType().SetTypeStrindex(b.str("samples"))
	s := p.Samples().AppendEmpty()
	s.SetStackIndex(stk)
	s.Values().Append(1)
	jsonData, err := pprofileotlp.NewExportRequestFromProfiles(b.p).MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	src := dir + "/x.otlp.json"
	if err := os.WriteFile(src, jsonData, 0o644); err != nil {
		t.Fatal(err)
	}
	jsonPath := dir + "/expected_profile.json"
	if err := os.WriteFile(jsonPath, []byte(`{
      "test_name": "cap",
      "pprof-regex": ".*x\\.otlp\\.json$",
      "scale_by_duration": false,
      "stacks": [{ "profile-type": "samples",
        "stack-content": [{ "regular_expression": ".*capFn.*", "percent": 100, "error_margin": 5 }] }]
    }`), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewStdReporter(os.Stdout, os.Stderr)
	Run(r, func() { AnalyzeResults(r, jsonPath, dir) })

	// The source .otlp.json must still be a valid OTLP payload (not clobbered).
	ps, err := LoadProfileSet(src)
	if err != nil {
		t.Fatalf("source .otlp.json was corrupted by capture: %v", err)
	}
	if samp, ok := ps.Samples("samples"); !ok || len(samp) != 1 {
		t.Errorf("source no longer parses to its sample: ok=%v samp=%+v", ok, samp)
	}
}

// TestLoadProfileSet_PbUsesContentFallback covers that a plain .pb file is
// content-sniffed (pprof here) rather than forced to OTLP, while .otlp.pb stays
// explicitly OTLP.
func TestLoadProfileSet_PbUsesContentFallback(t *testing.T) {
	dir := t.TempDir()

	// A google/pprof profile written with the common .pb suffix.
	pb := dir + "/prof.pb"
	if err := os.WriteFile(pb, buildPprof(t), 0o644); err != nil {
		t.Fatal(err)
	}
	ps, err := LoadProfileSet(pb)
	if err != nil {
		t.Fatalf(".pb pprof not parsed via fallback: %v", err)
	}
	if cpu, ok := ps.Samples("cpu"); !ok || len(cpu) != 1 || cpu[0].Val != 42 {
		t.Errorf(".pb should parse as pprof, got %+v", cpu)
	}

	// .otlp.pb remains explicitly OTLP.
	b := newOTLPBuilder(t)
	stk := b.stack(b.symLoc(b.fn("pbFn")))
	rp := b.p.ResourceProfiles().AppendEmpty()
	p := rp.ScopeProfiles().AppendEmpty().Profiles().AppendEmpty()
	p.SampleType().SetTypeStrindex(b.str("samples"))
	sm := p.Samples().AppendEmpty()
	sm.SetStackIndex(stk)
	sm.Values().Append(1)
	op := dir + "/x.otlp.pb"
	if err := os.WriteFile(op, b.marshal(), 0o644); err != nil {
		t.Fatal(err)
	}
	ops, err := LoadProfileSet(op)
	if err != nil {
		t.Fatalf(".otlp.pb not parsed: %v", err)
	}
	if samp, ok := ops.Samples("samples"); !ok || len(samp) != 1 || !strings.Contains(samp[0].Stack, "pbFn") {
		t.Errorf(".otlp.pb should parse as OTLP, got %+v", samp)
	}
}
