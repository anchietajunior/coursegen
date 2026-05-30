// Package state persists run history in a single JSON file
// (.coursegen/state.json).
//
// Why JSON and not SQLite? The MVP runs sequentially in a single process, so we
// need neither concurrent writers nor SQL queries — and JSON keeps the binary
// dependency-free. The API here is store-shaped so a SQLite backend can be
// dropped in later without touching callers.
package state

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/coursegen/coursegen/internal/course"
)

const schemaVersion = 1

// Execution is one lesson's production attempt within a run.
type Execution struct {
	Unit       string   `json:"unit"`
	Module     string   `json:"module"`
	Lesson     string   `json:"lesson"`
	Status     string   `json:"status"` // pending|running|ok|failed|skipped
	Attempt    int      `json:"attempt"`
	InputHash  string   `json:"input_hash"`
	OutputPath string   `json:"output_path"`
	Tokens     int      `json:"tokens"`
	DurationMs int64    `json:"duration_ms"`
	Error      string   `json:"error,omitempty"`
	Issues     []string `json:"issues,omitempty"`
}

// Run groups the executions of a single `tasks run` invocation.
type Run struct {
	ID         string       `json:"id"`
	Task       string       `json:"task"`
	Runner     string       `json:"runner"`
	Cmd        string       `json:"cmd"`
	Status     string       `json:"status"` // running|ok|partial|failed
	Total      int          `json:"total"`
	Version    string       `json:"version"`
	StartedAt  string       `json:"started_at"`
	FinishedAt string       `json:"finished_at"`
	Tokens     int          `json:"tokens"`
	Executions []*Execution `json:"executions"`
}

// Store is the persisted document.
type Store struct {
	path   string
	Schema int               `json:"schema"`
	Runs   []*Run            `json:"runs"`
	Cache  map[string]string `json:"cache"`
}

// Load reads the store from path, creating an empty one if absent.
func Load(path string) (*Store, error) {
	s := &Store{path: path, Schema: schemaVersion, Cache: map[string]string{}}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	} else if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}
	s.path = path
	if s.Cache == nil {
		s.Cache = map[string]string{}
	}
	return s, nil
}

// Save writes the store to disk (pretty-printed).
func (s *Store) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// CreateRun prepends a new running run.
func (s *Store) CreateRun(task, runner, cmd string, total int, version string) *Run {
	run := &Run{
		ID: generateID(), Task: task, Runner: runner, Cmd: cmd,
		Status: "running", Total: total, Version: version,
		StartedAt: time.Now().Format(time.RFC3339), Executions: []*Execution{},
	}
	s.Runs = append([]*Run{run}, s.Runs...)
	return run
}

// UpsertExecution finds an existing execution for the lesson or creates one.
func (s *Store) UpsertExecution(run *Run, l course.Lesson) *Execution {
	for _, e := range run.Executions {
		if e.Unit == l.Unit() {
			return e
		}
	}
	e := &Execution{
		Unit: l.Unit(), Module: l.ModuleID, Lesson: l.LessonID, Status: "pending",
	}
	run.Executions = append(run.Executions, e)
	return e
}

// Cached reports whether we already produced this exact output for the unit.
func (s *Store) Cached(unit, fingerprint, outputPath string) bool {
	if s.Cache[unit] != fingerprint {
		return false
	}
	_, err := os.Stat(outputPath)
	return err == nil
}

// Remember records a successful fingerprint for idempotent skipping.
func (s *Store) Remember(unit, fingerprint string) { s.Cache[unit] = fingerprint }

// LatestRun returns the most recent run, or nil.
func (s *Store) LatestRun() *Run {
	if len(s.Runs) == 0 {
		return nil
	}
	return s.Runs[0]
}

// FindRun returns the run with the given id, or nil.
func (s *Store) FindRun(id string) *Run {
	for _, r := range s.Runs {
		if r.ID == id {
			return r
		}
	}
	return nil
}

func generateID() string {
	b := make([]byte, 2)
	_, _ = rand.Read(b)
	return "run_" + time.Now().Format("20060102_150405") + "_" + hex.EncodeToString(b)
}
