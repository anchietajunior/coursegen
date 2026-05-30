# CourseGen — Pacote de Skills de Planejamento

Skills reutilizáveis e **agent-agnósticas** (Claude Code, Codex, Gemini CLI,
Cursor Agent, OpenCode) para transformar uma **ideia vaga de curso** em um pacote
de planejamento sólido e pronto para produção.

> **Princípio central: o agente DEFINE, a CLI ESCALA.**
> Estas skills cobrem apenas a **fase de definição**, dentro de um agente
> interativo. Nenhuma delas produz aulas, exercícios, projetos ou slides — isso é
> trabalho da CLI, e só depois que o curso for **APROVADO** no readiness check.

---

## Fluxo do sistema

```
Ideia do curso
  → course-discovery       → docs/00-course-discovery.md
  → course-prd             → docs/01-course-prd.md
  → market-research        → docs/02-market-research.md
  → learning-architecture  → docs/03-learning-architecture.md
  → module-specs           → docs/04-module-specs/module-XX.md
  → lesson-specs           → docs/05-lesson-specs/module-XX/lesson-XX-YY.md
  → course-readiness       → docs/06-course-readiness-checklist.md  (GATE)
  ───────────────────────────────────────────────────────────────
  → [APROVADO] → Produção em escala pela CLI
                 (aulas, exercícios, projetos, slides, review final)
```

Cada skill **lê** os artefatos das fases anteriores e **escreve** o seu próprio.
A `course-readiness` é um **portão**: pode reprovar e devolver um plano de correção.

---

## As skills

| # | Skill | Lê | Escreve |
|---|---|---|---|
| 1 | [`course-discovery`](course-discovery.md) | (só a ideia) | `docs/00-course-discovery.md` |
| 2 | [`course-prd`](course-prd.md) | 00 | `docs/01-course-prd.md` |
| 3 | [`market-research`](market-research.md) | 00, 01 | `docs/02-market-research.md` |
| 4 | [`learning-architecture`](learning-architecture.md) | 01, 02 | `docs/03-learning-architecture.md` |
| 5 | [`module-specs`](module-specs.md) | 01, 02, 03 | `docs/04-module-specs/module-XX.md` |
| 6 | [`lesson-specs`](lesson-specs.md) | 01, 03, 04 | `docs/05-lesson-specs/module-XX/lesson-XX-YY.md` |
| 7 | [`course-readiness`](course-readiness.md) | 00–05 | `docs/06-course-readiness-checklist.md` |

---

## Convenções compartilhadas

- **Artefatos numerados** em `docs/`, na ordem do fluxo (`00` → `06`).
- **Numeração de dois dígitos** para módulos (`module-01`) e aulas (`lesson-01-03`),
  consistente entre arquitetura, module specs e lesson specs.
- **Autossuficiência:** cada artefato deve ser legível por outro agente sem acesso
  à conversa original. Toda decisão importante é explícita; suposições são registradas.
- **Rastreabilidade:** toda decisão de uma fase deve ser rastreável até a fase
  anterior (objetivo do PRD → módulo → aula).
- **Idioma:** cada skill escreve na mesma língua usada nos documentos de origem.
- **Sobrescrita:** confirme com o usuário antes de sobrescrever um artefato existente.

---

## Como cada skill está especificada

Todas seguem o mesmo formato de 14 seções:

1. Nome · 2. Propósito · 3. Quando usar · 4. Quando não usar · 5. Inputs
obrigatórios · 6. Outputs gerados · 7. Processo interno · 8. Perguntas ao usuário ·
9. Regras de execução · 10. Template do arquivo gerado · 11. Critérios de
validação · 12. Erros comuns · 13. Exemplo de uso · 14. Prompt completo da skill.

O **item 14 (prompt completo)** é copiável e carrega a skill em qualquer agente.

---

## Regras de ouro do pacote

- ❌ Não pule o discovery.
- ❌ Não gere aulas completas, exercícios ou slides nesta fase.
- ❌ Não deixe decisões importantes implícitas.
- ❌ Não dependa de conhecimento prévio da conversa.
- ✅ Uma responsabilidade por skill, um artefato por skill.
- ✅ Só libere para a CLI depois de `course-readiness` = **APROVADO**.
