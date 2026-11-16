package analysis_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/DataDog/prof-correctness/analysis"
	"go.opentelemetry.io/collector/pdata/pprofile/pprofileotlp"
)

func TestAnalyzeResults_FirstProfileTolerance(t *testing.T) {
	for _, format := range []string{"pprof", "otlp"} {
		t.Run(format, func(t *testing.T) {
			for _, tc := range []struct {
				name        string
				values      []int64
				allowFirst  bool
				wantFailure bool
			}{
				{name: "no reports", allowFirst: true, wantFailure: true},
				{name: "only report fails", values: []int64{1}, allowFirst: true, wantFailure: true},
				{name: "only report passes", values: []int64{100}, allowFirst: true},
				{name: "warmup fails, next passes", values: []int64{1, 100}, allowFirst: true},
				{name: "later report fails", values: []int64{100, 1}, allowFirst: true, wantFailure: true},
				{name: "no tolerance", values: []int64{1, 100}, wantFailure: true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					dir := t.TempDir()
					profileType := "cpu-time"
					for i, value := range tc.values {
						name := filepath.Join(dir, fmt.Sprintf("profile_%03d.%s", i, format))
						if format == "pprof" {
							if err := os.Rename(writeMinimalPprof(t, dir, "workload", value), name); err != nil {
								t.Fatal(err)
							}
						} else {
							req := pprofileotlp.NewExportRequest()
							profiles := req.Profiles()
							profiles.Dictionary().StringTable().Append("", "samples")
							p := profiles.ResourceProfiles().AppendEmpty().ScopeProfiles().AppendEmpty().Profiles().AppendEmpty()
							p.SampleType().SetTypeStrindex(1)
							p.Samples().AppendEmpty().Values().Append(value)
							data, err := req.MarshalProto()
							if err != nil {
								t.Fatal(err)
							}
							if err := os.WriteFile(name, data, 0o644); err != nil {
								t.Fatal(err)
							}
						}
					}
					if format == "otlp" {
						profileType = "samples"
					}
					jsonPath := writeJSON(t, dir, fmt.Sprintf(`{
						"pprof-regex": "^profile_.*[.](pprof|otlp)$",
						"allow_first_profile_failure": %t,
						"stacks": [{
							"profile-type": %q,
							"stack-content": [{"regular_expression": ".*"}],
							"value-matching-sum": 100,
							"error-margin": 0
						}]
					}`, tc.allowFirst, profileType))
					var output bytes.Buffer
					r := analysis.NewStdReporter(&output, &output)
					analysis.Run(r, func() { analysis.AnalyzeResults(r, jsonPath, dir) })
					if r.Failed() != tc.wantFailure {
						t.Fatalf("Failed() = %v, want %v\n%s", r.Failed(), tc.wantFailure, &output)
					}
				})
			}
		})
	}
}
