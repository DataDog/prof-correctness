package analysis

import (
	"os"
	"strings"
	"testing"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.opentelemetry.io/collector/pdata/pprofile/pprofileotlp"
)

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
