// Package cli is the command dispatch layer (stdlib flag only — no external CLI
// framework, so the binary stays minimal and offline-buildable).
//
// The surface is verb-first on purpose: you GENERATE lessons / REVIEW lessons.
// "lesson" is the artifact; the verb is the operation. There is deliberately no
// generic "task" noun in what the user types, to avoid confusing the operation
// with the content it produces.
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anchietajunior/coursegen/internal/config"
	"github.com/anchietajunior/coursegen/internal/contextpack"
	"github.com/anchietajunior/coursegen/internal/course"
	"github.com/anchietajunior/coursegen/internal/executor"
	"github.com/anchietajunior/coursegen/internal/prompt"
	"github.com/anchietajunior/coursegen/internal/reporter"
	"github.com/anchietajunior/coursegen/internal/runner"
	"github.com/anchietajunior/coursegen/internal/state"
	"github.com/anchietajunior/coursegen/internal/tokens"
)

// Version of the CLI.
const Version = "0.1.2"

// Run dispatches a command and returns a process exit code.
func Run(args []string) int {
	if len(args) == 0 {
		printHelp()
		return 0
	}
	switch args[0] {
	case "init":
		return cmdInit(args[1:])
	case "setup":
		return cmdSetup(args[1:])
	case "doctor":
		return cmdDoctor()
	case "version", "--version", "-v":
		fmt.Println("coursegen " + Version)
		return 0
	case "readiness":
		return cmdReadiness(args[1:])
	case "generate", "gen":
		return cmdGenerate(args[1:])
	case "review":
		return cmdReview(args[1:])
	case "status":
		return cmdStatus()
	case "retry":
		return cmdRetry(args[1:])
	case "list":
		return cmdList()
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
	force := fs.Bool("force", false, "Sobrescreve coursegen.json existente")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	root, _ := os.Getwd()
	cfgPath := filepath.Join(root, "coursegen.json")
	if _, err := os.Stat(cfgPath); err == nil && !*force {
		fmt.Fprintln(os.Stderr, "coursegen.json já existe. Use --force para sobrescrever.")
		return 1
	}

	courseName := *name
	if courseName == "" {
		courseName = filepath.Base(root)
	}
	if err := os.WriteFile(cfgPath, []byte(configJSON(courseName, *language, *rnr)), 0o644); err != nil {
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

	fmt.Println("✓ coursegen.json criado")
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

// --- generate / review ------------------------------------------------------

func cmdGenerate(args []string) int {
	kind, rest := splitKind(args)
	switch kind {
	case "lessons":
		return runGenerateLessons(rest)
	case "":
		fmt.Fprintln(os.Stderr, "uso: coursegen generate <lessons|exercises|slides|projects>")
		return 1
	case "exercises", "slides", "projects":
		fmt.Fprintf(os.Stderr, "✗ `generate %s` é roadmap (v0.3) e ainda não foi implementado.\n", kind)
		return 1
	default:
		fmt.Fprintf(os.Stderr, "✗ alvo desconhecido para generate: '%s'.\n", kind)
		return 1
	}
}

func cmdReview(args []string) int {
	kind, _ := splitKind(args)
	switch kind {
	case "lessons":
		fmt.Fprintln(os.Stderr, "✗ `review lessons` é roadmap (v0.3) e ainda não foi implementado.")
		return 1
	case "":
		fmt.Fprintln(os.Stderr, "uso: coursegen review <lessons>")
		return 1
	default:
		fmt.Fprintf(os.Stderr, "✗ alvo desconhecido para review: '%s'.\n", kind)
		return 1
	}
}

// runGenerateLessons is the sequential production loop entrypoint.
func runGenerateLessons(rest []string) int {
	fs := flag.NewFlagSet("generate lessons", flag.ContinueOnError)
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

	run := p.store.CreateRun(operationGenerateLessons, r.Name(),
		"coursegen generate lessons", len(lessons), Version)
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

// --- retry ------------------------------------------------------------------

func cmdRetry(args []string) int {
	which, rest := splitKind(args)
	if which == "" {
		which = "failed"
	}
	if which != "failed" {
		fmt.Fprintln(os.Stderr, "uso: coursegen retry failed")
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

// --- list (generators) ------------------------------------------------------

func cmdList() int {
	rows := [][]string{
		{"generate lessons", "lesson", "sim", "Gera a aula completa a partir da lesson spec"},
		{"generate exercises", "lesson", "sim", "Gera exercícios da aula (roadmap v0.3)"},
		{"generate slides", "lesson", "sim", "Gera o deck de slides da aula (roadmap v0.3)"},
		{"review lessons", "lesson", "não", "Revisa as aulas geradas (roadmap v0.3)"},
	}
	fmt.Printf("%-20s %-8s %-10s %s\n", "COMANDO", "UNIDADE", "READINESS", "DESCRIÇÃO")
	for _, r := range rows {
		fmt.Printf("%-20s %-8s %-10s %s\n", r[0], r[1], r[2], r[3])
	}
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

// operationGenerateLessons is the internal slug for the lesson generation
// operation (used in state and run workdirs). It is an OPERATION identifier,
// never a "lesson" — the two are distinct concepts.
const operationGenerateLessons = "generate-lessons"

// splitKind pulls the leading non-flag positional (the target, e.g. "lessons")
// off args, returning it plus the remaining flag args.
func splitKind(args []string) (kind string, rest []string) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		return args[0], args[1:]
	}
	return "", args
}

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

// configJSON renders the default coursegen.json. Fields are documented in
// README.md / DESIGN.md (JSON has no comments). The context.shared list is the
// minimal pack shared by EVERY lesson; the module spec and the lesson spec of
// each lesson are added automatically.
func configJSON(name, language, runnerName string) string {
	return fmt.Sprintf(`{
  "version": 1,
  "course": {
    "name": %s,
    "language": %s
  },
  "paths": {
    "docs": "docs",
    "output": "output",
    "state": ".coursegen/state.json",
    "logs": ".coursegen/logs",
    "runs": ".coursegen/runs"
  },
  "readiness": {
    "required": true,
    "source": "docs/06-course-readiness-checklist.md",
    "approved_marker": "APROVADO"
  },
  "runners": {
    "default": %s
  },
  "execution": {
    "timeout_seconds": 900,
    "on_validation_failure": "warn"
  },
  "context": {
    "shared": [
      "docs/01-course-prd.md",
      "docs/02-market-research.md",
      "docs/03-learning-architecture.md"
    ],
    "max_tokens_estimate": 120000
  }
}
`, jsonString(name), jsonString(language), jsonString(runnerName))
}

// jsonString safely encodes s as a JSON string literal (handles quotes, etc.).
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func printHelp() {
	fmt.Print(`CourseGen — orquestra a produção de aulas com agentes de IA (uma aula por sessão isolada).

Uso:
  coursegen <comando> [opções]

Comandos:
  init                       Inicializa o projeto no diretório atual
  setup                      Instala as skills de planejamento no seu agente
  readiness check            Verifica se o curso está APROVADO para produção
  doctor                     Verifica disponibilidade dos runners
  list                       Lista os geradores disponíveis
  generate lessons           Gera as aulas (SEQUENCIAL, contexto limpo por aula)
  review lessons             Revisa as aulas geradas (roadmap v0.3)
  status                     Status da última execução
  retry failed               Reexecuta só as aulas que falharam
  runs list                  Lista as runs
  runs show RUN_ID           Detalhes de uma run
  version                    Versão

Nota de vocabulário: você GERA/REVISA *lessons* (o artefato). O verbo é a
operação; "lesson" é o conteúdo produzido — não há um conceito genérico de
"task" na linha de comando.

Opções de ` + "`generate lessons`" + `:
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
