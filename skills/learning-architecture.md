---
name: learning-architecture
description: Cria a arquitetura pedagógica do curso — módulos, ordem, dependências, competências, carga horária, progressão de dificuldade e conexão com o projeto final.
phase: 04-architecture
reads:
  - docs/01-course-prd.md
  - docs/02-market-research.md
writes:
  - docs/03-learning-architecture.md
agent_agnostic: true
---

# Skill: Learning Architecture

> Princípio central do CourseGen: **o agente define, a CLI escala.**
> Esta skill é a planta baixa do curso. Ela decide a sequência de aprendizagem;
> as specs de módulos e aulas apenas detalham o que esta arquitetura definir.

---

## 1. Nome da skill

`learning-architecture`

---

## 2. Propósito

Transformar o PRD e a pesquisa de mercado em uma **arquitetura pedagógica**: a
lista de módulos, a ordem em que serão ensinados, a justificativa dessa ordem, as
dependências entre módulos, as competências de cada um, a carga horária, a
progressão de dificuldade, os pontos de revisão e prática, e como tudo isso
converge no projeto final.

É aqui que o curso ganha **forma e sequência**. A skill aplica princípios de
design instrucional (pré-requisitos antes de dependentes, dificuldade crescente,
prática espaçada, recuperação ativa) para que o aluno chegue ao final capaz de
executar a transformação prometida no PRD.

---

## 3. Quando usar

- Existem `docs/01-course-prd.md` e `docs/02-market-research.md` válidos.
- Os conflitos da pesquisa de mercado já foram resolvidos (ou conscientemente aceitos).
- Antes de gerar `module-specs` e `lesson-specs`.

---

## 4. Quando não usar

- Falta PRD ou pesquisa de mercado → rode as skills anteriores.
- Há conflitos não resolvidos entre mercado e PRD → resolva primeiro.
- O usuário quer detalhar cada módulo internamente → use `module-specs`.
- O usuário quer detalhar aulas → use `lesson-specs`.
- O usuário quer produzir conteúdo/exercícios/slides → fora desta fase.

---

## 5. Inputs obrigatórios

| Input | Origem | Obrigatório |
|---|---|---|
| `docs/01-course-prd.md` | Skill `course-prd` | ✅ Sim |
| `docs/02-market-research.md` | Skill `market-research` | ✅ Sim |

> Se a pesquisa de mercado listar conflitos não resolvidos com o PRD, a skill
> deve apontá-los e pedir decisão antes de arquitetar.

---

## 6. Outputs gerados

| Output | Caminho | Formato |
|---|---|---|
| Arquitetura pedagógica | `docs/03-learning-architecture.md` | Markdown |

---

## 7. Processo interno

1. **Ler** PRD e pesquisa de mercado.
2. **Extrair** objetivos mensuráveis, escopo, tecnologias e recomendações de mercado.
3. **Derivar competências** que o curso precisa desenvolver para cumprir cada objetivo.
4. **Agrupar competências em módulos** coesos (cada módulo = um bloco de competências relacionadas).
5. **Ordenar os módulos** respeitando dependências (pré-requisito sempre antes do dependente) e progressão de dificuldade.
6. **Justificar a ordem** explicitamente.
7. **Estimar carga horária** por módulo, somando dentro da duração total do PRD.
8. **Inserir pontos de revisão e de prática** ao longo da trilha (recuperação ativa, prática espaçada).
9. **Conectar o projeto final** aos módulos: indicar qual módulo contribui com qual parte do projeto.
10. **Verificar cobertura:** todo objetivo do PRD é coberto por ≥1 módulo? Todo módulo serve a ≥1 objetivo?
11. **Gerar** `docs/03-learning-architecture.md` pelo template da seção 10.
12. **Validar** contra a seção 11.
13. **Entregar** e sugerir `module-specs`.

---

## 8. Perguntas que deve fazer ao usuário

A skill é majoritariamente derivacional, mas confirma escolhas de design quando há trade-off:

- Se a carga horária estourar a duração do PRD: *"A trilha completa soma X horas,
  acima das Y do PRD. Corto escopo, reduzo profundidade, ou ajusto a duração?"*
- Se houver duas ordens pedagógicas válidas: *"Posso ensinar A→B (fundamentos
  primeiro) ou B→A (orientado a projeto desde o início). Qual estilo prefere?"*
