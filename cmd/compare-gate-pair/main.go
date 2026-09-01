// compare-gate-pair diffs 3.14 vs 3.15 capture JSONs. Each side already
// asserts a shared profile.json band; this fails when asserted stacks still
// disagree inside it (e.g. both pass 20±5 but 18 vs 24).
//
// Families pair by stripping _3.14/_3.15 from the capture folder (else
// test_name). Keys are (profile-type, regular_expression). math.factorial
// and math.integer.factorial collapse to one key. When
// scenarios/<family>/profile.json exists, only those asserted keys are
// compared. -exclude skips families.
//
// Exit 0 within -max-pp; 1 on divergence or asserted ≥5% unmatched; 2 usage.
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

type runConfig struct {
	leftDir      string
	rightDir     string
	maxPP        int64
	scenariosDir string
	exclude      []string
	stdout       io.Writer
	stderr       io.Writer
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

func normalizeRegex(s string) string {
	s = strings.ReplaceAll(s, `math\.integer\.factorial`, `math\.factorial`)
	s = strings.ReplaceAll(s, "math.integer.factorial", "math.factorial")
	return s
}

func percentsFrom(c captureFile) map[stackKey]int64 {
	out := map[stackKey]int64{}
	for _, ts := range c.Stacks {
		for _, sc := range ts.StackContent {
			if sc.Percent == nil || sc.RegularExpression == "" || ts.ProfileType == "" {
				continue
			}
			k := stackKey{profileType: ts.ProfileType, regex: normalizeRegex(sc.RegularExpression)}
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

func parseExclude(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func isExcluded(family string, exclude []string) bool {
	for _, e := range exclude {
		if family == e || family == "python_"+e || e == "python_"+family {
			return true
		}
	}
	return false
}

func readAssertedFile(path string) ([]stackKey, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var c captureFile
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, false
	}
	seen := map[stackKey]struct{}{}
	var keys []stackKey
	for _, ts := range c.Stacks {
		for _, sc := range ts.StackContent {
			if sc.RegularExpression == "" || ts.ProfileType == "" {
				continue
			}
			k := stackKey{profileType: ts.ProfileType, regex: normalizeRegex(sc.RegularExpression)}
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return nil, false
	}
	return keys, true
}

func loadAssertedKeys(scenariosDir, family string) []stackKey {
	if scenariosDir == "" {
		return nil
	}
	candidates := []string{
		filepath.Join(scenariosDir, family, "profile.json"),
		filepath.Join(scenariosDir, family+"_3.14", "expected_profile.json"),
		filepath.Join(scenariosDir, family+"_3.15", "expected_profile.json"),
		filepath.Join(scenariosDir, family+"_gate", "profile.json"),
	}
	for _, p := range candidates {
		keys, ok := readAssertedFile(p)
		if ok {
			return keys
		}
	}
	return nil
}

func foldPercents(percents map[stackKey]int64, asserted []stackKey) map[stackKey]int64 {
	if len(asserted) == 0 {
		return percents
	}
	type compiledAssert struct {
		key stackKey
		re  *regexp.Regexp
	}
	compiled := make([]compiledAssert, 0, len(asserted))
	for _, a := range asserted {
		re, err := regexp.Compile(a.regex)
		if err != nil {
			continue
		}
		compiled = append(compiled, compiledAssert{key: a, re: re})
	}
	out := map[stackKey]int64{}
	for _, a := range compiled {
		var sum int64
		matched := false
		for k, p := range percents {
			if k.profileType != a.key.profileType {
				continue
			}
			if a.re.MatchString(k.regex) {
				sum += p
				matched = true
			}
		}
		if matched {
			out[a.key] = sum
		}
	}
	return out
}

func compare(left, right map[string]loadedCapture, maxPP int64, scenariosDir string, exclude []string, w io.Writer) []string {
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
		if isExcluded(family, exclude) {
			continue
		}
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
		asserted := loadAssertedKeys(scenariosDir, family)
		lpcts := foldPercents(l.percents, asserted)
		rpcts := foldPercents(r.percents, asserted)
		keys := map[stackKey]struct{}{}
		for k := range lpcts {
			keys[k] = struct{}{}
		}
		for k := range rpcts {
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
			lp, hasL := lpcts[k]
			rp, hasR := rpcts[k]
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
	return runOpts(runConfig{
		leftDir:  leftDir,
		rightDir: rightDir,
		maxPP:    maxPP,
		stdout:   stdout,
		stderr:   stderr,
	})
}

func runOpts(cfg runConfig) error {
	if cfg.maxPP < 0 {
		return fmt.Errorf("-max-pp must be >= 0")
	}
	left, err := collectSide(cfg.leftDir)
	if err != nil {
		return fmt.Errorf("left: %w", err)
	}
	right, err := collectSide(cfg.rightDir)
	if err != nil {
		return fmt.Errorf("right: %w", err)
	}
	if len(left) == 0 {
		return fmt.Errorf("no capture JSON found under %s", cfg.leftDir)
	}
	if len(right) == 0 {
		return fmt.Errorf("no capture JSON found under %s", cfg.rightDir)
	}
	failures := compare(left, right, cfg.maxPP, cfg.scenariosDir, cfg.exclude, cfg.stdout)
	if len(failures) == 0 {
		n := 0
		seen := map[string]struct{}{}
		for f := range left {
			seen[f] = struct{}{}
		}
		for f := range right {
			seen[f] = struct{}{}
		}
		for f := range seen {
			if !isExcluded(f, cfg.exclude) {
				n++
			}
		}
		fmt.Fprintf(cfg.stdout, "ok: %d families within %d pp\n", n, cfg.maxPP)
		return nil
	}
	for _, f := range failures {
		fmt.Fprintln(cfg.stderr, f)
	}
	return fmt.Errorf("%d divergence(s)", len(failures))
}

func main() {
	left := flag.String("left", "", "directory of 3.14 capture folders (data/<scenario>-<timestamp>-*)")
	right := flag.String("right", "", "directory of 3.15 capture folders")
	maxPP := flag.Int64("max-pp", 5, "max |percent_3.14-percent_3.15| in percentage points")
	scenarios := flag.String("scenarios", "scenarios", "scenarios dir for asserted profile.json keys")
	exclude := flag.String("exclude", "", "comma-separated families to skip (e.g. python_exceptions,python_live_heap)")
	flag.Parse()

	if *left == "" || *right == "" {
		fmt.Fprintln(os.Stderr, "error: -left and -right are required")
		flag.Usage()
		os.Exit(2)
	}

	err := runOpts(runConfig{
		leftDir:      *left,
		rightDir:     *right,
		maxPP:        *maxPP,
		scenariosDir: *scenarios,
		exclude:      parseExclude(*exclude),
		stdout:       os.Stdout,
		stderr:       os.Stderr,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if strings.Contains(err.Error(), "-max-pp") {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
