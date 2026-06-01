// make-matrix distributes test scenarios across N parallel jobs using greedy
// LPT bin-packing by scenario duration. It reads EXECUTION_TIME_SEC from each
// scenario's Dockerfile and outputs a GitHub Actions matrix JSON, e.g.:
//
//	{"include":[{"index":0,"scenarios":"python_cpu|python_basic_3.11"},…]}
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const defaultDurationSec = 30

func scenarioDuration(scenarioDir string) int {
	f, err := os.Open(filepath.Join(scenarioDir, "Dockerfile"))
	if err != nil {
		return defaultDurationSec
	}
	defer f.Close()

	re := regexp.MustCompile(`EXECUTION_TIME_SEC[= ]"?(\d+)"?`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := re.FindStringSubmatch(scanner.Text()); m != nil {
			v, _ := strconv.Atoi(m[1])
			return v
		}
	}
	return defaultDurationSec
}

type matrixEntry struct {
	Index     int    `json:"index"`
	Scenarios string `json:"scenarios"`
}

type matrix struct {
	Include []matrixEntry `json:"include"`
}

func main() {
	pattern := flag.String("pattern", "", "Regexp to filter scenario directory names (required)")
	numJobs := flag.Int("num-jobs", 1, "Number of parallel jobs")
	scenariosDir := flag.String("scenarios-dir", "scenarios", "Path to scenarios directory")
	flag.Parse()

	if *pattern == "" {
		fmt.Fprintln(os.Stderr, "error: -pattern is required")
		os.Exit(1)
	}

	re, err := regexp.Compile(*pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid pattern %q: %v\n", *pattern, err)
		os.Exit(1)
	}

	entries, err := os.ReadDir(*scenariosDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot read %s: %v\n", *scenariosDir, err)
		os.Exit(1)
	}

	type scenario struct {
		name     string
		duration int
	}
	var scenarios []scenario
	for _, e := range entries {
		if e.IsDir() && re.MatchString(e.Name()) {
			dir := filepath.Join(*scenariosDir, e.Name())
			scenarios = append(scenarios, scenario{e.Name(), scenarioDuration(dir)})
		}
	}

	if len(scenarios) == 0 {
		// No matching directories — fall back to a single job with the raw pattern.
		out, _ := json.Marshal(matrix{Include: []matrixEntry{{Index: 0, Scenarios: *pattern}}})
		fmt.Println(string(out))
		return
	}

	// Sort longest-first for better LPT packing.
	slices.SortFunc(scenarios, func(a, b scenario) int { return b.duration - a.duration })

	n := min(*numJobs, len(scenarios))
	bins := make([][]string, n)
	totals := make([]int, n)

	for _, s := range scenarios {
		idx := 0
		for i := 1; i < n; i++ {
			if totals[i] < totals[idx] {
				idx = i
			}
		}
		bins[idx] = append(bins[idx], s.name)
		totals[idx] += s.duration
	}

	var include []matrixEntry
	for i, bin := range bins {
		if len(bin) > 0 {
			include = append(include, matrixEntry{Index: i, Scenarios: strings.Join(bin, "|")})
		}
	}

	out, _ := json.Marshal(matrix{Include: include})
	fmt.Println(string(out))
}
