// compare-gate-pair diffs 3.14 vs 3.15 capture JSONs and exits non-zero when
// observed percents diverge. Each side already asserts a shared profile.json
// band independently; this catches the case where both pass the band (e.g.
// 20±5) but the captures still disagree (18 vs 24).
//
// Capture JSON is StackTestData written by analysis.captureProfData: a sibling
// .json next to each pprof under data/<scenario>-<timestamp>-*/.
//
// Families are paired by stripping _3.14 / _3.15 from the capture folder name
// (MkdirTemp layout from docker_helpers_test.go), falling back to test_name.
// Stacks are grouped by (profile-type, regular_expression) and compared on
// percent.
//
// Exit codes:
//
//	0  every paired stack is within -max-pp
//	1  a pair exceeds -max-pp, or a ≥5% stack is unmatched
//	2  usage error
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const unmatchedMinPP int64 = 5

var versionTok = regexp.MustCompile(`_3\.(?:14|15)`)

type stackKey struct {
	profileType string
	regex       string
}

type stackEntry struct {
	RegularExpression string `json:"regular_expression"`
	Percent           *int64 `json:"percent"`
}

type typedStacks struct {
	ProfileType  string       `json:"profile-type"`
	StackContent []stackEntry `json:"stack-content"`
}

type captureFile struct {
	TestName string        `json:"test_name"`
	Stacks   []typedStacks `json:"stacks"`
}

type loadedCapture struct {
	path     string
	family   string
	percents map[stackKey]int64
}

func familyFromName(name string) string {
	loc := versionTok.FindStringIndex(name)
	if loc == nil {
		return ""
	}
	return name[:loc[0]]
}

func familyFor(path, testName string) string {
	dir := filepath.Dir(path)
	for i := 0; i < 6 && dir != "." && dir != string(os.PathSeparator); i++ {
		if f := familyFromName(filepath.Base(dir)); f != "" {
			return f
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if f := familyFromName(testName); f != "" {
		return f
	}
	return testName
}

func percentsFrom(c captureFile) map[stackKey]int64 {
	out := map[stackKey]int64{}
	for _, ts := range c.Stacks {
		for _, sc := range ts.StackContent {
			if sc.Percent == nil || sc.RegularExpression == "" || ts.ProfileType == "" {
				continue
			}
			k := stackKey{profileType: ts.ProfileType, regex: sc.RegularExpression}
			out[k] += *sc.Percent
		}
	}
	return out
}

func loadJSON(path string) (loadedCapture, bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return loadedCapture{}, false, err
	}
	var c captureFile
	if err := json.Unmarshal(raw, &c); err != nil {
		return loadedCapture{}, false, nil
	}
	if len(c.Stacks) == 0 {
		return loadedCapture{}, false, nil
	}
	percents := percentsFrom(c)
	if len(percents) == 0 {
		return loadedCapture{}, false, nil
	}
	family := familyFor(path, c.TestName)
	if family == "" {
		return loadedCapture{}, false, fmt.Errorf("%s: cannot derive family from folder or test_name", path)
	}
	return loadedCapture{path: path, family: family, percents: percents}, true, nil
}

func collectSide(root string) (map[string]loadedCapture, error) {
	var found []loadedCapture
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		lc, ok, err := loadJSON(path)
		if err != nil {
			return err
		}
		if ok {
			found = append(found, lc)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(found, func(i, j int) bool { return found[i].path < found[j].path })
	// Last capture JSON per family wins so a later pprof (post-warmup) overrides
	// the first profile that allow_first_profile_failure would ignore.
	out := map[string]loadedCapture{}
	for _, lc := range found {
		out[lc.family] = lc
	}
	return out, nil
}

func absPP(a, b int64) int64 {
	if a < b {
		return b - a
	}
	return a - b
}

func keyString(k stackKey) string {
	return k.profileType + " " + k.regex
}

func compare(left, right map[string]loadedCapture, maxPP int64, w io.Writer) []string {
	families := map[string]struct{}{}
	for f := range left {
		families[f] = struct{}{}
	}
	for f := range right {
		families[f] = struct{}{}
	}
	names := make([]string, 0, len(families))
	for f := range families {
		names = append(names, f)
	}
	sort.Strings(names)

	var failures []string
	for _, family := range names {
		l, lOK := left[family]
		r, rOK := right[family]
		if !lOK {
			failures = append(failures, fmt.Sprintf("%s: missing on left (3.14)", family))
			continue
		}
		if !rOK {
			failures = append(failures, fmt.Sprintf("%s: missing on right (3.15)", family))
			continue
		}
		keys := map[stackKey]struct{}{}
		for k := range l.percents {
			keys[k] = struct{}{}
		}
		for k := range r.percents {
			keys[k] = struct{}{}
		}
		keyList := make([]stackKey, 0, len(keys))
		for k := range keys {
			keyList = append(keyList, k)
		}
		sort.Slice(keyList, func(i, j int) bool {
			return keyString(keyList[i]) < keyString(keyList[j])
		})
		for _, k := range keyList {
			lp, hasL := l.percents[k]
			rp, hasR := r.percents[k]
			switch {
			case hasL && hasR:
				diff := absPP(lp, rp)
				fmt.Fprintf(w, "%s %s: 3.14=%d 3.15=%d Δ=%d\n", family, keyString(k), lp, rp, diff)
				if diff > maxPP {
					failures = append(failures, fmt.Sprintf("%s %s: |%d-%d|=%d > %d", family, keyString(k), lp, rp, diff, maxPP))
				}
			case hasL && lp >= unmatchedMinPP:
				failures = append(failures, fmt.Sprintf("%s %s: %d%% on 3.14, unmatched on 3.15", family, keyString(k), lp))
			case hasR && rp >= unmatchedMinPP:
				failures = append(failures, fmt.Sprintf("%s %s: %d%% on 3.15, unmatched on 3.14", family, keyString(k), rp))
			}
		}
	}
	return failures
}

func run(leftDir, rightDir string, maxPP int64, stdout, stderr io.Writer) error {
	if maxPP < 0 {
		return fmt.Errorf("-max-pp must be >= 0")
	}
	left, err := collectSide(leftDir)
	if err != nil {
		return fmt.Errorf("left: %w", err)
	}
	right, err := collectSide(rightDir)
	if err != nil {
		return fmt.Errorf("right: %w", err)
	}
	if len(left) == 0 {
		return fmt.Errorf("no capture JSON found under %s", leftDir)
	}
	if len(right) == 0 {
		return fmt.Errorf("no capture JSON found under %s", rightDir)
	}
	failures := compare(left, right, maxPP, stdout)
	if len(failures) == 0 {
		fmt.Fprintf(stdout, "ok: %d families within %d pp\n", len(left), maxPP)
		return nil
	}
	for _, f := range failures {
		fmt.Fprintln(stderr, f)
	}
	return fmt.Errorf("%d divergence(s)", len(failures))
}

func main() {
	left := flag.String("left", "", "directory of 3.14 capture folders (data/<scenario>-<timestamp>-*)")
	right := flag.String("right", "", "directory of 3.15 capture folders")
	maxPP := flag.Int64("max-pp", 5, "max |percent_3.14-percent_3.15| in percentage points")
	flag.Parse()

	if *left == "" || *right == "" {
		fmt.Fprintln(os.Stderr, "error: -left and -right are required")
		flag.Usage()
		os.Exit(2)
	}

	err := run(*left, *right, *maxPP, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if strings.Contains(err.Error(), "-max-pp") {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
