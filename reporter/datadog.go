package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DatadogRecorder buffers metric points in memory and submits them to the
// Datadog HTTP API on Flush(). Designed for short-lived CI processes — one
// flush per test process, batched in a single HTTP call per metric type.
//
// Distribution metrics use POST /api/v1/distribution_points.
// Count / gauge metrics use POST /api/v1/series.
// Both authenticate via the DD-API-KEY header.
type DatadogRecorder struct {
	apiKey  string
	site    string // e.g. "datadoghq.com", "datad0g.com"
	source  string // sent as `host` tag-equivalent on metric points
	common  []string // tags added to every point (e.g. runner_label, git_repo)
	client  *http.Client

	// Buffered series. Distributions are accumulated per (metric, tag set)
	// and emitted as a single list-of-values point. Counts are summed.
	dist   safeBuffer[distPoint]
	counts safeBuffer[countPoint]
	gauges safeBuffer[gaugePoint]
}

// NewDatadogRecorder constructs a recorder. site defaults to "datadoghq.com"
// if empty. commonTags are added to every metric point — typically the
// runner_label, git_repo, git_branch. host is reported as the metric `host`;
// "github-actions" is a sensible default for GHA jobs.
func NewDatadogRecorder(apiKey, site, host string, commonTags []string) *DatadogRecorder {
	if site == "" {
		site = "datadoghq.com"
	}
	if host == "" {
		host = "github-actions"
	}
	return &DatadogRecorder{
		apiKey: apiKey,
		site:   site,
		source: host,
		common: commonTags,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

type distPoint struct {
	metric string
	tags   []string
	ts     int64
	values []float64
}

type countPoint struct {
	metric string
	tags   []string
	ts     int64
	value  float64
}

type gaugePoint struct {
	metric string
	tags   []string
	ts     int64
	value  float64
}

// --- MetricsRecorder impl ----------------------------------------------------

// AssertionFromAnalysis bridges the analysis package's AssertionEvent struct
// to this package's internal AssertionContext, so DatadogRecorder satisfies
// the analysis.MetricsSink interface without analysis depending on us.
// Each translation unit defines its own struct shape; this is the join point.
type AssertionEvent struct {
	Scenario, Language, ProfileType, Kind string
	ErrorPct                              float64
	ErrorMargin                           int64
	Passed, AllowFailure                  bool
}

func (d *DatadogRecorder) RecordAssertion(ev AssertionEvent) {
	ctx := AssertionContext{
		Scenario: ev.Scenario, Language: ev.Language, ProfileType: ev.ProfileType,
		Kind: AssertionKind(ev.Kind), ErrorPct: ev.ErrorPct,
		ErrorMargin: ev.ErrorMargin, Passed: ev.Passed,
		AllowFailure: ev.AllowFailure,
	}
	d.recordAssertion(ctx)
}

func (d *DatadogRecorder) recordAssertion(ctx AssertionContext) {
	ts := time.Now().Unix()
	tags := d.assertionTags(ctx)

	// error_pct as distribution — preserves shape across many assertions per scenario.
	d.dist.append(distPoint{
		metric: "prof_correctness.assertion.error_pct",
		tags:   tags,
		ts:     ts,
		values: []float64{ctx.ErrorPct},
	})
	// passed as count — sum yields pass count; combine with total count
	// (datadog's `as_count()` modifier) for a pass rate.
	var passed float64
	if ctx.Passed {
		passed = 1
	}
	d.counts.append(countPoint{
		metric: "prof_correctness.assertion.passed",
		tags:   tags,
		ts:     ts,
		value:  passed,
	})
	// Also a 1-per-assertion counter so we can compute pass rate.
	d.counts.append(countPoint{
		metric: "prof_correctness.assertion.total",
		tags:   tags,
		ts:     ts,
		value:  1,
	})
}

func (d *DatadogRecorder) RecordScenarioResult(scenario, language string, failed bool, durationSeconds float64) {
	ts := time.Now().Unix()
	tags := append(append([]string(nil), d.common...),
		"scenario:"+scenario,
		"language:"+language,
	)
	var f float64
	if failed {
		f = 1
	}
	d.counts.append(countPoint{
		metric: "prof_correctness.scenario.failed",
		tags:   tags,
		ts:     ts,
		value:  f,
	})
	d.counts.append(countPoint{
		metric: "prof_correctness.scenario.total",
		tags:   tags,
		ts:     ts,
		value:  1,
	})
	d.gauges.append(gaugePoint{
		metric: "prof_correctness.scenario.duration_seconds",
		tags:   tags,
		ts:     ts,
		value:  durationSeconds,
	})
}

func (d *DatadogRecorder) assertionTags(ctx AssertionContext) []string {
	out := make([]string, 0, len(d.common)+6)
	out = append(out, d.common...)
	out = append(out,
		"scenario:"+ctx.Scenario,
		"language:"+ctx.Language,
		"profile_type:"+ctx.ProfileType,
		"assertion_kind:"+string(ctx.Kind),
	)
	if ctx.AllowFailure {
		out = append(out, "allow_failure:true")
	}
	return out
}

// --- Submission --------------------------------------------------------------

func (d *DatadogRecorder) Flush() error {
	if err := d.flushDist(); err != nil {
		return fmt.Errorf("distribution_points: %w", err)
	}
	if err := d.flushSeries(); err != nil {
		return fmt.Errorf("series: %w", err)
	}
	return nil
}

func (d *DatadogRecorder) flushDist() error {
	pts := d.dist.drain()
	if len(pts) == 0 {
		return nil
	}
	type apiDistPoint struct {
		Metric string      `json:"metric"`
		Host   string      `json:"host,omitempty"`
		Tags   []string    `json:"tags,omitempty"`
		// Distributions submit points as [[timestamp, [values...]]].
		Points [][2]any `json:"points"`
	}
	payload := struct {
		Series []apiDistPoint `json:"series"`
	}{}
	// Group by (metric, sorted-tags) so we send one point per group.
	type key struct {
		metric string
		tags   string
	}
	grouped := map[key]*apiDistPoint{}
	for _, p := range pts {
		k := key{p.metric, joinSortedTags(p.tags)}
		ap, ok := grouped[k]
		if !ok {
			ap = &apiDistPoint{
				Metric: p.metric,
				Host:   d.source,
				Tags:   sortedTagsCopy(p.tags),
				Points: [][2]any{{p.ts, []float64{}}},
			}
			grouped[k] = ap
		}
		// First (and only) point's value-array is the second element.
		vals := ap.Points[0][1].([]float64)
		vals = append(vals, p.values...)
		ap.Points[0][1] = vals
	}
	for _, ap := range grouped {
		payload.Series = append(payload.Series, *ap)
	}
	return d.postJSON("/api/v1/distribution_points", payload)
}

func (d *DatadogRecorder) flushSeries() error {
	counts := d.counts.drain()
	gauges := d.gauges.drain()
	if len(counts) == 0 && len(gauges) == 0 {
		return nil
	}
	type apiSeriesPoint struct {
		Metric   string      `json:"metric"`
		Host     string      `json:"host,omitempty"`
		Tags     []string    `json:"tags,omitempty"`
		Type     string      `json:"type"`
		Interval *int64      `json:"interval,omitempty"`
		Points   [][2]float64 `json:"points"`
	}
	payload := struct {
		Series []apiSeriesPoint `json:"series"`
	}{}

	// Counts: sum values by (metric, tags) so the API receives one point per
	// group. Interval=1 marks them as 1-second buckets; the agent normalises.
	type key struct{ m, t string }
	cgrouped := map[key]*apiSeriesPoint{}
	intervalOne := int64(1)
	for _, c := range counts {
		k := key{c.metric, joinSortedTags(c.tags)}
		ap, ok := cgrouped[k]
		if !ok {
			ap = &apiSeriesPoint{
				Metric:   c.metric,
				Host:     d.source,
				Tags:     sortedTagsCopy(c.tags),
				Type:     "count",
				Interval: &intervalOne,
				Points:   [][2]float64{{float64(c.ts), 0}},
			}
			cgrouped[k] = ap
		}
		ap.Points[0][1] += c.value
	}
	for _, ap := range cgrouped {
		payload.Series = append(payload.Series, *ap)
	}

	// Gauges: last value wins per (metric, tags) — there's typically only one anyway.
	ggrouped := map[key]*apiSeriesPoint{}
	for _, g := range gauges {
		k := key{g.metric, joinSortedTags(g.tags)}
		ggrouped[k] = &apiSeriesPoint{
			Metric: g.metric,
			Host:   d.source,
			Tags:   sortedTagsCopy(g.tags),
			Type:   "gauge",
			Points: [][2]float64{{float64(g.ts), g.value}},
		}
	}
	for _, ap := range ggrouped {
		payload.Series = append(payload.Series, *ap)
	}

	return d.postJSON("/api/v1/series", payload)
}

func (d *DatadogRecorder) postJSON(path string, body any) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	url := "https://api." + d.site + path
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", d.apiKey)
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Datadog returns 202 Accepted on success. Anything else is a problem.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("POST %s: HTTP %d: %s", path, resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return nil
}

// --- helpers ----------------------------------------------------------------

// joinSortedTags returns a key suitable for map de-duplication: the tags
// joined by '\x00' after a stable sort. We don't reuse strings.Join because
// we want the side-effect-free version of the tag list to come out of
// sortedTagsCopy.
func joinSortedTags(tags []string) string {
	cp := sortedTagsCopy(tags)
	return strings.Join(cp, "\x00")
}

func sortedTagsCopy(tags []string) []string {
	cp := append([]string(nil), tags...)
	// Insertion sort — tag lists are short (<10).
	for i := 1; i < len(cp); i++ {
		for j := i; j > 0 && cp[j-1] > cp[j]; j-- {
			cp[j-1], cp[j] = cp[j], cp[j-1]
		}
	}
	return cp
}
