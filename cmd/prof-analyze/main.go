// prof-analyze runs the prof-correctness analyzer on a directory of pprof
// files against an expected_profile.json description, and exits non-zero if
// any assertion fails. It's the cross-platform CLI counterpart to the
// `TestAnalyze` test entry point — meant for use from external repos
// (e.g. dd-win-prof on Windows) via `go install`.
//
// Exit codes:
//   0  all assertions passed
//   1  one or more assertions failed
//   2  usage error (missing/invalid flags)
//
// Usage:
//
//	prof-analyze -expectedJson expected_profile.json -pprofPath ./out
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/DataDog/profiler-correctness/v1/analysis"
)

func main() {
	expectedJSON := flag.String("expectedJson", "", "Path to the expected_profile.json file (required)")
	pprofPath := flag.String("pprofPath", "", "Path to the directory containing pprof files (required)")
	flag.Parse()

	if *expectedJSON == "" || *pprofPath == "" {
		fmt.Fprintln(os.Stderr, "prof-analyze: -expectedJson and -pprofPath are required")
		flag.PrintDefaults()
		os.Exit(2)
	}

	r := analysis.NewStdReporter(os.Stdout, os.Stderr)
	analysis.Run(r, func() {
		analysis.AnalyzeResults(r, *expectedJSON, *pprofPath)
	})
	if r.Failed() {
		os.Exit(1)
	}
}
