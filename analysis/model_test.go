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
