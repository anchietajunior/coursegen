// Package cli is the command dispatch layer (stdlib flag only — no external CLI
// framework, so the binary stays minimal and offline-buildable).
package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coursegen/coursegen/internal/config"
	"github.com/coursegen/coursegen/internal/contextpack"
	"github.com/coursegen/coursegen/internal/course"
	"github.com/coursegen/coursegen/internal/executor"
	"github.com/coursegen/coursegen/internal/prompt"
	"github.com/coursegen/coursegen/internal/reporter"
	"github.com/coursegen/coursegen/internal/runner"
	"github.com/coursegen/coursegen/internal/state"
	"github.com/coursegen/coursegen/internal/tokens"
)

// Version of the CLI.
const Version = "0.1.0"

// Run dispatches a command and returns a process exit code.
func Run(args []string) int {
	if len(args) == 0 {
		printHelp()
		return 0
	}
	switch args[0] {
	case "init":
		return cmdInit(args[1:])
	case "doctor":
		return cmdDoctor()
	case "version", "--version", "-v":
		fmt.Println("coursegen " + Version)
		return 0
	case "status":
		return cmdStatus()
	case "readiness":
		return cmdReadiness(args[1:])
	case "tasks":
		return cmdTasks(args[1:])
	case "runs":
		return cmdRuns(args[1:])
	case "help", "-h", "--help":
		printHelp()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "comando desconhecido: %s\n\n", args[0])
		printHelp()
		return 1
	}
}

// --- project plumbing -------------------------------------------------------

type project struct {
	cfg    *config.Config
	store  *state.Store
	course *course.Course
	rep    *reporter.Reporter
}

func loadProject() (*project, error) {
	root, _ := os.Getwd()
	cfg, err := config.Load(root)
	if err != nil {
		return nil, err
	}
	st, err := state.Load(cfg.StatePath())
	if err != nil {
		return nil, err
	}
	return &project{cfg: cfg, store: st, course: course.New(cfg), rep: reporter.New()}, nil
}

func fail(err error) int {
	fmt.Fprintf(os.Stderr, "✗ %s\n", err.Error())
	return 1
}

// --- init -------------------------------------------------------------------

func cmdInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	name := fs.String("name", "", "Nome do curso")
	language := fs.String("language", "pt-BR", "Idioma dos artefatos")
	rnr := fs.String("runner", "claude", "Runner default")
	force := fs.Bool("force", false, "Sobrescreve coursegen.yml existente")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	root, _ := os.Getwd()
	cfgPath := filepath.Join(root, "coursegen.yml")
	if _, err := os.Stat(cfgPath); err == nil && !*force {
		fmt.Fprintln(os.Stderr, "coursegen.yml já existe. Use --force para sobrescrever.")
		return 1
	}

	courseName := *name
	if courseName == "" {
		courseName = filepath.Base(root)
	}
	if err := os.WriteFile(cfgPath, []byte(configYAML(courseName, *language, *rnr)), 0o644); err != nil {
		return fail(err)
	}

	dirs := []string{
		".coursegen", ".coursegen/logs", ".coursegen/runs",
		"output/lessons", "output/exercises", "output/projects", "output/slides", "output/reviews",
		"coursegen/prompts",
	}
	for _, d := range dirs {
		_ = os.MkdirAll(filepath.Join(root, d), 0o755)
	}
	_ = os.WriteFile(filepath.Join(root, "coursegen/prompts/generate-lesson.tmpl"),
		[]byte(prompt.DefaultTemplate()), 0o644)

	fmt.Println("✓ coursegen.yml criado")
	fmt.Println("✓ .coursegen/ e output/ criados")
	fmt.Println("✓ template de prompt em coursegen/prompts/generate-lesson.tmpl")
	fmt.Println("\nPróximo passo: garanta docs/ aprovado e rode `coursegen readiness check`.")
	return 0
}

// --- doctor -----------------------------------------------------------------

func cmdDoctor() int {
	root, _ := os.Getwd()
	for _, name := range runner.Names() {
		if name == "mock" {
			continue
		}
		r, err := runner.Resolve(name, root)
		if err != nil {
			continue
		}
		if r.Available() {
			v := r.Version()
			if v == "" {
				v = "ok"
			}
			fmt.Printf("✓ %-10s %s\n", name, v)
		} else {
			fmt.Printf("✗ %-10s não encontrado no PATH\n", name)
		}
	}
	fmt.Println("✓ mock       sempre disponível (gera conteúdo de teste, 0 tokens)")
	return 0
}

