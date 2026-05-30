# CourseGen CLI — Documento de Arquitetura

> Ferramenta para **orquestrar a produção em escala** de aulas, exercícios,
> projetos e slides de cursos, usando agentes externos (Claude Code, Codex,
> Gemini CLI, Cursor Agent, OpenCode).
>
> **Regra central: o agente DEFINE, a CLI ESCALA.**
> A definição do curso já aconteceu antes, dentro de um agente interativo, via o
> [pacote de skills](skills/README.md). A CLI **lê** os artefatos aprovados e
> **executa** a produção — ela nunca define o curso do zero.

> **Nota de implementação (atualizada).** A CLI é implementada em **Go**
> (`cmd/coursegen`, `internal/`) — escolhida por distribuir como **binário único
> sem runtime**. Este documento foi escrito originalmente assumindo Ruby; os
> conceitos (runner, task, workflow, context pack, fluxo, prompt, riscos,
> roadmap) são **agnósticos de linguagem** e permanecem válidos. Trechos de
> código em Ruby ao longo do texto são **ilustrativos do design** — a fonte da
> verdade da implementação é o código Go. As diferenças concretas estão anotadas
> nas seções de Arquitetura, Modelo de estado e ADRs.

---

## Índice

1. [Visão da CLI](#1-visão-da-cli)
2. [Arquitetura](#2-arquitetura)
3. [Estrutura de pastas](#3-estrutura-de-pastas)
4. [Comandos](#4-comandos)
5. [Formato dos YAMLs](#5-formato-dos-yamls)
6. [Interface dos runners](#6-interface-dos-runners)
7. [Fluxo de execução](#7-fluxo-de-execução)
8. [Prompt templates](#8-prompt-templates)
9. [Modelo de estado](#9-modelo-de-estado)
10. [MVP](#10-mvp)
11. [Roadmap](#11-roadmap)
12. [Riscos técnicos](#12-riscos-técnicos)
13. [Decisões arquiteturais (ADRs)](#13-decisões-arquiteturais-adrs)
14. [Exemplo completo de execução](#14-exemplo-completo-de-execução)

---

## 1. Visão da CLI

CourseGen é um **orquestrador de produção determinístico** sobre agentes de IA
não-determinísticos. Ele resolve um problema específico: gerar um curso inteiro
(dezenas a centenas de aulas) **sem** depender de uma única sessão gigante de
agente, que estouraria contexto, misturaria módulos e seria impossível de
retomar após uma falha.

A filosofia é a de um **build system para conteúdo educacional** (pense em
`make`, `bazel` ou um CI runner):

| Build system de software | CourseGen |
|---|---|
| Código-fonte | `docs/` (specs aprovadas) |
| Target / artefato | Aula, exercício, slide, projeto gerados |
| Regra de build | Task (`generate-lessons`) |
| Pipeline | Workflow (`production`, `full-build`) |
| Compilador | Runner (claude, codex, …) |
| Unidade de compilação isolada | Sessão por aula + context pack mínimo |
| Cache / incremental | `input_hash` + estado (JSON no MVP, SQLite no roadmap) |
| Retry de job que falhou | `tasks retry failed` |

### Princípios de design

1. **Isolamento por unidade.** Cada aula é gerada em uma sessão de agente
   isolada, num diretório de trabalho próprio, com um context pack mínimo.
   Nunca uma sessão vê o conteúdo de outra aula.
2. **Context pack mínimo.** Para gerar uma aula, o agente recebe apenas: PRD,
   market research, learning architecture, a **module spec correspondente** e a
   **lesson spec correspondente** — nunca todas as aulas.
3. **Determinismo na orquestração, criatividade no agente.** A CLI controla
   ordem, isolamento, retry, estado e validação. O agente só escreve conteúdo.
4. **Gate de readiness obrigatório.** Tasks de produção recusam-se a rodar se
   `docs/06-course-readiness-checklist.md` não estiver **APROVADO**.
5. **Idempotência e retomada.** Reexecutar é seguro; o que já foi gerado e não
   mudou é pulado. Falhas são retentáveis sem reprocessar o curso inteiro.
6. **Runner-agnóstico.** Trocar de agente é trocar um arquivo YAML, não código.
7. **Tudo auditável.** Cada execução registra prompt, contexto, stdout/stderr,
   exit code, duração e hash de entrada.

### O que a CLI faz

- Lê documentos aprovados · valida readiness · cria uma task por aula · monta o
  context pack · executa cada aula em sessão isolada · suporta múltiplos runners
  · gera aulas, exercícios, projetos e slides · roda reviews · salva logs ·
  permite retry · mostra status.

### O que a CLI **não** faz

- Não faz discovery, PRD, arquitetura ou specs (isso é o pacote de skills).
- Não cria módulos ou aulas novas; só produz o que as specs definem.
- Não roda se o curso não foi aprovado no readiness check.

---

## 2. Arquitetura

### Stack

Stack **implementada (Go)** — entre parênteses, o equivalente do design original
em Ruby:

| Camada | Tecnologia (Go) | Papel |
|---|---|---|
| CLI / comandos | **stdlib `flag`** + dispatch próprio (≈ Thor) | Comandos, subcomandos, flags, help |
| Configuração | **`gopkg.in/yaml.v3`** (≈ Psych) | `coursegen.yml`, runners (override) |
| Estado | **JSON** (`encoding/json`) — MVP; SQLite no roadmap | Runs, execuções, cache de idempotência |
| Execução externa | **`os/exec`** + `context` (≈ Open3) | Spawn de sessões de agente, captura de I/O, timeout |
| Templates de prompt | **`text/template`** embarcado (`//go:embed`) (≈ ERB) | Montagem do prompt a partir do context pack |
| Paralelismo | sequencial no MVP; **goroutines + worker pool** no roadmap (≈ threads) | Fan-out de aulas (I/O-bound) |
| Artefatos | Sistema de arquivos | `output/`, `.coursegen/runs/`, logs |

> **Dependências:** uma só (`yaml.v3`). Tudo o mais é stdlib → compila offline e
> gera **um binário único, estático (CGO off), sem runtime**, cross-compilado
> para macOS/Linux/Windows (`make release`). Esse era o principal ponto fraco do
> Ruby para distribuição.
>
> Por que paralelismo por goroutines (roadmap) e não processos no orquestrador? O
> trabalho pesado roda em **subprocessos externos** (o agente); o orquestrador
> fica bloqueado em I/O. Goroutines + um worker pool com semáforo dão fan-out de
> espera barato, com um único processo dono do estado (escritas serializadas).

### Componentes (domínio)

```
CourseGen::CLI                  # Thor — roteia comandos
├── Config                      # carrega/valida coursegen.yml + tasks + runners
├── Course                      # representa o projeto: docs, módulos, aulas
│   ├── Readiness               # parser/gate do 06-checklist
│   ├── ModuleSpec / LessonSpec # descoberta e parsing das specs
├── Task                        # definição declarativa (YAML) de uma task
│   └── TaskUnit                # uma unidade executável (1 aula, 1 módulo, curso)
├── Workflow                    # sequência/grafo de tasks
├── ContextPack                 # monta o conjunto mínimo de arquivos por unidade
├── PromptBuilder               # ERB: context pack + instruções → prompt final
├── Runner (interface)          # claude / codex / gemini / cursor / opencode
│   └── RunnerRegistry          # resolve runner por nome a partir de runners/*.yml
├── Executor                    # loop de execução: isolamento, timeout, paralelismo
├── Validator                   # checa o artefato contra os critérios de aceite
├── State (Store)               # repositório SQLite (runs, execs, artifacts, events)
├── Logger                      # logs estruturados por execução
└── Reporters                   # status/runs (tabela, --json, --watch)
```

### Diagrama de fluxo (alto nível)

```
                ┌────────────────────────────────────────────┐
 coursegen ─────▶ CLI (Thor)                                  │
                └───────────────┬────────────────────────────┘
                                │
                   ┌────────────▼────────────┐
                   │ Config + Course loader   │  lê coursegen.yml, docs/
                   └────────────┬────────────┘
                                │
                   ┌────────────▼────────────┐
                   │ Readiness gate           │  06-checklist == APROVADO?
                   └────────────┬────────────┘ (senão: aborta)
                                │
                   ┌────────────▼────────────┐
                   │ Planner                  │  descobre unidades (aulas)
                   │  → cria run + execs       │  grava estado (pending)
                   └────────────┬────────────┘
                                │  Parallel.each (in_threads: N)
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│ ContextPack  │       │ ContextPack  │       │ ContextPack  │  (mínimo, isolado)
│ PromptBuilder│       │ PromptBuilder│       │ PromptBuilder│
│ Runner(Open3)│       │ Runner(Open3)│       │ Runner(Open3)│  sessão isolada
│ Validator    │       │ Validator    │       │ Validator    │
│ → output/    │       │ → output/    │       │ → output/    │
│ → State      │       │ → State      │       │ → State      │  (mutex no writer)
└──────────────┘       └──────────────┘       └──────────────┘
        └───────────────────────┼───────────────────────┘
                   ┌────────────▼────────────┐
                   │ Run summary + Reporter   │
                   └─────────────────────────┘
```

---

## 3. Estrutura de pastas

### 3.1 Projeto do curso (alvo de runtime — o que a CLI opera)

```
course/
├── coursegen.yml                         # configuração do projeto
├── docs/                                 # ENTRADA (gerada pelas skills, read-only p/ a CLI)
│   ├── 00-course-discovery.md
│   ├── 01-course-prd.md
│   ├── 02-market-research.md
│   ├── 03-learning-architecture.md
│   ├── 04-module-specs/
│   │   ├── module-01.md
│   │   └── module-02.md
│   ├── 05-lesson-specs/
│   │   ├── module-01/
│   │   │   ├── lesson-01-01.md
│   │   │   └── lesson-01-02.md
│   │   └── module-02/
│   │       └── lesson-02-01.md
│   └── 06-course-readiness-checklist.md  # GATE
├── .coursegen/                           # estado interno (efêmero/auditável)
│   ├── state.sqlite3                      # estado canônico
│   ├── state.sqlite3-wal                  # WAL
│   ├── config.lock.json                   # snapshot de config resolvida por run
│   ├── logs/
│   │   └── run_20260530_153000_ab12/
│   │       └── lesson-01-01.log
│   └── runs/                              # workdirs isolados por execução
│       └── run_20260530_153000_ab12/
│           └── generate-lessons/
│               └── lesson-01-01/          # sessão isolada
│                   ├── PROMPT.md          # prompt montado
│                   ├── context/           # context pack copiado (opcional)
│                   ├── stdout.txt
│                   └── stderr.txt
└── output/                               # SAÍDA (artefatos produzidos)
    ├── lessons/
    │   └── module-01/
    │       ├── lesson-01-01.md
    │       └── lesson-01-02.md
    ├── exercises/
    │   └── module-01/
    │       └── lesson-01-01.md
    ├── projects/
    │   └── final-project.md
    ├── slides/
    │   └── module-01/
    │       └── lesson-01-01.md
    └── reviews/
        └── lessons/
            └── module-01/
                └── lesson-01-01.review.md
```

### 3.2 Repositório do gem (o tool em si)

```
coursegen/                                # este repositório
├── DESIGN.md                             # este documento
├── coursegen.gemspec
├── Gemfile
├── bin/
│   └── coursegen                         # executável
├── lib/
│   ├── coursegen.rb
│   └── coursegen/
│       ├── cli.rb                        # Thor root
│       ├── commands/                     # init, readiness, tasks, runs
│       ├── config.rb
│       ├── course/                       # course, readiness, module_spec, lesson_spec
│       ├── task.rb  workflow.rb
│       ├── context_pack.rb  prompt_builder.rb
│       ├── runners/                      # base.rb + cli_runner.rb
│       ├── executor.rb  validator.rb
│       ├── state/                        # store.rb + migrations/
│       ├── reporters/
│       └── defaults/                     # configs embarcadas (overridáveis)
│           ├── tasks/
│           │   ├── generate-lessons.yml
│           │   ├── review-lessons.yml
│           │   ├── generate-exercises.yml
│           │   ├── generate-slides.yml
│           │   ├── generate-projects.yml
│           │   └── package-course.yml
│           ├── workflows/
│           │   ├── production.yml  review.yml  slides.yml  full-build.yml
│           ├── runners/
│           │   ├── claude.yml  codex.yml  gemini.yml  cursor.yml  opencode.yml
│           └── prompts/
│               ├── generate-lesson.md.erb
│               ├── review-lesson.md.erb
│               ├── generate-exercises.md.erb
│               └── generate-slides.md.erb
└── spec/
```

> **Resolução de configuração (override em camadas):** padrões embarcados em
> `lib/coursegen/defaults/` ← sobrescritos por `coursegen/{tasks,runners,prompts,workflows}/`
> dentro do projeto do curso ← sobrescritos por flags da linha de comando.
> Isso permite ao instrutor customizar um prompt ou runner sem tocar no gem.

---

## 4. Comandos

> **Atualização — superfície verbo-first.** A CLI implementada **não** expõe um
> substantivo genérico `tasks`, para não confundir a *operação* com a *lesson*
> (o artefato). O design original usava `coursegen tasks run <task>`; a forma
> implementada é verbo-first. Mapa:
>
> | Design original | Implementado |
> |---|---|
> | `coursegen tasks run generate-lessons` | `coursegen generate lessons` |
> | `coursegen tasks run review-lessons` | `coursegen review lessons` |
> | `coursegen tasks run generate-slides` | `coursegen generate slides` |
> | `coursegen tasks list` | `coursegen list` |
> | `coursegen tasks status` | `coursegen status` |
> | `coursegen tasks retry failed` | `coursegen retry failed` |
>
> Internamente, o identificador da operação (`generate-lessons`) é gravado no
> estado como **`operation`** — nunca como "task" nem "lesson". Os exemplos
> abaixo mantêm a forma `tasks run` por serem do design original; leia-os pelo
> mapa acima.

Árvore de comandos (Thor):

```
coursegen
├── init
├── readiness check
├── tasks list
├── tasks run <TASK> [opções]
├── tasks status
├── tasks retry <failed|all|EXEC_ID>
├── runs list
├── runs show <RUN_ID>
├── workflow list
├── workflow run <WORKFLOW>
└── doctor                 # healthcheck dos runners (extra)
```

### 4.1 `coursegen init`

Faz scaffold da estrutura do projeto, gera `coursegen.yml`, cria `.coursegen/`,
inicializa o SQLite (migrations) e copia os defaults de tasks/runners/prompts
para referência.

```
coursegen init [--name "Nome do curso"] [--language pt-BR] [--runner claude] [--force]
```

| Flag | Default | Descrição |
|---|---|---|
| `--name` | (pergunta) | Nome do curso |
| `--language` | `pt-BR` | Idioma dos artefatos |
| `--runner` | `claude` | Runner default |
| `--force` | `false` | Sobrescreve config existente |

### 4.2 `coursegen readiness check`

Lê `docs/06-course-readiness-checklist.md`, extrai o veredito e os bloqueadores.

```
coursegen readiness check [--json] [--strict]
```

- **Exit code 0** se APROVADO; **≠0** se REPROVADO ou ausente.
- `--strict`: além do marcador, valida a presença mínima dos artefatos.
- É o gate que as tasks de produção consultam internamente.

```
$ coursegen readiness check
✓ Readiness: APROVADO  (docs/06-course-readiness-checklist.md)
  Bloqueadores: 0 · Avisos: 1
  ⚠ carga do módulo 02 está 1h acima do previsto
  → Liberado para produção pela CLI.
```

### 4.3 `coursegen tasks list`

```
coursegen tasks list [--json]
```

```
$ coursegen tasks list
TASK                UNIDADE   READINESS  DESCRIÇÃO
generate-lessons    lesson    sim        Gera a aula completa a partir da lesson spec
generate-exercises  lesson    sim        Gera exercícios da aula
review-lessons      lesson    não        Revisa aulas geradas contra a spec
generate-slides     lesson    sim        Gera o deck de slides da aula
generate-projects   module    sim        Gera o enunciado de projeto do módulo
package-course      course    sim        Empacota o curso para distribuição
```

### 4.4 `coursegen tasks run <TASK>`

Comando central. Cria uma run e fan-out de execuções.

```
coursegen tasks run generate-lessons --runner claude --parallel 3
coursegen tasks run generate-lessons --runner codex --lesson lesson-01-01
coursegen tasks run review-lessons   --runner claude
coursegen tasks run generate-exercises --runner codex
coursegen tasks run generate-slides  --runner claude
```

| Flag | Default | Descrição |
|---|---|---|
| `--runner NAME` | `coursegen.yml` | Runner a usar |
| `--parallel N` | `1` | Sessões simultâneas |
| `--lesson ID` | (todas) | Filtra uma aula (`lesson-XX-YY`) |
| `--module ID` | (todos) | Filtra um módulo (`module-XX`) |
| `--force` | `false` | Regera mesmo se output existe e hash bate |
| `--dry-run` | `false` | Planeja e imprime o que faria, sem executar |
| `--no-readiness` | `false` | **Escape hatch** — ignora o gate (registra aviso) |
| `--timeout S` | `coursegen.yml` | Timeout por sessão |
| `--continue` | `false` | Continua a última run em vez de criar nova |

### 4.5 `coursegen tasks status`

```
coursegen tasks status [--run RUN_ID] [--json] [--watch]
```

```
$ coursegen tasks status
Run run_20260530_153000_ab12 · generate-lessons · runner=claude · parallel=3
Status: running   12/18 ok · 1 falhou · 2 rodando · 3 pendentes   ⏱ 4m12s

UNIDADE          STATUS     TENT.  DURAÇÃO   OUTPUT
lesson-01-01     ✓ ok        1     0m38s    output/lessons/module-01/lesson-01-01.md
lesson-01-02     ✓ ok        1     0m41s    output/lessons/module-01/lesson-01-02.md
lesson-02-03     ✗ falhou    2     —        (timeout do runner)
lesson-02-04     ◐ rodando   1     0m12s    —
lesson-03-01     · pendente  0     —        —
```

`--watch` redesenha a cada 1s até a run terminar.

### 4.6 `coursegen tasks retry <failed|all|EXEC_ID>`

```
coursegen tasks retry failed [--run RUN_ID] [--runner NAME]
```

Reseta as execuções selecionadas para `pending` e reexecuta **apenas elas**,
incrementando `attempt`. Sem `--run`, usa a run mais recente.

### 4.7 `coursegen runs list` / `coursegen runs show RUN_ID`

```
$ coursegen runs list
RUN_ID                      TASK              RUNNER  STATUS    OK/TOTAL  INÍCIO
run_20260530_153000_ab12    generate-lessons  claude  partial   17/18     30/05 15:30
run_20260530_141500_9f0a    readiness         —       ok        —         30/05 14:15

$ coursegen runs show run_20260530_153000_ab12
# resumo + tabela de execuções + caminhos de log + comando original + diffs de status
```

### 4.8 `coursegen workflow run <WORKFLOW>` (extra, mas previsto nos conceitos)

```
coursegen workflow run production --runner claude --parallel 3
```

Executa uma sequência de tasks (ex.: `generate-lessons` → `generate-exercises` →
`review-lessons`). Para na primeira task com falha bloqueante (configurável).

### 4.9 `coursegen doctor` (extra)

Roda o `healthcheck` de cada runner configurado (ex.: `claude --version`) e
reporta disponibilidade, versão e variáveis de ambiente faltando. Útil antes de
uma run grande.

---

## 5. Formato dos YAMLs

### 5.1 `coursegen.yml` (raiz do projeto)

```yaml
version: 1

course:
  name: "Engenharia de Software com Agentes de IA"
  slug: "eng-software-agentes-ia"
  language: pt-BR

paths:
  docs: docs
  output: output
  state: .coursegen/state.sqlite3
  logs: .coursegen/logs
  runs: .coursegen/runs

readiness:
  required: true
  source: docs/06-course-readiness-checklist.md
  approved_marker: "APROVADO"          # token procurado na seção "Veredito"

runners:
  default: claude
  available: [claude, codex, gemini, cursor, opencode]

execution:
  parallel: 1                          # default; sobrescrito por --parallel
  timeout_seconds: 900
  retry:
    max_attempts: 2
    backoff_seconds: 10
  on_validation_failure: fail          # fail | warn

# Context pack compartilhado por TODA aula (o mínimo comum).
# As partes "por unidade" (module spec, lesson spec) são definidas na task.
context:
  shared:
    - docs/01-course-prd.md
    - docs/02-market-research.md
    - docs/03-learning-architecture.md
  max_tokens_estimate: 60000           # aviso se o pack estourar
```

### 5.2 `tasks/generate-lessons.yml`

```yaml
name: generate-lessons
description: "Gera a aula completa a partir da lesson spec."
unit: lesson                           # lesson | module | course → granularidade do fan-out
requires_readiness: true

# Como descobrir as unidades de trabalho.
discover:
  glob: "docs/05-lesson-specs/module-*/lesson-*-*.md"
  # captura módulo e aula a partir do nome do arquivo
  id_pattern: "lesson-(?<module>\\d{2})-(?<lesson>\\d{2})"

# Context pack: o que entra na sessão isolada de CADA aula.
context:
  inherit_shared: true                 # inclui coursegen.yml > context.shared
  per_unit:
    module_spec: "docs/04-module-specs/module-{module}.md"
    lesson_spec: "docs/05-lesson-specs/module-{module}/lesson-{module}-{lesson}.md"
  # explicitamente proibido: incluir outras lesson specs
  exclude_globs:
    - "docs/05-lesson-specs/**"        # exceto a injetada acima

prompt_template: prompts/generate-lesson.md.erb

output:
  path: "output/lessons/module-{module}/lesson-{module}-{lesson}.md"
  capture: stdout                      # stdout | file
  overwrite: if_changed                # always | if_changed | never

# Validação heurística do artefato gerado.
acceptance:
  min_bytes: 800
  must_include_sections:
    - "Título"
    - "Objetivo"
    - "Contexto"
    - "Motivação"
    - "Explicação conceitual"
    - "Explicação técnica"
    - "Exemplo prático"
    - "Boas práticas"
    - "Erros comuns"
    - "Checklist de aprendizado"
    - "Exercício da aula"
    - "Resumo final"
  forbid_patterns:                     # detecta vazamento de escopo
    - "(?i)novo módulo"
    - "(?i)aula seguinte:"
```

### 5.3 `tasks/review-lessons.yml`

```yaml
name: review-lessons
description: "Revisa a aula gerada contra a lesson spec e os critérios de aceite."
unit: lesson
requires_readiness: false              # review pode rodar sobre rascunhos

discover:
  glob: "output/lessons/module-*/lesson-*-*.md"   # revisa o que foi GERADO
  id_pattern: "lesson-(?<module>\\d{2})-(?<lesson>\\d{2})"

context:
  inherit_shared: true
  per_unit:
    lesson_spec: "docs/05-lesson-specs/module-{module}/lesson-{module}-{lesson}.md"
    module_spec: "docs/04-module-specs/module-{module}.md"
    generated_lesson: "output/lessons/module-{module}/lesson-{module}-{lesson}.md"

prompt_template: prompts/review-lesson.md.erb

output:
  path: "output/reviews/lessons/module-{module}/lesson-{module}-{lesson}.review.md"
  capture: stdout
  overwrite: always

acceptance:
  must_include_sections: ["Veredito", "Conformidade com a spec", "Correções sugeridas"]
```

### 5.4 `tasks/generate-slides.yml`

```yaml
name: generate-slides
description: "Gera o deck de slides (Markdown/Marp) da aula."
unit: lesson
requires_readiness: true

discover:
  glob: "docs/05-lesson-specs/module-*/lesson-*-*.md"
  id_pattern: "lesson-(?<module>\\d{2})-(?<lesson>\\d{2})"

context:
  inherit_shared: true
  per_unit:
    lesson_spec: "docs/05-lesson-specs/module-{module}/lesson-{module}-{lesson}.md"
    module_spec: "docs/04-module-specs/module-{module}.md"
    generated_lesson: "output/lessons/module-{module}/lesson-{module}-{lesson}.md"  # opcional

prompt_template: prompts/generate-slides.md.erb

output:
  path: "output/slides/module-{module}/lesson-{module}-{lesson}.md"
  capture: stdout
  overwrite: if_changed

acceptance:
  min_bytes: 300
  must_include_patterns: ["^---$", "^#"]    # separadores de slide + título
```

### 5.5 `runners/claude.yml`

```yaml
name: claude
description: "Claude Code em modo headless (print)."
bin: claude
healthcheck: "claude --version"

# Como o prompt chega ao agente.
prompt:
  via: stdin                 # stdin | arg | file
  # se via: arg → use {prompt} em args; se via: file → {prompt_file}

# Argumentos da invocação. Tokens disponíveis:
#   {prompt} {prompt_file} {workdir} {output_path}
args:
  - "-p"                     # modo print/não-interativo
  - "--output-format"
  - "text"

# Como o context pack é fornecido ao agente.
context:
  strategy: inline_in_prompt # inline_in_prompt | copy_to_workdir | path_args

# De onde sai o artefato.
output:
  capture: stdout            # casa com tasks[].output.capture
  strip_code_fences: false

# Execução do processo.
cwd: "{workdir}"             # cada sessão roda no seu workdir isolado
env:
  ANTHROPIC_API_KEY: "${ANTHROPIC_API_KEY}"
timeout_seconds: 900
kill_signal: TERM
```

### 5.6 `runners/codex.yml`

```yaml
name: codex
description: "OpenAI Codex CLI em modo não-interativo."
bin: codex
healthcheck: "codex --version"

prompt:
  via: arg
args:
  - "exec"                   # subcomando não-interativo
  - "{prompt}"

context:
  strategy: inline_in_prompt

output:
  capture: stdout
  strip_code_fences: false

cwd: "{workdir}"
env:
  OPENAI_API_KEY: "${OPENAI_API_KEY}"
timeout_seconds: 900
kill_signal: TERM
```

> Runners `gemini`, `cursor`, `opencode` seguem o mesmo esquema, variando `bin`,
> subcomando e env (`gemini -p`, `cursor-agent -p`, `opencode run`). **Os flags
> exatos dependem da versão de cada ferramenta** — por isso ficam isolados em
> YAML, atualizáveis sem mexer no código Ruby. `coursegen doctor` valida cada um.

### 5.7 `workflows/production.yml` (exemplo)

```yaml
name: production
description: "Pipeline completo de produção de conteúdo."
requires_readiness: true
steps:
  - task: generate-lessons
  - task: generate-exercises
  - task: review-lessons
    continue_on_failure: true     # review não bloqueia o pipeline
stop_on_failure: true             # default para as demais steps
```

---

## 6. Interface dos runners

Todo runner implementa a **mesma interface**. Como a variação entre `claude`,
`codex`, etc. é quase toda de *invocação de processo*, há uma única classe
concreta `CliRunner` dirigida por YAML; runners exóticos podem subclassear.

```ruby
module CourseGen
  # Entrada imutável para uma execução.
  Invocation = Struct.new(
    :prompt,         # String — prompt final já montado
    :context_files,  # Array<Pathname> — pack (quando strategy != inline)
    :workdir,        # Pathname — diretório isolado da sessão
    :output_path,    # Pathname — destino canônico do artefato
    :env,            # Hash — variáveis extras
    :timeout,        # Integer (s)
    keyword_init: true
  )

  # Resultado normalizado, agnóstico de ferramenta.
  RunResult = Struct.new(
    :status,         # :ok | :failed | :timeout
    :artifact,       # String — conteúdo capturado (ou nil)
    :stdout, :stderr,
    :exit_code,
    :duration_ms,
    :error,          # mensagem se falhou
    keyword_init: true
  )

  module Runners
    # Contrato comum.
    class Base
      def initialize(config) = @config = config
      def name      = @config.fetch("name")
      def available? = system("#{@config['healthcheck']} > /dev/null 2>&1")
      def run(invocation) = raise NotImplementedError
    end

    # Implementação genérica dirigida por YAML — cobre os 5 runners.
    class CliRunner < Base
      def run(invocation)
        cmd   = build_argv(invocation)         # expande args + tokens
        env   = build_env(invocation)          # resolve ${VARS}
        stdin = prompt_via_stdin? ? invocation.prompt : nil

        started = monotonic_ms
        out, err, status = capture(cmd, env:, stdin:, chdir: invocation.workdir,
                                        timeout: invocation.timeout)

        artifact = extract_artifact(out, invocation)   # stdout ou arquivo
        RunResult.new(
          status:      classify(status, out, err),
          artifact:,
          stdout: out, stderr: err,
          exit_code:   status&.exitstatus,
          duration_ms: monotonic_ms - started,
          error:       (err if !status&.success?)
        )
      end

      private

      # Open3 com timeout e kill de process group (evita órfãos).
      def capture(cmd, env:, stdin:, chdir:, timeout:)
        Open3.popen3(env, *cmd, chdir: chdir.to_s, pgroup: true) do |i, o, e, t|
          i.write(stdin) if stdin
          i.close
          unless t.join(timeout)
            Process.kill("-#{@config['kill_signal'] || 'TERM'}", t.pid) rescue nil
            return ["", "timeout após #{timeout}s", nil]
          end
          [o.read, e.read, t.value]
        end
      end
      # build_argv / build_env / extract_artifact / classify: expandem o YAML.
    end
  end
end
```

**Garantias da interface:**

- `run` é **stateless** e **idempotente** do ponto de vista do runner: tudo que
  precisa vem na `Invocation`; nada de estado global.
- A sessão roda em `invocation.workdir` (isolamento de processo + filesystem).
- `RunResult.status` é normalizado (`:ok/:failed/:timeout`) — o `Executor` não
  precisa saber qual ferramenta rodou.
- Timeout sempre mata o **grupo de processos**, evitando subprocessos órfãos.

---

## 7. Fluxo de execução

Detalhe do `generate-lessons --runner claude --parallel 3` (os 12 passos pedidos):

```
 1. Carregar coursegen.yml + task generate-lessons.yml + runner claude.yml
    → snapshot em .coursegen/config.lock.json (auditoria)

 2. READINESS GATE
    Readiness.parse(docs/06-course-readiness-checklist.md)
      ├─ veredito contém "APROVADO"?  não → abortar (exit ≠0) salvo --no-readiness
      └─ sim → seguir

 3. DESCOBRIR UNIDADES
    glob "docs/05-lesson-specs/module-*/lesson-*-*.md"
      → [ {module: "01", lesson: "01"}, {module:"01", lesson:"02"}, ... ]
    aplica filtros --lesson / --module

 4. CRIAR RUN + EXECS  (transação SQLite)
    runs.insert(status: running, total: N, cmd, runner, parallel, version)
    para cada unidade: task_executions.insert(status: pending, attempt: 0)

 5–11. FAN-OUT  Parallel.each(execs, in_threads: 3) do |exec|
    5.  Encontrar a module spec correspondente: module-{module}.md
    6.  MONTAR CONTEXT PACK (mínimo, isolado):
          shared:  01-prd, 02-market, 03-architecture
          unit:    04-module-specs/module-{module}.md
                   05-lesson-specs/module-{module}/lesson-{module}-{lesson}.md
        → calcular input_hash = sha256(prompt + arquivos do pack + versão template)
        → se output existe E hash igual E status ok E não --force → SKIP (marca skipped)
    6b. PromptBuilder.render(template, pack) → PROMPT.md no workdir isolado
            .coursegen/runs/<run>/generate-lessons/lesson-{m}-{l}/
    7.  CRIAR SESSÃO ISOLADA: Runner#run(Invocation{prompt, workdir, timeout, output_path})
            (subprocesso dedicado; só vê este workdir)
    8.  GERAR A AULA: runner devolve RunResult (stdout capturado)
    9.  VALIDAR + SALVAR:
            Validator.check(artifact, task.acceptance)
              ok  → grava output/lessons/module-{m}/lesson-{m}-{l}.md
              !ok → status failed (ou warn, conforme on_validation_failure)
   10.  LOGS: stdout.txt, stderr.txt, <unidade>.log; artifacts.insert(sha256, bytes)
   11.  ATUALIZAR ESTADO (mutex): task_executions.update(status, attempt, duration, paths)
        - escrita serializada por um único writer (Mutex) sobre SQLite em WAL

 12. RETRY (sob demanda)
    coursegen tasks retry failed → reseta failed→pending e reexecuta só essas
    Run summary: status ok | partial | failed; reporter imprime tabela.
```

**Isolamento e anti-mistura (requisitos críticos):**

- *Estouro de contexto* → context pack mínimo + uma sessão por aula. O agente
  nunca recebe "todas as aulas".
- *Mistura entre módulos* → cada sessão é um processo separado, com workdir
  separado, contendo só a module/lesson spec daquele item. Não há canal pelo qual
  o conteúdo de uma aula vaze para outra.
- *Concorrência de estado* → SQLite em WAL; todas as escritas passam por um único
  writer protegido por `Mutex` (o orquestrador é um só processo).

---

## 8. Prompt templates

Templates ERB recebem um objeto `pack` (o context pack montado). Para o agente
ter exatamente o contexto mínimo — nem mais, nem menos — o conteúdo é **inlinado**
no prompt (estratégia default `inline_in_prompt`).

### 8.1 `prompts/generate-lesson.md.erb`

```erb
Você é um instrutor técnico especialista.

Você está gerando UMA ÚNICA aula de um curso online.

Use APENAS o contexto fornecido abaixo. Não use conhecimento externo que
contradiga estes documentos, e não invente partes do curso que não estão aqui.

Sua tarefa é gerar SOMENTE a aula indicada na Lesson Spec.

Você NÃO deve:
- Alterar a arquitetura do curso
- Criar novos módulos
- Criar novas aulas
- Ignorar os critérios de aceite
- Misturar conteúdo de outras aulas
- Gerar slides
- Gerar o projeto final
- Gerar exercícios fora do escopo desta aula

A aula deve conter, nesta ordem, com cada seção como um cabeçalho Markdown:
- Título
- Objetivo
- Contexto
- Motivação
- Explicação conceitual
- Explicação técnica
- Exemplo prático
- Exemplo de código (quando aplicável)
- Boas práticas
- Erros comuns
- Checklist de aprendizado
- Exercício da aula
- Resumo final

REGRA DE SAÍDA: responda APENAS com o Markdown da aula. Sem preâmbulo, sem
comentários, sem blocos de cerca ao redor do documento inteiro.

Idioma da aula: <%= pack.language %>

=================== CONTEXTO ===================

----- COURSE PRD -----
<%= pack.shared[:course_prd] %>

----- MARKET RESEARCH -----
<%= pack.shared[:market_research] %>

----- LEARNING ARCHITECTURE -----
<%= pack.shared[:learning_architecture] %>

----- MODULE SPEC (módulo <%= pack.unit.module %>) -----
<%= pack.unit[:module_spec] %>

----- LESSON SPEC (aula <%= pack.unit.module %>-<%= pack.unit.lesson %>) -----
<%= pack.unit[:lesson_spec] %>

================ FIM DO CONTEXTO ================

Gere agora a aula <%= pack.unit.module %>-<%= pack.unit.lesson %> conforme a
Lesson Spec e os critérios de aceite. Respeite estritamente o escopo da aula.
```

### 8.2 `prompts/review-lesson.md.erb` (resumo)

```erb
Você é um revisor pedagógico e técnico rigoroso.
Avalie a AULA GERADA contra a LESSON SPEC e a MODULE SPEC.

Produza um relatório com as seções:
- Veredito (APROVADA | REPROVADA)
- Conformidade com a spec (item a item dos critérios de aceite)
- Cobertura de escopo (faltou algo? sobrou conteúdo de outra aula?)
- Qualidade técnica (correção dos exemplos de código)
- Correções sugeridas (acionáveis)

Não reescreva a aula. Apenas revise.

----- LESSON SPEC -----
<%= pack.unit[:lesson_spec] %>
----- AULA GERADA -----
<%= pack.unit[:generated_lesson] %>
```

### 8.3 `prompts/generate-slides.md.erb` (resumo)

```erb
Gere um deck de slides em Markdown (compatível com Marp) para a aula indicada.
Use `---` como separador de slide. Um conceito por slide, bullets curtos.
Não gere a aula em prosa; gere SLIDES. Baseie-se na Lesson Spec (e na aula
gerada, se fornecida). Idioma: <%= pack.language %>.

----- LESSON SPEC -----
<%= pack.unit[:lesson_spec] %>
<% if pack.unit[:generated_lesson] %>
----- AULA GERADA (referência) -----
<%= pack.unit[:generated_lesson] %>
<% end %>
```

> O **input_hash** inclui o hash do template renderizado. Editar um prompt
> invalida o cache e marca os artefatos afetados como regeneráveis.

---

## 9. Modelo de estado

SQLite em modo **WAL**, um único arquivo `.coursegen/state.sqlite3`. Migrations
versionadas. Quatro tabelas principais + uma de eventos.

```sql
-- 001_init.sql

PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE runs (
  id                TEXT PRIMARY KEY,         -- run_20260530_153000_ab12
  workflow          TEXT,                     -- nullable (run de task única)
  task              TEXT,                     -- generate-lessons (se task única)
  runner            TEXT NOT NULL,
  parallel          INTEGER NOT NULL DEFAULT 1,
  status            TEXT NOT NULL,            -- running|ok|partial|failed|canceled
  total             INTEGER NOT NULL DEFAULT 0,
  succeeded         INTEGER NOT NULL DEFAULT 0,
  failed            INTEGER NOT NULL DEFAULT 0,
  skipped           INTEGER NOT NULL DEFAULT 0,
  cmd               TEXT NOT NULL,            -- linha de comando original
  coursegen_version TEXT NOT NULL,
  started_at        TEXT NOT NULL,
  finished_at       TEXT
);

CREATE TABLE task_executions (
  id            TEXT PRIMARY KEY,             -- exec_...
  run_id        TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  task          TEXT NOT NULL,                -- generate-lessons
  unit          TEXT NOT NULL,                -- lesson-01-01 | module-01 | course
  module        TEXT,                         -- 01
  lesson        TEXT,                         -- 01
  runner        TEXT NOT NULL,
  status        TEXT NOT NULL,                -- pending|running|ok|failed|skipped
  attempt       INTEGER NOT NULL DEFAULT 0,
  input_hash    TEXT,                         -- sha256 do context pack + template
  output_path   TEXT,
  prompt_path   TEXT,
  log_path      TEXT,
  exit_code     INTEGER,
  error         TEXT,
  duration_ms   INTEGER,
  started_at    TEXT,
  finished_at   TEXT
);
CREATE INDEX idx_exec_run    ON task_executions(run_id);
CREATE INDEX idx_exec_status ON task_executions(status);
CREATE UNIQUE INDEX idx_exec_unit ON task_executions(run_id, task, unit);

CREATE TABLE artifacts (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  exec_id     TEXT NOT NULL REFERENCES task_executions(id) ON DELETE CASCADE,
  kind        TEXT NOT NULL,                  -- lesson|exercise|slide|project|review
  path        TEXT NOT NULL,
  bytes       INTEGER NOT NULL,
  sha256      TEXT NOT NULL,
  created_at  TEXT NOT NULL
);
CREATE INDEX idx_artifact_exec ON artifacts(exec_id);

CREATE TABLE events (                          -- trilha de auditoria append-only
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id    TEXT REFERENCES runs(id) ON DELETE CASCADE,
  exec_id   TEXT,
  ts        TEXT NOT NULL,
  level     TEXT NOT NULL,                     -- info|warn|error
  message   TEXT NOT NULL
);

CREATE TABLE schema_migrations (version TEXT PRIMARY KEY);
```

### Máquina de estados de uma execução

```
            ┌─────────┐  hash bate + output ok      ┌─────────┐
   criada → │ pending │ ───────────────────────────▶│ skipped │
            └────┬────┘                              └─────────┘
                 │ scheduler pega o slot
                 ▼
            ┌─────────┐  runner ok + validação ok    ┌─────────┐
            │ running │ ────────────────────────────▶│   ok    │
            └────┬────┘                              └─────────┘
                 │ erro | timeout | validação falhou
                 ▼
            ┌─────────┐  tasks retry failed
            │ failed  │ ──────────────────────▶ (volta a pending, attempt += 1)
            └─────────┘
```

**Idempotência:** `input_hash = sha256(template_render ‖ conteúdo de cada arquivo
do pack ‖ versão do runner)`. Em uma reexecução, se o hash bate e o output existe
e o status é `ok`, a unidade é `skipped`. `--force` ignora o cache.

**Concorrência:** um processo, N threads. Todas as transições passam por
`Store#with_write { ... }` (Mutex global) sobre conexão em WAL — leituras
concorrem livremente, escritas serializam. `busy_timeout` configurado como rede
de segurança.

---

## 10. MVP

Objetivo do MVP: **gerar todas as aulas de um curso aprovado, de ponta a ponta,
de forma retomável**, com dois runners.

### Escopo incluído

| Capacidade | Detalhe |
|---|---|
| `coursegen init` | Scaffold + `coursegen.yml` + SQLite + defaults |
| `coursegen readiness check` | Parse do 06-checklist; gate funcional |
| `coursegen tasks run generate-lessons` | Fluxo completo dos 12 passos |
| `coursegen tasks status` | Tabela do estado da run |
| `coursegen tasks retry failed` | Retry de falhas |
| Runners | **claude** e **codex** (via `CliRunner` + YAML) |
| Context pack | Mínimo, inline, isolado por aula |
| Execução | **Sequencial primeiro** (`--parallel 1`) |
| Estado | SQLite + WAL + idempotência por hash |
| Logs | stdout/stderr/log por execução |

### Fora do MVP (vem depois)

- Paralelismo (`--parallel N>1`) — entra logo após o sequencial estar sólido.
- `generate-exercises`, `generate-slides`, `generate-projects`, `package-course`.
- `review-lessons` e workflows.
- Runners `gemini`, `cursor`, `opencode`.
- `runs list/show`, `--watch`, `--json`, `doctor`.

### Critério de pronto do MVP

> Dado um `course/` com 06-checklist APROVADO e lesson specs válidas,
> `coursegen tasks run generate-lessons --runner claude` produz
> `output/lessons/module-XX/lesson-XX-YY.md` para todas as aulas, registra estado,
> e `coursegen tasks retry failed` recupera qualquer falha sem reprocessar o que
> já passou.

---

## 11. Roadmap

### v0.1 — Núcleo sequencial (MVP)
- `init`, `readiness check`, `tasks run generate-lessons`, `tasks status`, `tasks retry failed`.
- Runners `claude` e `codex`. Execução sequencial. SQLite + idempotência. Logs.
- Context pack mínimo isolado + prompt template de aula.

### v0.2 — Paralelismo e observabilidade
- `--parallel N` (Parallel + threads, writer serializado).
- `runs list`, `runs show`, `tasks status --watch`, saída `--json`.
- `coursegen doctor` (healthcheck dos runners). `--dry-run`.
- Runner `gemini`.

### v0.3 — Pipeline de conteúdo completo
- Tasks `generate-exercises`, `generate-slides`, `generate-projects`.
- `review-lessons` + `tasks run review-lessons`.
- Workflows (`production`, `review`, `slides`) + `workflow run`.
- Runners `cursor` e `opencode`. Overrides de prompt/runner por projeto.

### v1.0 — Produção robusta
- `package-course` (empacota curso distribuível: índice, navegação, ZIP/site).
- `full-build` workflow ponta a ponta.
- Política de retry com backoff exponencial e *circuit breaker* por runner.
- Métricas/custo: tokens e tempo por aula (quando o runner expõe).
- Validação de aceite plugável (validators customizados em Ruby).
- Modo CI (`--json`, exit codes estáveis, sem TTY) e cache compartilhável.

---

## 12. Riscos técnicos

| # | Risco | Impacto | Mitigação |
|---|---|---|---|
| 1 | **Concorrência no SQLite** sob `--parallel` | Corrupção/locks | WAL + writer único com Mutex + `busy_timeout`; orquestrador é processo único |
| 2 | **Drift de flags dos CLIs** (claude/codex/… mudam de versão) | Runs quebram | Invocação isolada em YAML; `doctor` + `healthcheck`; sem flags hardcoded |
| 3 | **Não-determinismo do LLM** | Aula fora do formato | Validação de aceite (seções/regex); `on_validation_failure: fail`; retry |
| 4 | **Vazamento de escopo** (agente cria módulos/menciona outras aulas) | Mistura de conteúdo | Context pack mínimo + `forbid_patterns`; prompt restritivo; review task |
| 5 | **Estouro de contexto/tokens** | Falha ou custo alto | Pack mínimo por aula; `max_tokens_estimate` com aviso; nunca "todas as aulas" |
| 6 | **Custo descontrolado** (centenas de sessões pagas) | $$ | `--dry-run`, idempotência (skip do que não mudou), `--lesson/--module`, limites |
| 7 | **Processos órfãos / travados** | Recursos presos | Timeout por sessão + kill de process group (`pgroup: true`) |
| 8 | **stdout poluído** (agente fala antes do markdown) | Artefato sujo | Regra "responda só o markdown"; `strip_code_fences`; validação de bytes/seções |
| 9 | **Parsing frágil do readiness** | Gate erra | Marcador estruturado (`approved_marker`) na seção "Veredito"; `--strict` confere artefatos |
| 10 | **Segredos em arquivo** | Vazamento de chave | Chaves só via env (`${VAR}`); YAML nunca guarda segredo; logs redatados |
| 11 | **Retomada inconsistente** após crash | Execs presas em `running` | Na inicialização, reconciliar `running` órfãos → `failed`; `--continue` |
| 12 | **Specs inválidas/ausentes** apesar do APROVADO | Pack quebrado | Validar existência dos arquivos do pack antes do spawn; falha clara por unidade |

---

## 13. Decisões arquiteturais (ADRs)

**ADR-000 — Go em vez de Ruby (revisão).** O design nasceu em Ruby, mas para
**distribuir** a CLI o ponto crítico é instalação sem fricção. Como a ferramenta
é cola I/O-bound (o trabalho pesado é o agente externo), performance não decide;
distribuição decide. Go entrega **binário único, estático, sem runtime**,
cross-compilado para macOS/Linux/Windows — eliminando o atrito de runtime + gems
nativas do Ruby. O design conceitual portou ~1:1.

**ADR-001 — stdlib `flag` + dispatch próprio (não um framework de CLI).** A
superfície de comandos é pequena (`init`, `doctor`, `readiness`, `tasks`,
`runs`). O `flag` da stdlib + um switch de subcomandos cobre tudo sem dependência
extra, mantendo o binário mínimo. (Cobra seria o equivalente ao Thor; dispensado
para manter zero deps além de `yaml.v3`.)

**ADR-002 — Estado em JSON no MVP, SQLite no roadmap.** O MVP roda **sequencial,
em um processo**: não há escritor concorrente nem necessidade de SQL ainda, e
JSON mantém o binário sem dependência nativa (CGO off). A API de `state` é
"store-shaped" para trocar por SQLite (consultas/transações/concorrência) quando
o paralelismo entrar, sem mexer nos chamadores.

**ADR-003 — Uma sessão (processo) por aula.** É o coração do isolamento:
garante zero mistura entre módulos e limita o contexto ao mínimo. Custo: overhead
de spawn — aceitável frente ao trabalho do LLM (segundos vs. dezenas de segundos).

**ADR-004 — Context pack mínimo, inline por padrão.** Inlinar o pack no prompt
torna o contexto **explícito e runner-agnóstico** (não depende de o agente
"resolver ler arquivos"). `copy_to_workdir`/`path_args` ficam como estratégias
opcionais para runners que leem arquivos de forma eficiente.

**ADR-005 — Captura por stdout como default.** Gerar "um arquivo markdown" é
determinístico via captura de stdout: o prompt manda responder só o conteúdo, e a
CLU grava. Evita depender do comportamento de escrita-em-disco do agente. `file`
fica disponível para casos multi-arquivo.

**ADR-006 — Runner único dirigido por config (`CliRunner`).** A diferença entre
os 5 runners é quase toda de invocação de processo. Um `struct` `Spec` (defaults
embarcados + override YAML por projeto) evita 5 implementações quase idênticas e
permite adicionar/ajustar runner sem recompilar o que importa.

**ADR-007 — Sequencial no MVP; goroutines (não processos) no roadmap.** O
orquestrador é I/O-bound (o trabalho pesado é o subprocesso do agente). O MVP é
estritamente sequencial — exigência do projeto (uma aula por vez, contexto limpo,
gasto previsível). Para paralelizar depois, goroutines + worker pool com semáforo
dão fan-out de espera barato, com escritas de estado serializadas por um mutex.

**ADR-008 — Idempotência por `input_hash`.** Reexecução barata e segura: pula o
que não mudou. Essencial para cursos grandes e para retry sem reprocessar tudo.

**ADR-009 — Configuração em camadas (defaults embarcados → projeto → flags).**
Os defaults de runner e o template de prompt são compilados no binário
(`//go:embed`); o projeto do curso pode sobrescrever em
`coursegen/{runners,prompts}/`; flags têm a última palavra. Instrutores
customizam sem recompilar; padrões sãos saem da caixa.

**ADR-010 — Readiness como gate de primeira classe.** Materializa a regra "agente
define, CLI escala": produção é bloqueada por padrão até `APROVADO`, com escape
hatch explícito e auditável (`--skip-readiness`).

---

## 14. Exemplo completo de execução

Curso já definido pelas skills; specs em `docs/`. Sessão de terminal completa:

```console
$ cd ~/cursos/eng-software-agentes-ia

# 1) Inicializa o projeto CourseGen (uma vez)
$ coursegen init --name "Engenharia de Software com Agentes de IA" --runner claude
✓ coursegen.yml criado
✓ .coursegen/state.sqlite3 inicializado (schema v1)
✓ defaults copiados em coursegen/{tasks,runners,prompts}/
→ Próximo passo: coursegen readiness check

# 2) Verifica o gate
$ coursegen readiness check
✓ Readiness: APROVADO  (docs/06-course-readiness-checklist.md)
  Bloqueadores: 0 · Avisos: 1
  → Liberado para produção pela CLI.

# 3) Confere os runners disponíveis
$ coursegen doctor
✓ claude   v1.x   (ANTHROPIC_API_KEY ok)
✓ codex    v0.x   (OPENAI_API_KEY ok)
✗ gemini   não encontrado no PATH

# 4) Lista as tasks
$ coursegen tasks list
generate-lessons    lesson   readiness=sim   Gera a aula completa
review-lessons      lesson   readiness=não   Revisa aulas geradas
...

# 5) Simula antes de gastar tokens
$ coursegen tasks run generate-lessons --runner claude --parallel 3 --dry-run
Plano (run NÃO criada):
  18 aulas descobertas em docs/05-lesson-specs/
  context pack/aula: 01-prd, 02-market, 03-architecture, module-XX, lesson-XX-YY
  output → output/lessons/module-XX/lesson-XX-YY.md
  estimativa de contexto/aula: ~42k tokens (limite de aviso: 60k) ✓

# 6) Roda de verdade (3 sessões isoladas em paralelo)
$ coursegen tasks run generate-lessons --runner claude --parallel 3
Run run_20260530_153000_ab12 · 18 aulas · runner=claude · parallel=3
  ✓ lesson-01-01  0m38s   output/lessons/module-01/lesson-01-01.md
  ✓ lesson-01-02  0m41s   output/lessons/module-01/lesson-01-02.md
  ✓ lesson-01-03  0m35s   output/lessons/module-01/lesson-01-03.md
  ...
  ✗ lesson-02-03  timeout após 900s  (tentativa 1/2)
  ...
Concluído: 17 ok · 1 falhou · 0 pulados   ⏱ 6m02s
Status da run: PARTIAL

# 7) Inspeciona o estado
$ coursegen tasks status
Run run_20260530_153000_ab12 · generate-lessons · PARTIAL · 17/18 ok
lesson-02-03   ✗ falhou   tent.1   —   (timeout do runner)

# 8) Investiga a falha
$ coursegen runs show run_20260530_153000_ab12
... (resumo + caminhos)
log:    .coursegen/logs/run_.../lesson-02-03.log
prompt: .coursegen/runs/run_.../generate-lessons/lesson-02-03/PROMPT.md

# 9) Reexecuta só o que falhou (incrementa attempt)
$ coursegen tasks retry failed
Reexecutando 1 execução falha de run_20260530_153000_ab12...
  ✓ lesson-02-03  0m44s   output/lessons/module-02/lesson-02-03.md  (tentativa 2)
Concluído: 18 ok · 0 falhou   ⏱ 0m44s
Status da run: OK

# 10) Reexecução do curso inteiro é barata (idempotência)
$ coursegen tasks run generate-lessons --runner claude --parallel 3
Run run_20260530_161200_cd34 · 18 aulas
  ⤼ 18 puladas (input_hash inalterado; nada a regerar)
Status da run: OK (0 sessões de agente disparadas)

# 11) Próximas fases do pipeline
$ coursegen tasks run generate-exercises --runner codex
$ coursegen tasks run generate-slides   --runner claude
$ coursegen tasks run review-lessons    --runner claude
# ou, de uma vez:
$ coursegen workflow run production --runner claude --parallel 3
```

**Estado final no disco:**

```
output/
├── lessons/module-01/lesson-01-01.md   ... lesson-02-03.md   (18 aulas)
├── exercises/...                        (após generate-exercises)
├── slides/...                           (após generate-slides)
└── reviews/lessons/...                  (após review-lessons)
.coursegen/state.sqlite3                 (1 run OK + execuções auditáveis)
```

> Resultado: o curso inteiro foi produzido em **sessões isoladas, retomáveis e
> auditáveis**, sem nenhuma sessão gigante, sem mistura entre módulos, e com o
> gate de readiness garantindo que a CLI só escalou o que o agente já havia
> definido e aprovado.
```
