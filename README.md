# CourseGen

Sistema para criar cursos online de programação, engenharia de software e IA
usando agentes (Claude Code, Codex, Gemini CLI, Cursor Agent, OpenCode).

> **Regra central: o agente DEFINE, a CLI ESCALA.**
> Primeiro um agente interativo *planeja* o curso usando um pacote de **skills**.
> Depois a **CLI** lê o planejamento aprovado e *produz* aulas, exercícios,
> projetos e slides em escala — em sessões isoladas, retomáveis e auditáveis.

---

## Como funciona

```
        FASE 1 — DEFINIÇÃO (agente + skills)          FASE 2 — PRODUÇÃO (CLI)
  ┌──────────────────────────────────────────┐   ┌──────────────────────────┐
  Ideia → discovery → PRD → market research →     readiness check (gate)
  → arquitetura → module specs → lesson specs →   → generate-lessons
  → readiness check                               → generate-exercises
                  │                                → generate-slides
                  ▼   gera docs/ aprovado          → review / package
            docs/00..06  ───────────────────────▶  lê docs/ e gera output/
  └──────────────────────────────────────────┘   └──────────────────────────┘
```

A CLI **só** executa produção depois que o readiness check aprovar o curso.

---

## Estrutura do repositório

```
coursegen/
├── README.md          # este arquivo — porta de entrada
├── DESIGN.md          # arquitetura completa da CLI (Ruby)
└── skills/            # pacote de skills de planejamento (agent-agnósticas)
    ├── README.md      # índice e convenções das skills
    ├── course-discovery.md
    ├── course-prd.md
    ├── market-research.md
    ├── learning-architecture.md
    ├── module-specs.md
    ├── lesson-specs.md
    └── course-readiness.md
```

Quando você usa o sistema em um curso, o **projeto do curso** tem esta forma
(criada/preenchida pelas skills e pela CLI):

```
course/
├── coursegen.yml      # config da CLI
├── docs/              # ENTRADA: planejamento gerado pelas skills (00..06)
├── .coursegen/        # estado interno (SQLite, logs, runs)
└── output/            # SAÍDA: lessons/ exercises/ projects/ slides/ reviews/
```

---

## Parte 1 — Usando as skills (planejamento)

As skills rodam **dentro de um agente interativo** (Claude Code, Codex, etc.).
Cada skill tem uma responsabilidade única, lê os artefatos anteriores e gera o
próximo. Rode-as **em ordem**:

| Ordem | Skill | Gera |
|---|---|---|
| 1 | `course-discovery` | `docs/00-course-discovery.md` |
| 2 | `course-prd` | `docs/01-course-prd.md` |
| 3 | `market-research` | `docs/02-market-research.md` |
| 4 | `learning-architecture` | `docs/03-learning-architecture.md` |
| 5 | `module-specs` | `docs/04-module-specs/module-XX.md` |
| 6 | `lesson-specs` | `docs/05-lesson-specs/module-XX/lesson-XX-YY.md` |
| 7 | `course-readiness` | `docs/06-course-readiness-checklist.md` (gate) |

### Instalando as skills

A forma recomendada é o comando **`coursegen setup`** (as skills vêm
**embarcadas no binário**, então funciona logo após `brew install`):

```bash
coursegen setup                 # escolhe o agente interativamente
coursegen setup --agent claude  # ou direto, sem prompt
coursegen setup --list          # lista as skills embarcadas
```

O que ele faz, espelhando o padrão do Compozy:

1. **Pergunta qual agente você usa** (claude, codex, gemini, cursor, opencode) —
   ou aceita `--agent`.
2. Escreve as 7 skills no **store agnóstico** `~/.agents/skills/<nome>/SKILL.md`.
3. **Symlinka** (ou `--copy`) para o diretório do agente — ex.: `~/.claude/skills/`
   (Claude Code), `~/.cursor/skills-cursor/` (Cursor). Para agentes sem diretório
   de skills conhecido, deixa no store agnóstico e indica como apontá-los.
4. Registra a escolha em `~/.config/coursegen/state.yml`.

Depois, invoque pelo nome no agente, ex.: `/course-discovery`.

> **Instalação manual (alternativa).** Como as skills são Markdown autossuficiente
> (frontmatter `name`/`description`; campos extras são ignorados pelo agente),
> dá para copiá-las à mão: `cp skills/<nome>.md $DEST/<nome>/SKILL.md`. Em agentes
> sem conceito de "skill", use o item **14 (Prompt completo)** de cada arquivo —
> ele é copiável e autossuficiente.

**Como executar uma skill:** abra o agente no diretório do curso e carregue a
skill. Cada arquivo em `skills/` tem, no item **14 (Prompt completo da skill)**,
um prompt copiável e autossuficiente — cole-o no agente para ativá-la.

```
# Exemplo (fase de definição, dentro do agente)
1. Comece pela course-discovery → o agente te entrevista e gera docs/00.
2. Rode course-prd, depois market-research, learning-architecture, module-specs
   e lesson-specs — cada uma lê o que a anterior gerou.
3. Rode course-readiness por último. Ela APROVA ou REPROVA o curso.
   Só avance para a CLI quando o veredito for APROVADO.
```