// --- status -----------------------------------------------------------------

func cmdStatus() int {
	p, err := loadProject()
	if err != nil {
		return fail(err)
	}
	p.rep.StatusTable(p.store.LatestRun())
	return 0
}

// --- readiness --------------------------------------------------------------

func cmdReadiness(args []string) int {
	if len(args) == 0 || args[0] != "check" {
		fmt.Fprintln(os.Stderr, "uso: coursegen readiness check")
		return 1
	}
	p, err := loadProject()
	if err != nil {
		return fail(err)
	}
	r := course.CheckReadiness(p.cfg)
	if !r.Exists {
		fmt.Fprintf(os.Stderr, "✗ readiness não encontrado: %s\n", rel(p.cfg, r.Source))
		return 2
	}
	mark := "✗"
	if r.Approved {
		mark = "✓"
	}
	fmt.Printf("%s Readiness: %s  (%s)\n", mark, r.Verdict, rel(p.cfg, r.Source))
	fmt.Printf("  Bloqueadores: %s · Avisos: %s\n", intOr(r.Blockers), intOr(r.Warnings))
	if r.Approved {
		fmt.Println("  → Liberado para produção pela CLI.")
		return 0
	}
	fmt.Println("  → Bloqueado. Corrija e rode a skill course-readiness.")
	return 1
}

// --- tasks ------------------------------------------------------------------

func cmdTasks(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "uso: coursegen tasks <list|run|status|retry>")
		return 1
	}
	switch args[0] {
	case "list":
		return tasksList()
	case "run":
		return tasksRun(args[1:])
	case "status":
		return cmdStatus()
	case "retry":
		return tasksRetry(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "subcomando de tasks desconhecido: %s\n", args[0])
		return 1
	}
}

func tasksList() int {
	rows := [][]string{
		{"generate-lessons", "lesson", "sim", "Gera a aula completa a partir da lesson spec"},
		{"generate-exercises", "lesson", "sim", "Gera exercícios da aula (roadmap v0.3)"},
		{"review-lessons", "lesson", "não", "Revisa as aulas geradas (roadmap v0.3)"},
		{"generate-slides", "lesson", "sim", "Gera o deck de slides da aula (roadmap v0.3)"},
	}
	fmt.Printf("%-18s %-8s %-10s %s\n", "TASK", "UNIDADE", "READINESS", "DESCRIÇÃO")
	for _, r := range rows {
		fmt.Printf("%-18s %-8s %-10s %s\n", r[0], r[1], r[2], r[3])
	}
	return 0
}

func tasksRun(args []string) int {
	var task string
	rest := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		task, rest = args[0], args[1:]
	}

	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	rnr := fs.String("runner", "", "Runner a usar (claude, codex, mock, ...)")
	parallel := fs.Int("parallel", 1, "MVP roda sempre sequencial")
	lesson := fs.String("lesson", "", "Filtra uma aula (ex.: lesson-01-01 ou 01-01)")
	moduleF := fs.String("module", "", "Filtra um módulo (ex.: 01 ou module-01)")
	force := fs.Bool("force", false, "Regera mesmo se nada mudou")
	timeout := fs.Int("timeout", 0, "Timeout por sessão (segundos)")
	skipReadiness := fs.Bool("skip-readiness", false, "Ignora o gate (escape hatch)")
	dryRun := fs.Bool("dry-run", false, "Planeja sem executar")
	if err := fs.Parse(rest); err != nil {
		return 1
	}

	if task != "generate-lessons" {
		fmt.Fprintf(os.Stderr, "✗ task '%s' ainda não implementada no MVP. Disponível: generate-lessons.\n", task)
		return 1
	}

	p, err := loadProject()
	if err != nil {
		return fail(err)
	}
	if *parallel > 1 {
		p.rep.Warn(fmt.Sprintf("MVP roda em modo sequencial; ignorando --parallel=%d.", *parallel))
	}

	if code := gate(p, !*skipReadiness); code != 0 {
		return code
	}

	lessons, err := p.course.Lessons(*moduleF, *lesson)
	if err != nil {
		return fail(err)
	}
	if len(lessons) == 0 {
		return fail(fmt.Errorf("nenhuma lesson spec encontrada em %s/05-lesson-specs/ (esperado: module-XX/lesson-XX-YY.md)", p.cfg.DocsPath()))
	}

	runnerName := firstNonEmpty(*rnr, p.cfg.Runners.Default)
	r, err := resolveRunner(runnerName, p.cfg.Root)
	if err != nil {
		return fail(err)
	}

	if *dryRun {
		return printPlan(p, lessons, r)
	}

	run := p.store.CreateRun("generate-lessons", r.Name(),
		"coursegen tasks run generate-lessons", len(lessons), Version)
	if err := p.store.Save(); err != nil {
		return fail(err)
	}

	p.rep.RunHeader(run, len(lessons), r.Name())
	ex, err := executor.New(p.cfg, p.course, r, p.rep, p.store, *force, *timeout)
	if err != nil {
		return fail(err)
	}
	if err := ex.GenerateLessons(lessons, run, false); err != nil {
		return fail(err)
	}
	p.rep.RunSummary(run)
	return 0
}

