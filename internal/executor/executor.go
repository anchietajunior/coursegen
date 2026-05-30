// Package executor runs the sequential production loop.
//
// For each lesson, IN ORDER, ONE AT A TIME:
//  1. build the minimal context pack (token economy)
//  2. skip if the output already exists and inputs are unchanged
//  3. render the prompt
//  4. run ONE fresh agent session in an isolated workdir
//     → a new process = a clean context (no bleed from the previous lesson)
//  5. validate + write output/lessons/module-XX/lesson-XX-YY.md
//  6. persist state + show status
//  7. move on — the next lesson gets a brand new session
//
// Nothing runs in parallel. This is intentional: predictable token spend, clean
// context per lesson, and no cross-module mixing.
package executor

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coursegen/coursegen/internal/config"
	"github.com/coursegen/coursegen/internal/contextpack"
	"github.com/coursegen/coursegen/internal/course"
	"github.com/coursegen/coursegen/internal/prompt"
	"github.com/coursegen/coursegen/internal/reporter"
	"github.com/coursegen/coursegen/internal/runner"
	"github.com/coursegen/coursegen/internal/state"
	"github.com/coursegen/coursegen/internal/tokens"
	"github.com/coursegen/coursegen/internal/validator"
)

// Executor wires together everything needed to produce lessons.
type Executor struct {
	cfg     *config.Config
	course  *course.Course
	runner  runner.Runner
	report  *reporter.Reporter
	store   *state.Store
	builder *prompt.Builder
	force   bool
	timeout time.Duration
}

// New builds an Executor (loads the prompt template).
func New(cfg *config.Config, c *course.Course, r runner.Runner, rep *reporter.Reporter,
	st *state.Store, force bool, timeoutSeconds int) (*Executor, error) {

	b, err := prompt.NewBuilder(cfg.PromptTemplatePath())
	if err != nil {
		return nil, err
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = cfg.Execution.TimeoutSeconds
	}
	return &Executor{
		cfg: cfg, course: c, runner: r, report: rep, store: st,
		builder: b, force: force, timeout: time.Duration(timeoutSeconds) * time.Second,
	}, nil
}

// GenerateLessons runs the sequential loop over lessons. onlyFailed restricts
// work to executions currently in "failed" (used by retry).
func (e *Executor) GenerateLessons(lessons []course.Lesson, run *state.Run, onlyFailed bool) error {
	total := len(lessons)
	cumulative := 0

	for i, lesson := range lessons {
		index := i + 1
		exec := e.store.UpsertExecution(run, lesson)
		if onlyFailed && exec.Status != "failed" {
			continue
		}

		e.report.LessonStart(index, total, lesson)
		e.processLesson(lesson, exec, run)

		cumulative += exec.Tokens
		run.Tokens = cumulative
		if err := e.store.Save(); err != nil {
			return err
		}

		e.report.LessonDone(index, total, lesson, exec.Status, exec.DurationMs,
			exec.Tokens, cumulative, exec.Issues)
	}

	e.finalize(run)
	return e.store.Save()
}

func (e *Executor) processLesson(lesson course.Lesson, exec *state.Execution, run *state.Run) {
	pack, err := contextpack.Build(e.cfg, e.course, lesson)
	if err != nil {
		exec.Status = "failed"
		exec.Error = err.Error()
		exec.DurationMs = 0
		return
	}

	fingerprint := pack.Fingerprint(e.builder.Fingerprint())
	outputPath := e.outputPathFor(lesson)
	exec.InputHash = fingerprint
	exec.OutputPath = e.rel(outputPath)

	// 2. Idempotent skip — the cornerstone of NOT re-spending tokens.
	if !e.force && e.store.Cached(lesson.Unit(), fingerprint, outputPath) {
		exec.Status = "skipped"
		exec.DurationMs = 0
		exec.Tokens = 0
		return
	}

	exec.Status = "running"
	exec.Attempt++

	promptText, err := e.builder.Render(pack)
	if err != nil {
		exec.Status = "failed"
		exec.Error = err.Error()
		return
	}

	workdir := e.prepareWorkdir(run, lesson, promptText)

	// 4. ONE isolated, fresh session. New process ⇒ clean context.
	res := e.runner.Run(runner.Invocation{
		Prompt: promptText, Workdir: workdir,
		OutputPath: outputPath, Timeout: e.timeout,
	})

	e.persistLogs(workdir, res)
	exec.Tokens = tokens.Estimate(promptText, res.Artifact)
	exec.DurationMs = res.Duration.Milliseconds()

	if res.OK() {
		if err := writeOutput(outputPath, res.Artifact); err != nil {
			exec.Status = "failed"
			exec.Error = err.Error()
			return
		}
		issues := validator.CheckLesson(res.Artifact)
		exec.Issues = issues
		if len(issues) == 0 || e.cfg.Execution.OnValidationFailure == "warn" {
			exec.Status = "ok"
			exec.Error = ""
			e.store.Remember(lesson.Unit(), fingerprint)
		} else {
			exec.Status = "failed"
			exec.Error = "validação: " + strings.Join(issues, "; ")
		}
	} else {
		exec.Status = "failed"
		exec.Error = res.Err
	}
}

func (e *Executor) finalize(run *state.Run) {
	failed, anyOK := 0, false
	for _, x := range run.Executions {
		switch x.Status {
		case "failed":
			failed++
		case "ok":
			anyOK = true
		}
	}
	switch {
	case failed == 0:
		run.Status = "ok"
	case anyOK:
		run.Status = "partial"
	default:
		run.Status = "failed"
	}
	run.FinishedAt = time.Now().Format(time.RFC3339)
}

func (e *Executor) outputPathFor(l course.Lesson) string {
	return filepath.Join(e.cfg.OutputPath(), "lessons", l.ModuleDir(), l.Unit()+".md")
}

// prepareWorkdir creates the isolated session dir and writes PROMPT.md.
func (e *Executor) prepareWorkdir(run *state.Run, l course.Lesson, promptText string) string {
	dir := filepath.Join(e.cfg.RunsPath(), run.ID, run.Operation, l.Unit())
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "PROMPT.md"), []byte(promptText), 0o644)
	return dir
}

func (e *Executor) persistLogs(workdir string, res runner.Result) {
	_ = os.WriteFile(filepath.Join(workdir, "stdout.txt"), []byte(res.Stdout), 0o644)
	_ = os.WriteFile(filepath.Join(workdir, "stderr.txt"), []byte(res.Stderr), 0o644)
}

func (e *Executor) rel(path string) string {
	return strings.TrimPrefix(path, e.cfg.Root+string(os.PathSeparator))
}

func writeOutput(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.TrimRight(content, "\n")+"\n"), 0o644)
}