Detalhes, convenções e regras: [`skills/README.md`](skills/README.md).

---

## Parte 2 — Usando a CLI (produção)

Depois que `docs/06-course-readiness-checklist.md` estiver **APROVADO**, a CLI
escala a produção. Cada aula é gerada em **um agente separado, sequencialmente**
(aula 1 → termina → contexto limpo → aula 2 …), com um **context pack mínimo**, e
o status é exibido enquanto roda.

### Instalação

```bash
# Homebrew (quando o tap estiver publicado):
brew install coursegen/tap/coursegen

# Em seguida, escolha seu agente e instale as skills de planejamento:
coursegen setup
```

`coursegen setup` (detalhado na [Parte 1](#instalando-as-skills)) pergunta qual
agente você usa e copia as skills para o lugar certo — as skills vêm
**embarcadas no binário**, então nada extra é baixado.

### Rodando a CLI a partir do código

O MVP está implementado em **Go** (`cmd/coursegen`, `internal/`) — compila para
um **binário único, sem runtime**, fácil de distribuir. Requer Go 1.26+ apenas
para compilar.

```bash
make build            # gera ./bin/coursegen
# ou:
go build -o bin/coursegen ./cmd/coursegen
# ou instale no PATH:
go install github.com/coursegen/coursegen/cmd/coursegen@latest
```

Para distribuir binários prontos (sem Go na máquina do usuário):

```bash
make release          # cross-compila para darwin/linux/windows em ./dist (CGO off)
```

> **Teste sem gastar tokens:** use `--runner mock` (gera conteúdo de exemplo).
> Há um curso pronto em `examples/sample-course/` para experimentar:
> ```bash
> make build
> cd examples/sample-course && ../../bin/coursegen generate lessons --runner mock
> ```
> Ou direto do código: `go run ./cmd/coursegen <comando>`.

### Comandos mínimos

A superfície é **verbo-first**: você *gera* / *revisa* **lessons** (o artefato).
Não há um substantivo genérico "task" na linha de comando — o verbo é a operação,
"lesson" é o conteúdo produzido. Os exemplos assumem o binário no PATH (ou use
`./bin/coursegen` / `go run ./cmd/coursegen`):

```bash
# Inicializa o projeto (uma vez): cria coursegen.yml, .coursegen/, output/
coursegen init --name "Meu Curso" --runner claude

# Confere o gate (precisa estar APROVADO)
coursegen readiness check

# Lista os geradores disponíveis
coursegen list

# Verifica quais runners estão disponíveis no PATH
coursegen doctor

# (Opcional) Planeja sem executar e mostra a estimativa de tokens
coursegen generate lessons --runner claude --dry-run

# Gera todas as aulas — SEQUENCIAL, uma sessão isolada por aula,
# contexto limpo entre aulas (aula 1 termina → aula 2 começa do zero)
coursegen generate lessons --runner claude

# Acompanha o status e recupera falhas (só reexecuta o que falhou)
coursegen status
coursegen retry failed

# Regera só uma aula / um módulo (--force ignora o cache)
coursegen generate lessons --runner claude --lesson lesson-01-01 --force
coursegen generate lessons --runner claude --module 02

# Outras fases de produção (roadmap v0.3 — ainda não implementadas)
# coursegen generate exercises --runner codex
# coursegen generate slides    --runner claude
# coursegen review lessons     --runner claude
```

> **Economia de tokens é critério do projeto.** A execução é sempre sequencial
> (nunca tudo ao mesmo tempo); cada aula roda num processo de agente novo (contexto
> limpo, sem arrasto da aula anterior); o context pack é o mínimo necessário; e
> aulas já geradas e inalteradas são **puladas** (0 tokens) numa nova execução.

Os artefatos saem em `output/` (`lessons/`, `exercises/`, `slides/`, `reviews/`).
Cada aula é gerada em sessão isolada, com um **context pack mínimo** (PRD, market
research, arquitetura, a module spec e a lesson spec daquela aula) — nunca todas
as aulas de uma vez, evitando estouro de contexto e mistura entre módulos.

Arquitetura completa, YAMLs, runners, modelo de estado e roadmap:
[`DESIGN.md`](DESIGN.md).

> **MVP implementado em Go** (`cmd/coursegen`, `internal/`): `init`, `doctor`,
> `readiness check`, `list`, `generate lessons` (sequencial),
> `tasks status`, `tasks retry failed`, `runs list/show`. Runners: `claude`,
> `codex`, `gemini`, `cursor`, `opencode` (dirigidos por YAML) + `mock` para
> testes sem custo. Binário único, sem runtime. A arquitetura completa e o
> roadmap (paralelismo, demais tasks) estão em [`DESIGN.md`](DESIGN.md).

---

## Regras de ouro

- ❌ Não pule o discovery, nem produza conteúdo na fase de definição.
- ❌ A CLI não define o curso e não roda sem readiness **APROVADO**.
- ✅ Uma responsabilidade por skill; uma sessão isolada por aula na CLI.
- ✅ Tudo rastreável: objetivo do PRD → módulo → aula → artefato gerado.