func tasksRetry(args []string) int {
	which := "failed"
	rest := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		which, rest = args[0], args[1:]
	}
	if which != "failed" {
		fmt.Fprintln(os.Stderr, "uso: coursegen tasks retry failed")
		return 1
	}

	fs := flag.NewFlagSet("retry", flag.ContinueOnError)
	rnr := fs.String("runner", "", "Runner a usar no retry")
	if err := fs.Parse(rest); err != nil {
		return 1
	}

	p, err := loadProject()
	if err != nil {
		return fail(err)
	}
	run := p.store.LatestRun()
	if run == nil {
		return fail(fmt.Errorf("nenhuma run para reexecutar"))
	}

	var lessons []course.Lesson
	for _, e := range run.Executions {
		if e.Status == "failed" {
			lessons = append(lessons, course.Lesson{
				ModuleID: e.Module, LessonID: e.Lesson,
				SpecPath: p.course.LessonSpecPath(e.Module, e.Lesson),
			})
		}
	}
	if len(lessons) == 0 {
		p.rep.Info(fmt.Sprintf("Nenhuma execução falha na última run (%s). Nada a fazer.", run.ID))
		return 0
	}

	runnerName := firstNonEmpty(*rnr, run.Runner)
	r, err := resolveRunner(runnerName, p.cfg.Root)
	if err != nil {
		return fail(err)
	}
	run.Runner = r.Name()

	p.rep.Info(fmt.Sprintf("Reexecutando %d execução(ões) falha(s) de %s…\n", len(lessons), run.ID))
	p.rep.RunHeader(run, len(lessons), r.Name())
	ex, err := executor.New(p.cfg, p.course, r, p.rep, p.store, false, 0)
	if err != nil {
		return fail(err)
	}
	if err := ex.GenerateLessons(lessons, run, true); err != nil {
		return fail(err)
	}
	p.rep.RunSummary(run)
	return 0
}

// --- runs -------------------------------------------------------------------

func cmdRuns(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "uso: coursegen runs <list|show RUN_ID>")
		return 1
	}
	p, err := loadProject()
	if err != nil {
		return fail(err)
	}
	switch args[0] {
	case "list":
		p.rep.RunsTable(p.store.Runs)
		return 0
	case "show":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "uso: coursegen runs show RUN_ID")
			return 1
		}
		run := p.store.FindRun(args[1])
		if run == nil {
			return fail(fmt.Errorf("run não encontrada: %s", args[1]))
		}
		p.rep.StatusTable(run)
		fmt.Printf("\ncomando: %s\n", run.Cmd)
		fmt.Printf("início:  %s  · fim: %s\n", run.StartedAt, orDash(run.FinishedAt))
		fmt.Printf("logs:    %s/%s/\n", p.cfg.RunsPath(), run.ID)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "subcomando de runs desconhecido: %s\n", args[0])
		return 1
	}
}

// --- helpers ----------------------------------------------------------------

