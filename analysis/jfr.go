// Package analysis — JFR support.
//
// This file adds the ability to read JFR (Java Flight Recorder) files and map
// JFR events directly into the analyzer's neutral ProfileSet.  The mapping from
// event names to profile semantics intentionally lives here: prof-correctness is
// the consumer that knows which JDK / Datadog profiler events should satisfy a
// given expected_profile.json assertion.
package analysis

import (
	"fmt"
	"io"
	"strings"

	"github.com/grafana/jfr-parser/parser"
	"github.com/grafana/jfr-parser/parser/types"
)

// FromJFR builds a ProfileSet directly from Java Flight Recorder events.
func FromJFR(data []byte) (*ProfileSet, error) {
	p := parser.NewParser(data, parser.Options{
		SymbolProcessor: parser.ProcessSymbols,
	})
	ps := newProfileSet()

	var cpuTotal int64
	var durationNanos uint64
	seenChunks := map[jfrChunkKey]bool{}

	for {
		event, err := p.ParseRawEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("jfr ParseRawEvent: %w", err)
		}
		if event.Type == nil {
			continue
		}

		header := p.ChunkHeader()
		chunk := jfrChunkKey{startNanos: header.StartNanos, durationNanos: header.DurationNanos}
		if !seenChunks[chunk] {
			seenChunks[chunk] = true
			durationNanos += header.DurationNanos
		}

		switch event.Type.Name {
		case "jdk.ExecutionSample", "datadog.ExecutionSample":
			fields, err := p.DecodeRawEventFields(event)
			if err != nil {
				return nil, fmt.Errorf("jfr decode %s: %w", event.Type.Name, err)
			}
			cpuTotal += addJFRCPU(ps, p, fields)
		}
	}

	if cpuTotal > 0 && durationNanos > 0 {
		ps.addProfileDuration("cpu", cpuTotal, float64(durationNanos)/1e9)
	}

	return ps.finalize(), nil
}

type jfrChunkKey struct {
	startNanos    uint64
	durationNanos uint64
}

func addJFRCPU(ps *ProfileSet, p *parser.Parser, fields map[string]parser.RawField) int64 {
	stackField, ok := jfrField(fields, "stackTrace")
	if !ok {
		return 0
	}

	// Match jfr-parser/pprof's CPU semantics: execution samples from sleeping
	// threads do not count as CPU samples. If the state is absent, keep the
	// sample rather than silently dropping producer-specific events.
	if state, ok := jfrField(fields, "state"); ok {
		if ts := p.GetThreadState(types.ThreadStateRef(state.Uint64)); ts != nil && ts.Name == "STATE_SLEEPING" {
			return 0
		}
	}

	folded := foldJFRStack(p, types.StackTraceRef(stackField.Uint64))
	if folded == "" {
		return 0
	}

	val := int64(1)
	if weight, ok := jfrField(fields, "weight"); ok && weight.Uint64 > 0 {
		val = int64(weight.Uint64)
	}

	ps.add("cpu", StackSample{Stack: folded, Val: val, Labels: executionSampleLabels(fields)})
	return val
}

func executionSampleLabels(fields map[string]parser.RawField) map[string][]string {
	labels := map[string][]string{}
	if spanID, ok := jfrField(fields, "spanId"); ok && spanID.Uint64 != 0 {
		labels[LabelSpanID] = []string{fmt.Sprintf("%d", spanID.Uint64)}
	}
	if localRootSpanID, ok := jfrField(fields, "localRootSpanId"); ok && localRootSpanID.Uint64 != 0 {
		labels[LabelLocalRootSID] = []string{fmt.Sprintf("%d", localRootSpanID.Uint64)}
	}
	traceHi, traceHiOK := jfrField(fields, "traceIdHi")
	traceLo, traceLoOK := jfrField(fields, "traceIdLo")
	if (traceHiOK || traceLoOK) && (traceHi.Uint64 != 0 || traceLo.Uint64 != 0) {
		labels[LabelTraceID] = []string{fmt.Sprintf("%016x%016x", traceHi.Uint64, traceLo.Uint64)}
	}
	return labels
}

func jfrField(fields map[string]parser.RawField, name string) (parser.RawValue, bool) {
	field, ok := fields[name]
	if !ok {
		return parser.RawValue{}, false
	}
	return field.First()
}

func foldJFRStack(p *parser.Parser, stackRef types.StackTraceRef) string {
	st := p.GetStacktrace(stackRef)
	if st == nil || len(st.Frames) == 0 {
		return ""
	}
	// JFR frames are leaf-first.  The analyzer uses root-first folded stacks,
	// matching FromPprof and the historical expected_profile.json captures.
	frames := make([]string, 0, len(st.Frames))
	for i := len(st.Frames) - 1; i >= 0; i-- {
		name := jfrFrameName(p, st.Frames[i].Method)
		if name != "" {
			frames = append(frames, name)
		}
	}
	return strings.Join(frames, ";")
}

func jfrFrameName(p *parser.Parser, methodRef types.MethodRef) string {
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
