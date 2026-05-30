// Package runner abstracts the external agent CLIs (claude, codex, gemini,
// cursor, opencode) behind one interface. Each Run spawns a NEW process in an
// isolated workdir — a fresh process means a clean context, with no bleed from
// the previous lesson.
package runner

import "time"

// Status values for a run result.
const (
	StatusOK      = "ok"
	StatusFailed  = "failed"
	StatusTimeout = "timeout"
)

// Invocation is the immutable input to a single isolated agent session.
type Invocation struct {
	Prompt     string
	Workdir    string
	OutputPath string
	Timeout    time.Duration
	Env        map[string]string
}

// Result is the normalized output, agnostic of which tool ran.
type Result struct {
	Status   string
	Artifact string
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Err      string
}

func (r Result) OK() bool { return r.Status == StatusOK }

// Runner is the common contract. Stateless: everything comes in the Invocation.
type Runner interface {
	Name() string
	Available() bool
	Version() string
	Run(inv Invocation) Result
}
