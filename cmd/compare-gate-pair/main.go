// compare-gate-pair diffs 3.14 vs 3.15 capture percents on asserted
// (profile-type, regex, labels) keys from the family's expected_profile.json.
// Exit 1 if |Δ| > -max-pp or an asserted key is unmatched at ≥5%.
// math.factorial and math.integer.factorial are one key. -exclude skips families.
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
	"slices"
	"sort"
	"strings"
)

const unmatchedMinPP int64 = 5

var (
	versionTok  = regexp.MustCompile(`_3\.(?:14|15)`)
	unquoteMeta = regexp.MustCompile(`\\(.)`)
	// captureProfData writes <pprof-base>.json: profile.json, profile.<pid>.json,
	// or profiles.<pid>.<n>.json. Sidecars (*.info.json, *.internal_metadata.json)
	// and expected_profile.json do not match.
	captureJSON = regexp.MustCompile(`^profiles?(?:\.\d+)*\.json$`)
)

type labelSpec struct {
	Key         string   `json:"key"`
	Values      []string `json:"values"`
	ValuesRegex string   `json:"values_regex"`
}

type stackKey struct {
	profileType string
	regex       string
	labels      string
}

type stackEntry struct {
	RegularExpression string      `json:"regular_expression"`
	Percent           *int64      `json:"percent"`
	Labels            []labelSpec `json:"labels"`
}

type typedStacks struct {
	ProfileType  string       `json:"profile-type"`
	StackContent []stackEntry `json:"stack-content"`
}

type captureFile struct {
	TestName string        `json:"test_name"`
	Stacks   []typedStacks `json:"stacks"`
}

type stackObs struct {
	percent int64
	labels  []labelSpec
}

type assertedKey struct {
	stackKey
	labelSpecs []labelSpec
}

type loadedCapture struct {
	family   string
	percents map[stackKey]stackObs
}

type runConfig struct {
	leftDir, rightDir, scenariosDir string
	maxPP                           int64
	exclude                         []string
	stdout, stderr                  io.Writer
}

func familyFromName(name string) string {
	loc := versionTok.FindStringIndex(name)
	if loc == nil {
		return ""
	}
	return name[:loc[0]]
}

func familyOf(path, testName string) string {
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if f := familyFromName(part); f != "" {
			return f
		}
	}
	if f := familyFromName(testName); f != "" {
		return f
	}
	return testName
}

func normalizeRegex(s string) string {
	s = strings.ReplaceAll(s, `math\.integer\.factorial`, `math\.factorial`)
	return strings.ReplaceAll(s, "math.integer.factorial", "math.factorial")
}

func labelsKey(labels []labelSpec) string {
	if len(labels) == 0 {
		return ""
	}
	sorted := append([]labelSpec(nil), labels...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Key < sorted[j].Key })
	var b strings.Builder
	for _, l := range sorted {
		b.WriteString(l.Key)
		b.WriteByte('=')
		vals := append([]string(nil), l.Values...)
		sort.Strings(vals)
		for _, v := range vals {
			b.WriteString(v)
			b.WriteByte(',')
		}
		if l.ValuesRegex != "" {
			b.WriteByte('~')
			b.WriteString(l.ValuesRegex)
		}
		b.WriteByte(';')
	}
	return b.String()
}

func labelsMatch(asserted, capture []labelSpec) bool {
	if len(asserted) == 0 {
		return true
	}
	byKey := make(map[string]labelSpec, len(capture))
	for _, c := range capture {
		byKey[c.Key] = c
	}
	for _, a := range asserted {
		c, ok := byKey[a.Key]
		if !ok {
			return false
		}
		if a.ValuesRegex != "" {
			re, err := regexp.Compile(a.ValuesRegex)
			if err != nil {
				return false
			}
			if len(c.Values) == 0 {
				return false
			}
			for _, v := range c.Values {
				if !re.MatchString(v) {
					return false
				}
			}
			continue
		}
		if len(a.Values) != len(c.Values) {
			return false
		}
		av := append([]string(nil), a.Values...)
		cv := append([]string(nil), c.Values...)
		sort.Strings(av)
		sort.Strings(cv)
		if !slices.Equal(av, cv) {
			return false
		}
	}
	return true
}

func percentsFrom(c captureFile) map[stackKey]stackObs {
	out := map[stackKey]stackObs{}
	for _, ts := range c.Stacks {
		for _, sc := range ts.StackContent {
			if sc.Percent == nil || sc.RegularExpression == "" || ts.ProfileType == "" {
				continue
			}
			k := stackKey{ts.ProfileType, normalizeRegex(sc.RegularExpression), labelsKey(sc.Labels)}
			obs := out[k]
			obs.percent += *sc.Percent
			if obs.labels == nil {
				obs.labels = sc.Labels
			}
			out[k] = obs
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
		return loadedCapture{}, false, err
	}
	percents := percentsFrom(c)
	if len(percents) == 0 {
		return loadedCapture{}, false, nil
	}
	family := familyOf(path, c.TestName)
	if family == "" {
		return loadedCapture{}, false, nil
	}
	return loadedCapture{family: family, percents: percents}, true, nil
}

func collectSide(root string) (map[string]loadedCapture, error) {
	out := map[string]loadedCapture{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !captureJSON.MatchString(d.Name()) {
			return nil
		}
		lc, ok, err := loadJSON(path)
		if err != nil {
			return err
		}
		if ok {
			out[lc.family] = lc
		}
		return nil
	})
	return out, err
}

