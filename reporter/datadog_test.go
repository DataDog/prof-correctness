package reporter

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// fakeBackend captures the JSON payloads sent to /api/v1/series and
// /api/v1/distribution_points so tests can assert on metric shape.
type fakeBackend struct {
	mu    sync.Mutex
	calls map[string][]map[string]any // path -> list of decoded payloads
}

func newFakeBackend(t *testing.T) (*httptest.Server, *fakeBackend) {
	t.Helper()
	fb := &fakeBackend{calls: map[string][]map[string]any{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("DD-API-KEY") != "test-key" {
			t.Errorf("missing/incorrect DD-API-KEY header: %q", r.Header.Get("DD-API-KEY"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("wrong content type: %q", r.Header.Get("Content-Type"))
		}
		body, _ := io.ReadAll(r.Body)
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Errorf("invalid JSON body on %s: %v", r.URL.Path, err)
		}
		fb.mu.Lock()
		fb.calls[r.URL.Path] = append(fb.calls[r.URL.Path], decoded)
		fb.mu.Unlock()
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(srv.Close)
	return srv, fb
}

// rewriteToTestServer replaces the recorder's site so requests go to the
// test server. This pokes at unexported state on purpose — keeps the public
// constructor honest about its URL pattern.
func rewriteToTestServer(r *DatadogRecorder, srv *httptest.Server) {
	// Strip "https://api." prefix from the test server's URL so the
	// existing "https://api.<site>" template produces srv.URL.
	host := strings.TrimPrefix(srv.URL, "http://")
	r.site = host
	// Override the client to use http (not https) — easier than wiring TLS.
	tr := &http.Transport{Proxy: nil}
	r.client.Transport = &rewriteTransport{base: tr, target: srv.URL}
}

type rewriteTransport struct {
	base   http.RoundTripper
	target string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace scheme + host with the httptest server.
	newURL := t.target + req.URL.Path
	if req.URL.RawQuery != "" {
		newURL += "?" + req.URL.RawQuery
	}
	r2 := req.Clone(req.Context())
	u, err := req.URL.Parse(newURL)
	if err != nil {
		return nil, err
	}
	r2.URL = u
	r2.Host = u.Host
	return t.base.RoundTrip(r2)
}

func TestDatadogRecorder_FlushEmits(t *testing.T) {
	srv, fb := newFakeBackend(t)

	rec := NewDatadogRecorder("test-key", "datadoghq.com", "ci-host",
		[]string{"git_repo:DataDog/prof-correctness", "runner_label:ubuntu-8-core-latest"})
	rewriteToTestServer(rec, srv)

	// Two assertions, same scenario but different profile_type — should
	// produce one distribution point per unique tag set.
	rec.RecordAssertion(AssertionEvent{
		Scenario: "python_many_threads", Language: "python",
		ProfileType: "wall-time", Kind: "value",
		ErrorPct: 2.3, ErrorMargin: 15, Passed: true,
	})
	rec.RecordAssertion(AssertionEvent{
		Scenario: "python_many_threads", Language: "python",
		ProfileType: "wall-time", Kind: "value",
		ErrorPct: 7.1, ErrorMargin: 15, Passed: true,
	})
	rec.RecordAssertion(AssertionEvent{
		Scenario: "python_many_threads", Language: "python",
		ProfileType: "cpu-time", Kind: "value",
		ErrorPct: 12.5, ErrorMargin: 15, Passed: false,
	})
	rec.RecordScenarioResult("python_many_threads", "python", false, 32.4)

	if err := rec.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	fb.mu.Lock()
	defer fb.mu.Unlock()

	// One call to /api/v1/distribution_points and one to /api/v1/series.
	if len(fb.calls["/api/v1/distribution_points"]) != 1 {
		t.Fatalf("expected 1 distribution_points call, got %d: %+v",
			len(fb.calls["/api/v1/distribution_points"]), fb.calls)
	}
	if len(fb.calls["/api/v1/series"]) != 1 {
		t.Fatalf("expected 1 series call, got %d", len(fb.calls["/api/v1/series"]))
	}

	// Distribution should have 2 groups (wall-time + cpu-time), each with the right values.
	distPayload := fb.calls["/api/v1/distribution_points"][0]
	series, ok := distPayload["series"].([]any)
	if !ok || len(series) != 2 {
		t.Fatalf("distribution series shape unexpected: %T %v", distPayload["series"], distPayload)
	}
	totalValues := 0
	for _, s := range series {
		obj := s.(map[string]any)
		if obj["metric"] != "prof_correctness.assertion.error_pct" {
			t.Errorf("unexpected metric: %v", obj["metric"])
		}
		pts := obj["points"].([]any)
		if len(pts) != 1 {
			t.Errorf("expected 1 point per group, got %d", len(pts))
		}
		vals := pts[0].([]any)[1].([]any)
		totalValues += len(vals)
	}
	if totalValues != 3 {
		t.Errorf("expected 3 distribution values total (2 wall-time + 1 cpu-time), got %d", totalValues)
	}

	// Series should contain assertion.passed, assertion.total, scenario.failed,
	// scenario.total, scenario.duration_seconds.
	seriesPayload := fb.calls["/api/v1/series"][0]
	metrics := map[string]bool{}
	for _, s := range seriesPayload["series"].([]any) {
		metrics[s.(map[string]any)["metric"].(string)] = true
	}
	for _, want := range []string{
		"prof_correctness.assertion.passed",
		"prof_correctness.assertion.total",
		"prof_correctness.scenario.failed",
		"prof_correctness.scenario.total",
		"prof_correctness.scenario.duration_seconds",
	} {
		if !metrics[want] {
			t.Errorf("missing metric %q in series payload; got %v", want, metrics)
		}
	}
}

func TestDatadogRecorder_NoFlushOnEmpty(t *testing.T) {
	srv, fb := newFakeBackend(t)
	rec := NewDatadogRecorder("test-key", "datadoghq.com", "host", nil)
	rewriteToTestServer(rec, srv)
	if err := rec.Flush(); err != nil {
		t.Fatalf("Flush with no data: %v", err)
	}
	if len(fb.calls) != 0 {
		t.Errorf("expected no HTTP calls on empty Flush, got %d", len(fb.calls))
	}
}

func TestLanguageFromScenarioFolder(t *testing.T) {
	cases := map[string]string{
		"scenarios/python_many_threads":   "python",
		"./scenarios/ruby_heap_4x":        "ruby",
		"scenarios/ddprof":                "ddprof",
		"python_gil_contention_3.11":      "python",
		"":                                "unknown",
	}
	for in, want := range cases {
		if got := LanguageFromScenarioFolder(in); got != want {
			t.Errorf("LanguageFromScenarioFolder(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestScenarioFromFolder(t *testing.T) {
	cases := map[string]string{
		"scenarios/python_many_threads":   "python_many_threads",
		"./scenarios/ruby_heap_4x":        "ruby_heap_4x",
		"python_gil_contention_3.11":      "python_gil_contention_3.11",
	}
	for in, want := range cases {
		if got := ScenarioFromFolder(in); got != want {
			t.Errorf("ScenarioFromFolder(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNoopRecorder(t *testing.T) {
	var r MetricsRecorder = NoopRecorder{}
	r.RecordAssertion(AssertionEvent{}) // does not panic
	r.RecordScenarioResult("s", "lang", true, 1.0)
	if err := r.Flush(); err != nil {
		t.Errorf("NoopRecorder.Flush returned non-nil: %v", err)
	}
}
