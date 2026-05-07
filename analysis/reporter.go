package analysis

import (
	"fmt"
	"io"
	"os"
)

// Reporter is the minimal interface used by the analyzer to log progress and
// signal failures. *testing.T satisfies it, so test code can pass `t` directly.
// Non-test consumers (CLI, other repos) can use StdReporter or implement their
// own.
type Reporter interface {
	Logf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

// fatalSentinel is the panic value used by StdReporter.Fatalf to unwind the
// stack. Run() recognises it and returns normally instead of propagating.
type fatalSentinel struct{ msg string }

// StdReporter is a Reporter that writes to plain io.Writers and tracks whether
// any error/fatal was reported. It mimics testing.T semantics: Fatalf stops
// execution (via panic), Errorf records a failure but lets execution continue.
//
// Use Run() to invoke analyzer code under a StdReporter — it recovers the
// Fatalf panic so callers can inspect Failed() and exit cleanly.
type StdReporter struct {
	Out    io.Writer
	Err    io.Writer
	failed bool
}

// NewStdReporter writes informational output to out and error output to err.
// Pass nil for either to default to os.Stdout / os.Stderr.
func NewStdReporter(out, err io.Writer) *StdReporter {
	if out == nil {
		out = os.Stdout
	}
	if err == nil {
		err = os.Stderr
	}
	return &StdReporter{Out: out, Err: err}
}

func (r *StdReporter) Logf(format string, args ...any) {
	fmt.Fprintln(r.Out, fmt.Sprintf(format, args...))
}

func (r *StdReporter) Errorf(format string, args ...any) {
	r.failed = true
	fmt.Fprintln(r.Err, fmt.Sprintf(format, args...))
}

func (r *StdReporter) Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	r.failed = true
	fmt.Fprintln(r.Err, msg)
	panic(fatalSentinel{msg: msg})
}

// Failed returns true if Errorf or Fatalf was called.
func (r *StdReporter) Failed() bool { return r.failed }

// Run invokes fn under the given StdReporter, recovering Fatalf panics so the
// caller can check r.Failed() and exit appropriately. Other panics propagate.
func Run(r *StdReporter, fn func()) {
	defer func() {
		if rec := recover(); rec != nil {
			if _, ok := rec.(fatalSentinel); ok {
				return
			}
			panic(rec)
		}
	}()
	fn()
}
