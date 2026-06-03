// list-scenarios discovers scenario directories matching a regex and emits a
// JSON matrix description for GitHub Actions:
//
//	[{"shard":"1/9","regex":"(^|/)(scenario_a|scenario_b|scenario_c)$"}, ...]
//
// Each matrix entry covers up to -chunk-size scenarios; scenarios are sorted
// alphabetically and packed into chunks in order. The regex is path-anchored
// (matches the trailing directory component) because the Go test runner
// matches the regex against filepath.Dir of each Dockerfile, e.g.
// "scenarios/python_cpu" — so a bare scenario name would substring-match
// (python_cpu ⊂ python_cpu_sleep_sync_3.12).
//
// Usage:
//
//	go run ./cmd/list-scenarios -pattern 'python.*' -chunk-size 3
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type matrixEntry struct {
	// Shard is a short human-readable label like "3/9". Used for the job name.
	Shard string `json:"shard"`
	// Regex is the path-anchored regex consumed by the Go test runner via
	// TEST_SCENARIOS. It matches against filepath.Dir(dockerfile), e.g.
	// "scenarios/python_cpu", so a bare name would substring-match.
	Regex string `json:"regex"`
	// Names is the comma-separated list of scenario directory names in this
	// chunk. Used for human-facing surfaces (Slack notifications, artifact
	// names) where the regex blob is unreadable.
	Names string `json:"names"`
}

func run(pattern, scenariosDir string, chunkSize int) ([]matrixEntry, error) {
	// Anchor the user pattern so e.g. "python" doesn't accidentally match
	// "python_basic_3.10". The non-capturing group preserves precedence of
	// any alternation inside the user pattern.
	re, err := regexp.Compile(`^(?:` + pattern + `)$`)
	if err != nil {
		return nil, fmt.Errorf("invalid -pattern: %w", err)
	}

	entries, err := os.ReadDir(scenariosDir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", scenariosDir, err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() && re.MatchString(e.Name()) {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	if len(names) == 0 {
		return nil, fmt.Errorf("no scenarios matched pattern %q in %s", pattern, scenariosDir)
	}

	// Pack into chunks of at most chunkSize, preserving sorted order.
	var chunks [][]string
	for i := 0; i < len(names); i += chunkSize {
		end := i + chunkSize
		if end > len(names) {
			end = len(names)
		}
		chunks = append(chunks, names[i:end])
	}

	out := make([]matrixEntry, len(chunks))
	for i, c := range chunks {
		out[i] = matrixEntry{
			Shard: fmt.Sprintf("%d/%d", i+1, len(chunks)),
			Regex: "(^|/)(" + strings.Join(c, "|") + ")$",
			Names: strings.Join(c, ", "),
		}
	}
	return out, nil
}

func main() {
	pattern := flag.String("pattern", "", "regex selecting scenario directory names (anchored as ^pattern$)")
	scenariosDir := flag.String("scenarios-dir", "scenarios", "path to the scenarios directory")
	chunkSize := flag.Int("chunk-size", 3, "max scenarios per matrix entry")
	flag.Parse()

	if *pattern == "" {
		fmt.Fprintln(os.Stderr, "error: -pattern is required")
		flag.Usage()
		os.Exit(2)
	}
	if *chunkSize < 1 {
		fmt.Fprintln(os.Stderr, "error: -chunk-size must be >= 1")
		os.Exit(2)
	}

	abs, _ := filepath.Abs(*scenariosDir)
	out, err := run(*pattern, *scenariosDir, *chunkSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v (resolved scenarios dir: %s)\n", err, abs)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding output: %v\n", err)
		os.Exit(1)
	}
}
