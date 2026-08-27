// Command harvest-margins aggregates captureProfData JSON dumps produced by
// repeated scenario runs (TestFlakiness or TestScenarios) and prints n / mean /
// stdev / min / max plus a suggested integer bound that would hold every sample
// around the mean.
//
// Usage:
//
//	go run ./cmd/harvest-margins [-kind percent|value|auto] <data-dir> [more dirs...]
//
// Each argument is a directory containing captured *.json next to pprof files
// (the layout written under ./data/<scenario>-<timestamp>-*). Recurses.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type captureFile struct {
	TestName string `json:"test_name"`
	Stacks   []struct {
		ProfileType  string `json:"profile-type"`
		StackContent []struct {
			RegularExpression string `json:"regular_expression"`
			Value             *int64 `json:"value"`
			Percent           *int64 `json:"percent"`
			Labels            []struct {
				Key         string   `json:"key"`
				Values      []string `json:"values"`
				ValuesRegex string   `json:"values_regex"`
			} `json:"labels"`
		} `json:"stack-content"`
	} `json:"stacks"`
}

type aggKey struct {
	profileType string
	regex       string
	labels      string
}

func main() {
	kind := flag.String("kind", "auto", "which field to aggregate: percent, value, or auto (prefer percent when set)")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: harvest-margins [-kind percent|value|auto] <data-dir> [...]")
		os.Exit(2)
	}

	obs := map[aggKey][]float64{}
	for _, root := range flag.Args() {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".json") {
				return nil
			}
			if strings.HasPrefix(filepath.Base(path), "expected") {
				return nil
			}
			raw, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			var data captureFile
			if unmarshalErr := json.Unmarshal(raw, &data); unmarshalErr != nil {
				return nil
			}
			if data.TestName == "" || len(data.Stacks) == 0 {
				return nil
			}
			for _, st := range data.Stacks {
				for _, sc := range st.StackContent {
					k := aggKey{profileType: st.ProfileType, regex: sc.RegularExpression, labels: formatLabels(sc.Labels)}
					switch *kind {
					case "percent":
						if sc.Percent != nil {
							obs[k] = append(obs[k], float64(*sc.Percent))
						}
					case "value":
						if sc.Value != nil {
							obs[k] = append(obs[k], float64(*sc.Value))
						}
					default:
						if sc.Percent != nil {
							obs[k] = append(obs[k], float64(*sc.Percent))
						} else if sc.Value != nil {
							obs[k] = append(obs[k], float64(*sc.Value))
						}
					}
				}
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", root, err)
			os.Exit(1)
		}
	}

	if len(obs) == 0 {
		fmt.Fprintln(os.Stderr, "harvest-margins: no captured stacks found")
		os.Exit(1)
	}

	keys := make([]aggKey, 0, len(obs))
	for k := range obs {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].profileType != keys[j].profileType {
			return keys[i].profileType < keys[j].profileType
		}
		if keys[i].regex != keys[j].regex {
			return keys[i].regex < keys[j].regex
		}
		return keys[i].labels < keys[j].labels
	})

	fmt.Printf("%-12s %-40s %-20s %5s %10s %10s %10s %10s %8s\n",
		"type", "regex", "labels", "n", "mean", "stdev", "min", "max", "bound")
	for _, k := range keys {
		xs := obs[k]
		n := float64(len(xs))
		sum, minV, maxV := 0.0, xs[0], xs[0]
		for _, x := range xs {
			sum += x
			if x < minV {
				minV = x
			}
			if x > maxV {
				maxV = x
			}
		}
		mean := sum / n
		var varSum float64
		for _, x := range xs {
			d := x - mean
			varSum += d * d
		}
		stdev := math.Sqrt(varSum / n)
		bound := 0
		for _, x := range xs {
			diff := int(math.Ceil(math.Abs(x - mean)))
			if diff > bound {
				bound = diff
			}
		}
		regex := truncate(k.regex, 40)
		labels := truncate(k.labels, 20)
		fmt.Printf("%-12s %-40s %-20s %5d %10.2f %10.2f %10.2f %10.2f %8d\n",
			k.profileType, regex, labels, len(xs), mean, stdev, minV, maxV, bound)
	}
}

func formatLabels(labels []struct {
	Key         string   `json:"key"`
	Values      []string `json:"values"`
	ValuesRegex string   `json:"values_regex"`
}) string {
	if len(labels) == 0 {
		return ""
	}
	parts := make([]string, 0, len(labels))
	for _, l := range labels {
		switch {
		case len(l.Values) > 0:
			parts = append(parts, l.Key+"="+strings.Join(l.Values, ","))
		case l.ValuesRegex != "":
			parts = append(parts, l.Key+"~"+l.ValuesRegex)
		default:
			parts = append(parts, l.Key)
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, ";")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
