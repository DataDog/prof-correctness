package analysis

import (
	"os"
	"strings"
	"testing"

	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.opentelemetry.io/collector/pdata/pprofile/pprofileotlp"
)

// buildOTLPRequest constructs a minimal OTLP profiles payload with two paired
// heap profiles (alloc_space bytes + alloc_objects count) sharing one stack
// whose leaf is symbolized as "allocate_chunk" in mapping /app/heap_workload,
// plus a second, unsymbolized location (mapping only) to exercise the
// synthetic-frame path.
func buildOTLPRequest(t *testing.T) []byte {
	t.Helper()
	profiles := pprofile.NewProfiles()
	dict := profiles.Dictionary()

	// String table (index 0 must be empty per OTLP convention).
	st := dict.StringTable()
	add := func(s string) int32 {
		st.Append(s)
		return int32(st.Len() - 1)
	}
	_ = add("") // 0
	allocSpace := add("alloc_space")
	bytesUnit := add("bytes")
	allocObjects := add("alloc_objects")
	countUnit := add("count")
	fnName := add("allocate_chunk")
	binFile := add("/app/heap_workload")
	libFile := add("/usr/local/lib/liblibdd_profiling_heap_gotter_ffi.so")

	// Function table.
	fn := dict.FunctionTable().AppendEmpty()
	fn.SetNameStrindex(fnName)

	// Mapping table: [0]=workload binary, [1]=gotter library.
	m0 := dict.MappingTable().AppendEmpty()
	m0.SetFilenameStrindex(binFile)
	m1 := dict.MappingTable().AppendEmpty()
	m1.SetFilenameStrindex(libFile)

	// Location table:
	//  [0] symbolized frame in the workload binary
	//  [1] unsymbolized frame (mapping only) in the gotter library
	l0 := dict.LocationTable().AppendEmpty()
	l0.SetMappingIndex(0)
	ln := l0.Lines().AppendEmpty()
	ln.SetFunctionIndex(0)
	l1 := dict.LocationTable().AppendEmpty()
	l1.SetMappingIndex(1)

	// Stack: leaf first -> [gotter(unsymbolized), allocate_chunk].
	stk := dict.StackTable().AppendEmpty()
	stk.LocationIndices().Append(1)
	stk.LocationIndices().Append(0)

	rp := profiles.ResourceProfiles().AppendEmpty()
	rp.Resource().Attributes().PutStr("service.name", "full_host_alloc_test")
	sp := rp.ScopeProfiles().AppendEmpty()

	// alloc_space (bytes)
	p1 := sp.Profiles().AppendEmpty()
	p1.SampleType().SetTypeStrindex(allocSpace)
	p1.SampleType().SetUnitStrindex(bytesUnit)
	s1 := p1.Samples().AppendEmpty()
	s1.SetStackIndex(0)
	s1.Values().Append(65536)

	// A sampling-style profile whose count is encoded as timestamps rather
	// than an explicit value (as the eBPF CPU "samples" profile does).
	p3 := sp.Profiles().AppendEmpty()
	p3.SampleType().SetTypeStrindex(add("samples"))
	p3.SampleType().SetUnitStrindex(countUnit)
	s3 := p3.Samples().AppendEmpty()
	s3.SetStackIndex(0)
	s3.TimestampsUnixNano().Append(1, 2, 3) // 3 occurrences, no explicit value

	// alloc_objects (count)
	p2 := sp.Profiles().AppendEmpty()
	p2.SampleType().SetTypeStrindex(allocObjects)
	p2.SampleType().SetUnitStrindex(countUnit)
	s2 := p2.Samples().AppendEmpty()
	s2.SetStackIndex(0)
	s2.Values().Append(1)

	data, err := pprofileotlp.NewExportRequestFromProfiles(profiles).MarshalProto()
	if err != nil {
		t.Fatalf("marshal proto: %v", err)
	}
	return data
}

func TestParseOTLP_HeapProfiles(t *testing.T) {
	prof, err := parseOTLP(buildOTLPRequest(t), false)
	if err != nil {
		t.Fatalf("parseOTLP: %v", err)
	}

	// Union of sample types must contain both heap types.
	got := map[string]string{}
	for _, st := range prof.SampleType {
		got[st.Type] = st.Unit
	}
	if got["alloc_space"] != "bytes" {
		t.Errorf("alloc_space unit = %q, want bytes (types=%v)", got["alloc_space"], got)
	}
	if got["alloc_objects"] != "count" {
		t.Errorf("alloc_objects unit = %q, want count (types=%v)", got["alloc_objects"], got)
	}

	if err := prof.CheckValid(); err != nil {
		t.Fatalf("converted profile invalid: %v", err)
	}

	// Reuse the real analyzer path: alloc_space samples must total 65536 and
	// carry a frame matching the workload binary.
	samples := getProfileType(t, prof, "alloc_space")
	var total int64
	var sawSymbolized, sawSynthetic bool
	for _, ss := range samples {
		total += ss.Val
		// Symbolized frame keeps its function name.
		if strings.Contains(ss.Stack, "allocate_chunk") {
			sawSymbolized = true
		}
		// Unsymbolized frame (mapping only) is given a synthetic frame named
		// after the mapping basename.
		if strings.Contains(ss.Stack, "gotter") {
			sawSynthetic = true
		}
	}
	if total != 65536 {
		t.Errorf("alloc_space total = %d, want 65536", total)
	}
	if !sawSymbolized {
		t.Errorf("expected symbolized frame allocate_chunk; stacks=%v", samples)
	}
	if !sawSynthetic {
		t.Errorf("expected synthetic mapping-basename frame for unsymbolized location; stacks=%v", samples)
	}

	// alloc_objects must total 1 and share the same stack.
	objs := getProfileType(t, prof, "alloc_objects")
	var objTotal int64
	for _, ss := range objs {
		objTotal += ss.Val
	}
	if objTotal != 1 {
		t.Errorf("alloc_objects total = %d, want 1", objTotal)
	}

	// The timestamp-encoded sampling profile must count 3 (len of timestamps).
	sampCount := getProfileType(t, prof, "samples")
	var sampTotal int64
	for _, ss := range sampCount {
		sampTotal += ss.Val
	}
	if sampTotal != 3 {
		t.Errorf("samples total = %d, want 3 (from 3 timestamps, no explicit value)", sampTotal)
	}
}

func TestReadProfileFile_OTLPSuffixDetection(t *testing.T) {
	dir := t.TempDir()
	data := buildOTLPRequest(t)
	path := dir + "/full_host_alloc_test.otlp"
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	prof, err := ReadProfileFile(path)
	if err != nil {
		t.Fatalf("ReadProfileFile: %v", err)
	}
	found := false
	for _, st := range prof.SampleType {
		if st.Type == "alloc_space" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected alloc_space sample type from .otlp file")
	}
}
