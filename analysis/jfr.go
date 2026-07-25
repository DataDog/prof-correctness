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
//  2. Each JFR file is parsed with github.com/grafana/jfr-parser/pprof.
//  3. Per-metric pprof profiles are written as <stem>_<metric>.pprof alongside
//     the source JFR file (e.g. profile.jfr → profile_cpu.pprof).
//  4. The normal pprof analysis loop then picks those up via the usual
//     filename regex.
package analysis

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/pprof/profile"
	jfrpprof "github.com/grafana/jfr-parser/pprof"
)

// parseJFR converts raw JFR bytes into a map of profile-name → pprof profile.
func parseJFR(data []byte) (map[string]*profile.Profile, error) {
	profiles, err := jfrpprof.ParseJFR(data, &jfrpprof.ParseInput{
		StartTime: time.Unix(0, 0),
		EndTime:   time.Unix(0, 0),
		// Keep prof-correctness values sample-like (1 per CPU/wall event) while
		// reusing jfr-parser's pprof conversion, which otherwise scales CPU/wall
		// samples by 1e9/SampleRate.
		SampleRate: 1_000_000_000,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("jfr ParseJFR: %w", err)
	}

	result := make(map[string]*profile.Profile, len(profiles.Profiles))
	for _, parsed := range profiles.Profiles {
		data, err := parsed.Profile.MarshalVT()
		if err != nil {
			return nil, fmt.Errorf("jfr marshal profile %s: %w", parsed.Metric, err)
		}
		prof, err := profile.ParseData(data)
		if err != nil {
			return nil, fmt.Errorf("jfr parse pprof profile %s: %w", parsed.Metric, err)
		}
		for _, fn := range prof.Function {
			fn.Name = strings.ReplaceAll(fn.Name, "/", ".")
		}
		result[jfrProfileName(parsed.Metric, prof)] = prof
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