func parseExclude(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' })
}

func loadAssertedKeys(scenariosDir, family string) []assertedKey {
	if scenariosDir == "" {
		return nil
	}
	for _, ver := range []string{"_3.14", "_3.15"} {
		keys := keysFromFile(filepath.Join(scenariosDir, family+ver, "expected_profile.json"))
		if len(keys) > 0 {
			return keys
		}
	}
	return nil
}

func keysFromFile(path string) []assertedKey {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var c captureFile
	if json.Unmarshal(raw, &c) != nil {
		return nil
	}
	seen := map[stackKey]struct{}{}
	var keys []assertedKey
	for _, ts := range c.Stacks {
		for _, sc := range ts.StackContent {
			if sc.RegularExpression == "" || ts.ProfileType == "" {
				continue
			}
			k := stackKey{ts.ProfileType, normalizeRegex(sc.RegularExpression), labelsKey(sc.Labels)}
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			keys = append(keys, assertedKey{stackKey: k, labelSpecs: sc.Labels})
		}
	}
	return keys
}

func flattenPercents(percents map[stackKey]stackObs) map[stackKey]int64 {
	out := map[stackKey]int64{}
	for k, o := range percents {
		out[k] = o.percent
	}
	return out
}

func foldPercents(percents map[stackKey]stackObs, asserted []assertedKey) map[stackKey]int64 {
	if len(asserted) == 0 {
		return flattenPercents(percents)
	}
	out := map[stackKey]int64{}
	for _, a := range asserted {
		re, err := regexp.Compile(a.regex)
		if err != nil {
			continue
		}
		for k, o := range percents {
			body := unquoteMeta.ReplaceAllString(strings.TrimSuffix(strings.TrimPrefix(k.regex, "^"), "$"), "$1")
			if k.profileType == a.profileType && re.MatchString(body) && labelsMatch(a.labelSpecs, o.labels) {
				out[a.stackKey] += o.percent
			}
		}
	}
	return out
}

func families(left, right map[string]loadedCapture, skip []string) []string {
	seen := map[string]struct{}{}
	for f := range left {
		seen[f] = struct{}{}
	}
	for f := range right {
		seen[f] = struct{}{}
	}
	var names []string
	for f := range seen {
		if !slices.Contains(skip, f) {
			names = append(names, f)
		}
	}
	sort.Strings(names)
	return names
}

func compare(left, right map[string]loadedCapture, maxPP int64, scenariosDir string, skip []string, w io.Writer) []string {
	var failures []string
	for _, family := range families(left, right, skip) {
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
		if len(asserted) > 0 && len(l.percents) > 0 && len(r.percents) > 0 && len(lpcts) == 0 && len(rpcts) == 0 {
			failures = append(failures, fmt.Sprintf("%s: folded to nothing (asserted regexes matched no capture keys)", family))
			continue
		}
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
			li := keyList[i].profileType + keyList[i].regex + keyList[i].labels
			lj := keyList[j].profileType + keyList[j].regex + keyList[j].labels
			return li < lj
		})
		for _, k := range keyList {
			label := k.profileType + " " + k.regex
			if k.labels != "" {
				label += " " + k.labels
			}
			lp, hasL := lpcts[k]
			rp, hasR := rpcts[k]
			switch {
			case hasL && hasR:
				diff := lp - rp
				if diff < 0 {
					diff = -diff
				}
				fmt.Fprintf(w, "%s %s: 3.14=%d 3.15=%d Δ=%d\n", family, label, lp, rp, diff)
				if diff > maxPP {
					failures = append(failures, fmt.Sprintf("%s %s: |%d-%d|=%d > %d", family, label, lp, rp, diff, maxPP))
				}
			case hasL && lp >= unmatchedMinPP:
				failures = append(failures, fmt.Sprintf("%s %s: %d%% on 3.14, unmatched on 3.15", family, label, lp))
			case hasR && rp >= unmatchedMinPP:
				failures = append(failures, fmt.Sprintf("%s %s: %d%% on 3.15, unmatched on 3.14", family, label, rp))
			}
		}
	}
	return failures
}

func run(cfg runConfig) error {
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
		fmt.Fprintf(cfg.stdout, "ok: %d families within %d pp\n", len(families(left, right, cfg.exclude)), cfg.maxPP)
		return nil
	}
	for _, f := range failures {
		fmt.Fprintln(cfg.stderr, f)
	}
	return fmt.Errorf("%d divergence(s)", len(failures))
}

func main() {
	left := flag.String("left", "", "directory of 3.14 capture folders")
	right := flag.String("right", "", "directory of 3.15 capture folders")
	maxPP := flag.Int64("max-pp", 5, "max |percent_3.14-percent_3.15| in percentage points")
	scenarios := flag.String("scenarios", "scenarios", "scenarios dir for asserted expected_profile.json keys")
	exclude := flag.String("exclude", "", "comma-separated families to skip")
	flag.Parse()

	if *left == "" || *right == "" {
		fmt.Fprintln(os.Stderr, "error: -left and -right are required")
		flag.Usage()
		os.Exit(2)
	}

	err := run(runConfig{
		leftDir: *left, rightDir: *right, maxPP: *maxPP,
		scenariosDir: *scenarios, exclude: parseExclude(*exclude),
		stdout: os.Stdout, stderr: os.Stderr,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
