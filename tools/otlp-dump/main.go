// Command otlp_dump is a tiny, dependency-free OTLP/HTTP profiles receiver that
// writes each received export request verbatim to a ".otlp" file.
//
// The datadog-agent host-profiler only exports profiles over OTLP (no local
// file/pprof exporter). The prof-correctness analyzer reads the OTLP format
// natively (analysis/otlp.go), so this sidecar does not parse the payload - it
// just persists the raw protobuf bytes so the analyzer can read them from
// /app/data.
//
// otlphttpexporter POSTs the (unstable) profiles signal to
// "<endpoint>/v1development/profiles" as protobuf, optionally gzip-encoded.
package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
)

var seq atomic.Uint64

func writeDump(outDir string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body io.Reader = r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "bad gzip", http.StatusBadRequest)
			return
		}
		defer gz.Close()
		body = gz
	}

	data, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	n := seq.Add(1)
	name := fmt.Sprintf("profiles_%03d.otlp", n)
	if err := os.WriteFile(filepath.Join(outDir, name), data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "otlp_dump: write %s: %v\n", name, err)
		http.Error(w, "write error", http.StatusInternalServerError)
		return
	}
	fmt.Printf("otlp_dump: wrote %s (%d bytes)\n", name, len(data))

	// An empty body is a valid empty ExportProfilesServiceResponse.
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(http.StatusOK)
}

func main() {
	addr := os.Getenv("SINK_ADDR")
	if addr == "" {
		addr = "0.0.0.0:4318"
	}
	outDir := os.Getenv("SINK_OUT_DIR")
	if outDir == "" {
		outDir = "/app/data"
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "otlp_dump: mkdir %s: %v\n", outDir, err)
		os.Exit(1)
	}

	h := func(w http.ResponseWriter, r *http.Request) { writeDump(outDir, w, r) }
	mux := http.NewServeMux()
	mux.HandleFunc("/v1development/profiles", h)
	mux.HandleFunc("/v1/profiles", h) // tolerate a signal-version bump

	fmt.Printf("otlp_dump listening on %s, writing .otlp to %s\n", addr, outDir)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "otlp_dump: serve: %v\n", err)
		os.Exit(1)
	}
}
