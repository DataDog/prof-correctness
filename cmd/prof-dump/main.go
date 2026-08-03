// Command prof-dump prints the neutral ProfileSet the analyzer builds from a
// profile file (pprof or OTLP), so you can see what stacks, values and labels
// look like after conversion. Handy for deciding how to write expectations and
// for inspecting how each format's tags/resource attributes land as labels.
//
// Usage:
//
//	go run ./cmd/prof-dump [-n 5] <file.otlp|file.pprof> [more files...]
//
//	-n   number of sample lines to print per profile type (default 5)
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/DataDog/prof-correctness/analysis"
)

func main() {
	n := flag.Int("n", 5, "sample lines to print per profile type")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: prof-dump [-n N] <profile-file> [...]")
		os.Exit(2)
	}
	for _, path := range flag.Args() {
		dump(path, *n)
	}
}

func dump(path string, n int) {
	ps, err := analysis.LoadProfileSet(path)
	if err != nil {
		fmt.Printf("%s: ERROR %v\n\n", path, err)
		return
	}
	fmt.Printf("== %s ==\n", path)
	for _, t := range ps.SampleTypes() {
		samples, _ := ps.Samples(t)

		var total int64
		labelKeys := map[string]struct{}{}
		for _, s := range samples {
			total += s.Val
			for k := range s.Labels {
				labelKeys[k] = struct{}{}
			}
		}

		fmt.Printf("\n  profile-type=%q  samples=%d  total-value=%d  duration=%.2fs\n",
			t, len(samples), total, ps.Duration(t))
		fmt.Printf("  label-keys observed: %v\n", sortedKeys(labelKeys))

		limit := n
		if limit > len(samples) {
			limit = len(samples)
		}
		for i := 0; i < limit; i++ {
			s := samples[i]
			fmt.Printf("    [%d] val=%d labels=%s\n", i, s.Val, fmtLabels(s.Labels))
			fmt.Printf("        stack: %s\n", s.Stack)
		}
	}
	fmt.Println()
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func fmtLabels(labels map[string][]string) string {
	if len(labels) == 0 {
		return "{}"
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	s := "{"
	for i, k := range keys {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%s=%v", k, labels[k])
	}
	return s + "}"
}