func gate(p *project, required bool) int {
	r := course.CheckReadiness(p.cfg)
	if r.Approved {
		p.rep.Info("✓ Readiness: APROVADO — liberado para produção.\n")
		return 0
	}
	msg := fmt.Sprintf("curso NÃO está APROVADO (veredito: %s). Rode a skill course-readiness ou use --skip-readiness.", r.Verdict)
	if required && p.cfg.ReadinessRequired() {
		fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
		return 1
	}
	p.rep.Warn(msg + " (prosseguindo por --skip-readiness)\n")
	return 0
}

func resolveRunner(name, root string) (runner.Runner, error) {
	r, err := runner.Resolve(name, root)
	if err != nil {
		return nil, err
	}
	if !r.Available() {
		return nil, fmt.Errorf("runner '%s' indisponível (binário no PATH?). Rode `coursegen doctor`", name)
	}
	return r, nil
}

func printPlan(p *project, lessons []course.Lesson, r runner.Runner) int {
	p.rep.Info(fmt.Sprintf("Plano (run NÃO criada) — runner=%s, modo sequencial:", r.Name()))
	total := 0
	for i, l := range lessons {
		pack, err := contextpack.Build(p.cfg, p.course, l)
		if err != nil {
			return fail(err)
		}
		t := pack.InputTokens()
		total += t
		p.rep.Info(fmt.Sprintf("  %2d. %-14s  ~%s tokens de contexto → output/lessons/%s/%s.md",
			i+1, l.Unit(), tokens.Human(t), l.ModuleDir(), l.Unit()))
	}
	p.rep.Info(fmt.Sprintf("\n%d aulas · ~%s tokens de contexto no total (só entrada; saída não estimada).",
		len(lessons), tokens.Human(total)))
	if total > p.cfg.MaxTokensEstimate() {
		p.rep.Warn(fmt.Sprintf("contexto acumulado acima do limite de aviso (%s).", tokens.Human(p.cfg.MaxTokensEstimate())))
	}
	return 0
}

func configYAML(name, language, runnerName string) string {
	return fmt.Sprintf(`version: 1

course:
  name: "%s"
  language: %s

paths:
  docs: docs
  output: output
  state: .coursegen/state.json
  logs: .coursegen/logs
  runs: .coursegen/runs

readiness:
  required: true
  source: docs/06-course-readiness-checklist.md
  approved_marker: "APROVADO"

runners:
  default: %s

execution:
  timeout_seconds: 900
  on_validation_failure: warn      # warn | fail

# Context pack compartilhado por TODA aula (o mínimo comum).
# A module spec e a lesson spec daquela aula são adicionadas automaticamente.
context:
  shared:
    - docs/01-course-prd.md
    - docs/02-market-research.md
    - docs/03-learning-architecture.md
  max_tokens_estimate: 120000
`, name, language, runnerName)
}

func printHelp() {
	fmt.Print(`CourseGen — orquestra a produção de aulas com agentes de IA (uma aula por sessão isolada).

Uso:
  coursegen <comando> [opções]

Comandos:
  init                          Inicializa o projeto no diretório atual
  readiness check               Verifica se o curso está APROVADO para produção
  doctor                        Verifica disponibilidade dos runners
  tasks list                    Lista as tasks disponíveis
  tasks run generate-lessons    Gera as aulas (SEQUENCIAL, contexto limpo por aula)
  tasks status                  Status da última execução
  tasks retry failed            Reexecuta só as aulas que falharam
  runs list                     Lista as runs
  runs show RUN_ID              Detalhes de uma run
  status                        Atalho para ` + "`tasks status`" + `
  version                       Versão

Opções de ` + "`tasks run generate-lessons`" + `:
  --runner NAME     claude | codex | gemini | cursor | opencode | mock
  --lesson ID       Filtra uma aula (lesson-01-01)
  --module ID       Filtra um módulo (01)
  --force           Regera mesmo se nada mudou
  --timeout S       Timeout por sessão (segundos)
  --skip-readiness  Ignora o gate
  --dry-run         Planeja e estima tokens, sem executar
`)
}

func rel(cfg *config.Config, path string) string {
	return strings.TrimPrefix(path, cfg.Root+string(os.PathSeparator))
}

func intOr(p *int) string {
	if p == nil {
		return "?"
	}
	return fmt.Sprintf("%d", *p)
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
