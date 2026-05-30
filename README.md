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

As skills são arquivos Markdown **autossuficientes** — "instalar" é apenas
torná-las disponíveis para o seu agente. Cada arquivo já traz o frontmatter
`name` e `description` exigidos (campos extras como `reads`/`writes` são
metadados do CourseGen e são ignorados pelo agente). Escolha uma forma:

**Opção A — Claude Code (recomendado).** Cada skill vira uma pasta com um
`SKILL.md`. Defina `DEST` como global (reutilizável entre cursos) ou por projeto
(versionado junto com o curso) e rode:

```bash
# Global (todos os cursos):
DEST="$HOME/.claude/skills"
# ...ou por projeto (dentro do repositório do curso):
# DEST="./.claude/skills"

for f in skills/*.md; do
  name=$(basename "$f" .md)
  [ "$name" = "README" ] && continue
  mkdir -p "$DEST/$name"
  cp "$f" "$DEST/$name/SKILL.md"
done
```

Depois, invoque pelo nome no Claude Code, ex.: `/course-discovery`.

**Opção B — Qualquer outro agente (Codex, Gemini CLI, Cursor, OpenCode).** Não
há um diretório padrão único entre as ferramentas. Use o item **14 (Prompt
completo da skill)** de cada arquivo — ele é autossuficiente: copie o prompt e
cole no agente, ou simplesmente aponte o agente para o arquivo `skills/<nome>.md`.

**Opção C — Sem instalação.** Mantenha a pasta `skills/` no projeto do curso e
peça ao agente para ler `skills/<nome>.md` no momento de usar.

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

### Rodando a CLI

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
> cd examples/sample-course && ../../bin/coursegen tasks run generate-lessons --runner mock
> ```
> Ou direto do código: `go run ./cmd/coursegen <comando>`.

### Comandos mínimos

Os exemplos abaixo usam `coursegen` (assuma o binário no PATH; ou use
`./bin/coursegen` / `go run ./cmd/coursegen`):

```bash
# Inicializa o projeto (uma vez): cria coursegen.yml, .coursegen/, output/
coursegen init --name "Meu Curso" --runner claude

# Confere o gate (precisa estar APROVADO)
coursegen readiness check

# Lista as tasks disponíveis
coursegen tasks list

# Verifica quais runners estão disponíveis no PATH
coursegen doctor

# (Opcional) Planeja sem executar e mostra a estimativa de tokens
coursegen tasks run generate-lessons --runner claude --dry-run

# Gera todas as aulas — SEQUENCIAL, uma sessão isolada por aula,
# contexto limpo entre aulas (aula 1 termina → aula 2 começa do zero)
coursegen tasks run generate-lessons --runner claude

# Acompanha o status e recupera falhas (só reexecuta o que falhou)
coursegen tasks status
coursegen tasks retry failed

# Regera só uma aula / um módulo (--force ignora o cache)
coursegen tasks run generate-lessons --runner claude --lesson lesson-01-01 --force
coursegen tasks run generate-lessons --runner claude --module 02

# Outras fases de produção (roadmap v0.3 — ainda não implementadas)
# coursegen tasks run generate-exercises --runner codex
# coursegen tasks run generate-slides    --runner claude
# coursegen tasks run review-lessons     --runner claude
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
> `readiness check`, `tasks list`, `tasks run generate-lessons` (sequencial),
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
