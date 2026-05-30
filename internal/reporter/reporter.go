// Package reporter handles all user-facing status output. In sequential mode it
// shows progress as each lesson runs, plus a final summary and the status/runs
// tables. TTY-aware: it updates a line in place when interactive, and falls
// back to plain lines in logs/CI.
package reporter

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/coursegen/coursegen/internal/course"
	"github.com/coursegen/coursegen/internal/state"
	"github.com/coursegen/coursegen/internal/tokens"
)

var symbols = map[string]string{
	"ok": "✓", "failed": "✗", "skipped": "⤼", "running": "▶", "pending": "·",
}

// Reporter writes progress to an io.Writer.
type Reporter struct {
	w       io.Writer
	tty     bool
	pending bool
}

// New returns a Reporter writing to stdout.
func New() *Reporter {
	return &Reporter{w: os.Stdout, tty: isTerminal(os.Stdout)}
}

func (r *Reporter) RunHeader(run *state.Run, total int, runner string) {
	r.line(fmt.Sprintf("CourseGen · %s · runner=%s · modo sequencial (contexto limpo a cada aula)",
		run.Operation, runner))
	noun := "aulas encontradas"
	if total == 1 {
		noun = "aula encontrada"
	}
	r.line(fmt.Sprintf("%d %s.\n", total, noun))
}

func (r *Reporter) LessonStart(index, total int, l course.Lesson) {
	head := fmt.Sprintf("  [%d/%d] %s", index, total, l.Unit())
	if r.tty {
		fmt.Fprintf(r.w, "%s  ⏳ gerando… (sessão isolada)", head)
		r.pending = true
	} else {
		r.line(head + "  ▶ gerando…")
	}
}

func (r *Reporter) LessonDone(index, total int, l course.Lesson, status string, durationMs int64, tok, cumulative int, issues []string) {
	r.clearPending()
	sym := symbols[status]
	note := ""
	switch status {
	case "skipped":
		note = "  (inalterado — 0 tokens)"
	case "failed":
		note = "  ← falhou"
	default:
		if len(issues) > 0 {
			note = "  ⚠ " + strings.Join(issues, "; ")
		}
	}
	r.line(fmt.Sprintf("  [%d/%d] %s  %s %-7s %-6s ~%s tok (acum. ~%s)%s",
		index, total, l.Unit(), sym, status, formatDuration(durationMs),
		tokens.Human(tok), tokens.Human(cumulative), note))
}

func (r *Reporter) RunSummary(run *state.Run) {
	ok, failed, skipped := countStatuses(run)
	r.line("")
	r.line(fmt.Sprintf("Resumo: %d ok · %d falhou · %d pulados · ~%s tokens · status=%s",
		ok, failed, skipped, tokens.Human(run.Tokens), strings.ToUpper(run.Status)))
	if failed > 0 {
		r.line("→ Reexecute as falhas com: coursegen tasks retry failed")
	}
}

func (r *Reporter) StatusTable(run *state.Run) {
	if run == nil {
		r.line("Nenhuma execução registrada ainda. Rode `coursegen tasks run generate-lessons`.")
		return
	}
	ok, _, skipped := countStatuses(run)
	done := ok + skipped
	r.line(fmt.Sprintf("Run %s · %s · runner=%s · status=%s · %d/%d concluídas (%d ok, %d pulados) · ~%s tokens",
		run.ID, run.Operation, run.Runner, strings.ToUpper(run.Status), done, run.Total, ok, skipped, tokens.Human(run.Tokens)))
	r.line("")
	r.line(fmt.Sprintf("%-16s %-9s %-6s %-9s %s", "UNIDADE", "STATUS", "TENT.", "DURAÇÃO", "OUTPUT/ERRO"))
	for _, e := range run.Executions {
		detail := e.OutputPath
		if e.Status == "failed" {
			detail = truncate(e.Error, 50)
		}
		r.line(fmt.Sprintf("%-16s %-9s %-6d %-9s %s",
			e.Unit, symbols[e.Status]+" "+e.Status, e.Attempt, formatDuration(e.DurationMs), detail))
	}
}

func (r *Reporter) RunsTable(runs []*state.Run) {
	if len(runs) == 0 {
		r.line("Nenhuma run registrada.")
		return
	}
	r.line(fmt.Sprintf("%-28s %-18s %-9s %-9s %-9s %s",
		"RUN_ID", "OPERAÇÃO", "RUNNER", "STATUS", "DONE/TOT", "INÍCIO"))
	for _, run := range runs {
		ok, _, skipped := countStatuses(run)
		r.line(fmt.Sprintf("%-28s %-18s %-9s %-9s %-9s %s",
			run.ID, run.Operation, run.Runner, run.Status,
			fmt.Sprintf("%d/%d", ok+skipped, run.Total), run.StartedAt))
	}
}

func (r *Reporter) Info(msg string)  { r.line(msg) }
func (r *Reporter) Warn(msg string)  { r.line("⚠ " + msg) }
func (r *Reporter) Error(msg string) { r.line("✗ " + msg) }
func (r *Reporter) Raw(msg string)   { r.line(msg) }

func (r *Reporter) clearPending() {
	if r.tty && r.pending {
		fmt.Fprint(r.w, "\r\033[K")
		r.pending = false
	}
}

func (r *Reporter) line(text string) { fmt.Fprintln(r.w, text) }

func countStatuses(run *state.Run) (ok, failed, skipped int) {
	for _, e := range run.Executions {
		switch e.Status {
		case "ok":
			ok++
		case "failed":
			failed++
		case "skipped":
			skipped++
		}
	}
	return
}

func formatDuration(ms int64) string {
	if ms <= 0 {
		return "0.0s"
	}
	s := float64(ms) / 1000.0
	if s < 60 {
		return fmt.Sprintf("%.1fs", s)
	}
	return fmt.Sprintf("%dm%ds", int(s)/60, int(s)%60)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