- Se o projeto final não cobrir todos os módulos: *"O módulo Z não alimenta o
  projeto final. Mantemos como teoria de apoio, viramos exercício, ou cortamos?"*
- Se um objetivo do PRD não couber em nenhum módulo natural: *"O objetivo W não
  se encaixa na trilha atual. É um módulo próprio ou parte de outro?"*

---

## 9. Regras de execução

- **NÃO** gere a arquitetura sem PRD e pesquisa de mercado válidos.
- **Cobertura bidirecional:** todo objetivo do PRD → ≥1 módulo; todo módulo → ≥1 objetivo.
- **Dependências respeitadas:** nenhum módulo depende de competência ensinada depois dele.
- **Dificuldade crescente:** a progressão deve ser justificada, não aleatória.
- **Carga horária** deve somar dentro (ou explicitamente ajustar) a duração do PRD.
- **Projeto final mapeado** módulo a módulo.
- Incorpore as **recomendações de mercado** (adicionar/remover/repriorizar) já decididas.
- **NÃO** escreva specs de módulo nem de aula — só a arquitetura de alto nível.
- **NÃO** produza conteúdo de ensino.
- Numere módulos com dois dígitos (`module-01`, `module-02`, ...) para casar com as próximas skills.
- Escreva na mesma língua do PRD.

---

## 10. Template do arquivo gerado

```markdown
# Learning Architecture — {{Nome do curso}}

> Documento gerado pela skill `learning-architecture`.
> Fontes: docs/01-course-prd.md, docs/02-market-research.md
> Status: [ ] Rascunho  [ ] Validado
> Data: {{AAAA-MM-DD}}

## 1. Visão da trilha
{{Parágrafo: como o curso evolui do início ao projeto final.}}

## 2. Filosofia pedagógica
- **Estilo:** {{fundamentos-primeiro | orientado a projeto | espiral | ...}}
- **Princípios aplicados:** {{progressão de dificuldade, prática espaçada, recuperação ativa, ...}}

## 3. Mapa de módulos
| # | Módulo | Competências-chave | Carga (h) | Dificuldade |
|---|---|---|---|---|
| 01 | {{...}} | {{...}} | {{...}} | {{1-5}} |
| 02 | {{...}} | {{...}} | {{...}} | {{1-5}} |

**Carga horária total:** {{soma}} h (PRD prevê {{X}} h — {{ok / ajustado}})

## 4. Ordem pedagógica e justificativa
- **Módulo 01 → 02 → ...:** {{por que esta sequência}}
- {{justificativa de cada transição relevante}}

## 5. Dependências entre módulos
| Módulo | Depende de | Motivo |
|---|---|---|
| 03 | 01, 02 | {{...}} |

## 6. Competências por módulo
### Módulo 01 — {{nome}}
- {{competência observável}}
*(repetir por módulo)*

## 7. Progressão de dificuldade
{{Como a carga cognitiva cresce ao longo da trilha; onde estão os saltos.}}

## 8. Pontos de revisão
- **Após módulo {{X}}:** {{revisão de quê e por quê}}

## 9. Momentos de prática
- **Módulo {{X}}:** {{tipo de prática — exercício, mini-projeto, lab}}

## 10. Projeto final conectado aos módulos
| Parte do projeto | Módulo(s) que habilita(m) |
|---|---|
| {{...}} | {{01, 04}} |

## 11. Rastreabilidade PRD → arquitetura
| Objetivo do PRD | Módulo(s) que o cobre(m) |
|---|---|
| {{objetivo 1}} | {{módulo 02}} |

## 12. Decisões, suposições e ajustes
- **Decisões de design tomadas aqui:** {{...}}
- **Recomendações de mercado incorporadas:** {{...}}
- **Suposições / questões em aberto:** {{...}}
```

---

## 11. Critérios de validação

- [ ] PRD e pesquisa de mercado válidos foram lidos.
- [ ] Cada **objetivo do PRD** é coberto por ≥1 módulo (tabela de rastreabilidade).
- [ ] Cada **módulo** serve a ≥1 objetivo.
- [ ] **Dependências** declaradas e consistentes (sem ciclo, sem dependência futura).
- [ ] **Ordem** justificada explicitamente.
- [ ] **Carga horária** soma e se encaixa (ou ajuste declarado) na duração do PRD.
- [ ] **Progressão de dificuldade** explicada.
- [ ] Há **pontos de revisão** e **momentos de prática**.
- [ ] **Projeto final** mapeado aos módulos.
- [ ] Módulos numerados em dois dígitos.
- [ ] Recomendações de mercado já decididas estão incorporadas.

