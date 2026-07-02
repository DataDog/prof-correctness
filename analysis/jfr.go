// Package analysis — JFR support.
//
// This file adds the ability to read JFR (Java Flight Recorder) files and
// convert them to pprof profiles that the rest of the analysis pipeline can
// consume.
//
// Flow:
//
//  1. At the start of AnalyzeResults, convertJFRFiles walks the output folder
//     for any *.jfr files.
//  2. Each JFR file is parsed with github.com/grafana/jfr-parser/parser.
//  3. Per-metric pprof profiles are written as <stem>_<metric>.pprof alongside
//     the source JFR file (e.g. profile.jfr → profile_cpu.pprof).
//  4. The normal pprof analysis loop then picks those up via the usual
//     filename regex.
//
// Supported JFR metrics (= pprof profile-type names):
//   - "cpu"                : jdk.ExecutionSample (non-sleeping threads)
//   - "wall"               : jdk.ExecutionSample when event=wall, or
//                            Datadog WallClockSample
//   - "alloc_in_tlab"      : jdk.ObjectAllocationInNewTLAB
//   - "alloc_outside_tlab" : jdk.ObjectAllocationOutsideTLAB
//   - "lock"               : jdk.JavaMonitorEnter
package analysis

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/pprof/profile"
	"github.com/grafana/jfr-parser/parser"
	"github.com/grafana/jfr-parser/parser/types"
)

// parseJFR converts raw JFR bytes into a map of metric name → pprof profile.
func parseJFR(data []byte) (map[string]*profile.Profile, error) {
	p := parser.NewParser(data, parser.Options{
		SymbolProcessor: parser.ProcessSymbols,
	})

	type builder struct {
		prof     *profile.Profile
		mapping  *profile.Mapping
		funcByID map[types.MethodRef]*profile.Function
		locByID  map[types.MethodRef]*profile.Location
	}

	builders := make(map[string]*builder)

	getBuilder := func(metric, sampleType, sampleUnit string) *builder {
		b, ok := builders[metric]
		if !ok {
			m := &profile.Mapping{ID: 1, HasFunctions: true}
			b = &builder{
				prof: &profile.Profile{
					SampleType: []*profile.ValueType{{Type: sampleType, Unit: sampleUnit}},
					PeriodType: &profile.ValueType{Type: sampleType, Unit: "nanoseconds"},
					Period:     10_000_000, // default 100 Hz
					Mapping:    []*profile.Mapping{m},
				},
				mapping:  m,
				funcByID: make(map[types.MethodRef]*profile.Function),
				locByID:  make(map[types.MethodRef]*profile.Location),
			}
			builders[metric] = b
		}
		return b
	}

	// resolveFrameName returns "ClassName.methodName" for a JFR method reference.
	// JVM internal '/' separators in class names are normalised to '.'.
	resolveFrameName := func(methodRef types.MethodRef) string {
		m := p.GetMethod(methodRef)
		if m == nil {
			return ""
		}
		methodName := p.GetSymbolString(m.Name)
		cls := p.GetClass(m.Type)
		if cls == nil {
			return methodName
		}
		clsName := strings.ReplaceAll(p.GetSymbolString(cls.Name), "/", ".")
		return clsName + "." + methodName
	}

	// addSample appends one stack-trace observation to the named metric profile.
	addSample := func(metric, sampleType, sampleUnit string, stackRef types.StackTraceRef, count int64) {
		st := p.GetStacktrace(stackRef)
		if st == nil || len(st.Frames) == 0 {
			return
		}
		b := getBuilder(metric, sampleType, sampleUnit)

		// JFR frames[0] = leaf (top of stack), which matches the pprof convention
		// that sample.Location[0] is the leaf.
		locs := make([]*profile.Location, 0, len(st.Frames))
		for _, frame := range st.Frames {
			loc, ok := b.locByID[frame.Method]
			if !ok {
				fnName := resolveFrameName(frame.Method)
				if fnName == "" {
					continue
				}
				fn, fnOK := b.funcByID[frame.Method]
				if !fnOK {
					fn = &profile.Function{
						ID:   uint64(len(b.prof.Function) + 1),
						Name: fnName,
					}
					b.prof.Function = append(b.prof.Function, fn)
					b.funcByID[frame.Method] = fn
				}
				loc = &profile.Location{
					ID:      uint64(len(b.prof.Location) + 1),
					Mapping: b.mapping,
					Line:    []profile.Line{{Function: fn}},
				}
				b.prof.Location = append(b.prof.Location, loc)
				b.locByID[frame.Method] = loc
			}
			locs = append(locs, loc)
		}
		if len(locs) == 0 {
			return
		}
		b.prof.Sample = append(b.prof.Sample, &profile.Sample{
			Location: locs,
			Value:    []int64{count},
		})
	}

	var event string
	for {
		typ, err := p.ParseEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Non-fatal: a truncated JFR file (e.g. from dumponexit=true) may end
			// mid-chunk. Return what we have so far plus the error description.
			return nil, fmt.Errorf("jfr ParseEvent: %w", err)
		}

		switch typ {
		case p.TypeMap.T_EXECUTION_SAMPLE:
			ts := p.GetThreadState(p.ExecutionSample.State)
			if ts != nil && ts.Name != "STATE_SLEEPING" {
				addSample("cpu", "cpu", "samples", p.ExecutionSample.StackTrace, 1)
			}
			if event == "wall" {
				addSample("wall", "wall", "samples", p.ExecutionSample.StackTrace, 1)
			}
		case p.TypeMap.T_WALL_CLOCK_SAMPLE:
			addSample("wall", "wall", "samples",
				p.WallClockSample.StackTrace, int64(p.WallClockSample.Samples))
		case p.TypeMap.T_ALLOC_IN_NEW_TLAB:
			addSample("alloc_in_tlab", "alloc_in_new_tlab_objects", "count",
				p.ObjectAllocationInNewTLAB.StackTrace, 1)
		case p.TypeMap.T_ALLOC_OUTSIDE_TLAB:
			addSample("alloc_outside_tlab", "alloc_outside_tlab_objects", "count",
				p.ObjectAllocationOutsideTLAB.StackTrace, 1)
		case p.TypeMap.T_MONITOR_ENTER:
			addSample("lock", "contentions", "count",
				p.JavaMonitorEnter.StackTrace, 1)
		case p.TypeMap.T_ACTIVE_SETTING:
			if p.ActiveSetting.Name == "event" {
				event = p.ActiveSetting.Value
			}
		}
	}

	result := make(map[string]*profile.Profile, len(builders))
	for metric, b := range builders {
		result[metric] = b.prof
	}
	return result, nil
}

// convertJFRFiles walks dir for *.jfr files and converts each to a set of
// per-metric pprof files written into the same directory.
// Output files are named <stem>_<metric>.pprof (e.g. profile_cpu.pprof).
// Errors are non-fatal: they are logged through r and the function continues.
func convertJFRFiles(r Reporter, dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		r.Logf("convertJFRFiles: reading dir %s: %v", dir, err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jfr") {
			continue
		}
		jfrPath := filepath.Join(dir, entry.Name())
		stem := strings.TrimSuffix(entry.Name(), ".jfr")

		data, err := os.ReadFile(jfrPath)
		if err != nil {
			r.Logf("convertJFRFiles: reading %s: %v", jfrPath, err)
			continue
		}

		profiles, err := parseJFR(data)
		if err != nil {
			r.Logf("convertJFRFiles: parsing %s: %v", jfrPath, err)
			continue
		}
		if len(profiles) == 0 {
			r.Logf("convertJFRFiles: no profiles found in %s", jfrPath)
			continue
		}

		for metric, prof := range profiles {
			outPath := filepath.Join(dir, stem+"_"+metric+".pprof")
			var buf bytes.Buffer
			if err := prof.Write(&buf); err != nil {
				r.Logf("convertJFRFiles: serialising %s metric %s: %v", entry.Name(), metric, err)
				continue
			}
			if err := os.WriteFile(outPath, buf.Bytes(), 0644); err != nil {
				r.Logf("convertJFRFiles: writing %s: %v", outPath, err)
				continue
			}
			r.Logf("Converted JFR %s → %s (%d samples)",
				entry.Name(), filepath.Base(outPath), len(prof.Sample))
		}
	}
}