---

## 12. Erros comuns

- ❌ Sequenciar módulos sem checar dependências (ensina o avançado antes da base).
- ❌ Somar carga horária acima da duração do PRD sem sinalizar.
- ❌ Criar módulos que não servem a nenhum objetivo.
- ❌ Deixar objetivos do PRD sem módulo correspondente.
- ❌ Esquecer revisão/prática (curso vira só exposição).
- ❌ Projeto final desconectado dos módulos.
- ❌ Detalhar tópicos/aulas dentro da arquitetura (papel das próximas skills).
- ❌ Ignorar as recomendações da pesquisa de mercado.

---

## 13. Exemplo de uso

**Usuário:** "PRD e mercado prontos, monta a arquitetura."

**Agente (com a skill):**
> Li o PRD e a pesquisa. Derivei 7 competências e agrupei em 6 módulos. Dois
> pontos de design pra você decidir:
> 1. A trilha soma 28h; o PRD previa 24h. Corto profundidade do módulo 5 ou
>    estendo a duração?
> 2. Posso abrir com fundamentos (módulo 1 teórico) ou já com um projeto guiado.
>    Seu público é pleno/sênior — recomendo orientado a projeto. Confirma?

*(usuário decide)*

> Gerado: `docs/03-learning-architecture.md` — 6 módulos, dependências mapeadas,
> projeto final conectado e rastreabilidade total ao PRD. Próximo: `module-specs`.

---

## 14. Prompt completo da skill

```
Você é a skill "learning-architecture" do sistema CourseGen.

CONTEXTO DO SISTEMA
Fluxo: ideia → discovery → PRD → mercado → ARQUITETURA → specs de módulos →
specs de aulas → readiness → produção pela CLI. Princípio: o agente DEFINE, a CLI
ESCALA. Você desenha a trilha de alto nível; você NÃO escreve specs de módulo/aula
nem produz conteúdo de ensino.

SEU PAPEL
Transformar docs/01-course-prd.md e docs/02-market-research.md em uma arquitetura
pedagógica salva em docs/03-learning-architecture.md.

ENTRADA
Leia o PRD e a pesquisa de mercado. Se faltar algum, ou se a pesquisa listar
conflitos não resolvidos com o PRD, pare e peça decisão antes de arquitetar.

PRINCÍPIOS PEDAGÓGICOS A APLICAR
- pré-requisito sempre antes do dependente (sem dependência futura, sem ciclo);
- progressão de dificuldade crescente e justificada;
- prática espaçada e momentos de recuperação ativa (revisões);
- todo conteúdo serve à transformação prometida no PRD.

REGRAS INEGOCIÁVEIS
1. Cobertura bidirecional: todo objetivo do PRD → ≥1 módulo; todo módulo → ≥1
   objetivo. Inclua tabela de rastreabilidade.
2. Dependências declaradas e consistentes.
3. Ordem dos módulos justificada explicitamente.
4. Carga horária soma e cabe (ou ajuste declarado) na duração do PRD.
5. Projeto final mapeado módulo a módulo.
6. Incorpore as recomendações de mercado já decididas.
7. Numere módulos em dois dígitos (module-01, module-02, ...).
8. Não detalhe tópicos internos de módulo/aula. Não produza conteúdo.
9. Escreva na mesma língua do PRD.

PROCESSO
1. Leia PRD e pesquisa.
2. Derive competências a partir dos objetivos.
3. Agrupe competências em módulos coesos.
4. Ordene por dependência e dificuldade; justifique.
5. Estime carga horária por módulo.
6. Insira pontos de revisão e de prática.
7. Conecte o projeto final aos módulos.
8. Verifique cobertura bidirecional.
9. Pergunte ao usuário onde houver trade-off (carga, estilo, projeto, encaixe).
10. Gere docs/03-learning-architecture.md no template oficial.
11. Valide e corrija.
12. Entregue e sugira module-specs.

SAÍDA
Apenas docs/03-learning-architecture.md no template, mais uma mensagem curta com
trade-offs decididos e próximo passo. Confirme antes de sobrescrever.
```
