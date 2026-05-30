---
name: course-readiness
description: Verifica se todos os artefatos de planejamento estão completos e coerentes, e aprova ou reprova o curso para produção em escala pela CLI. É um gate — pode reprovar.
phase: 07-readiness
reads:
  - docs/00-course-discovery.md
  - docs/01-course-prd.md
  - docs/02-market-research.md
  - docs/03-learning-architecture.md
  - docs/04-module-specs/
  - docs/05-lesson-specs/
writes:
  - docs/06-course-readiness-checklist.md
agent_agnostic: true
---

# Skill: Course Readiness

> Princípio central do CourseGen: **o agente define, a CLI escala.**
> Esta skill é o **portão** entre as duas fases. Ela só deixa o curso passar para
> a produção pela CLI se a definição estiver completa, coerente e rastreável.
> Reprovar é uma saída legítima e esperada.

---

## 1. Nome da skill

`course-readiness`

---

## 2. Propósito

Verificar se o pacote de planejamento do curso está **pronto para produção em
escala pela CLI**. A skill audita todos os artefatos das fases anteriores,
checa completude e coerência entre eles, e emite um **veredito binário**:
APROVADO ou REPROVADO.

Se reprovar, **não basta dizer que falhou**: a skill lista exatamente o que está
faltando, em qual artefato, e as correções necessárias para aprovar. É o último
controle de qualidade antes de a CLI gastar tempo e tokens produzindo conteúdo
sobre uma base frágil.

---

## 3. Quando usar

- Todos os artefatos de planejamento existem (discovery → lesson specs).
- Antes de acionar a CLI para produzir aulas, exercícios, projetos e slides.
- Quando o usuário pergunta "o curso está pronto?", "posso mandar produzir?".
- Como gate de revisão recorrente após qualquer mudança no planejamento.

---

## 4. Quando não usar

- Ainda faltam artefatos de fases anteriores → rode a skill correspondente
  (a readiness vai apontar isso mesmo, mas é desperdício rodá-la cedo demais).
- O usuário quer **produzir** o conteúdo → isso é a CLI, e só depois de APROVADO.
- O usuário quer editar um artefato específico → use a skill daquela fase.

---

## 5. Inputs obrigatórios

| Input | Origem | Obrigatório |
|---|---|---|
| `docs/00-course-discovery.md` | `course-discovery` | ✅ Sim |
| `docs/01-course-prd.md` | `course-prd` | ✅ Sim |
| `docs/02-market-research.md` | `market-research` | ✅ Sim |
| `docs/03-learning-architecture.md` | `learning-architecture` | ✅ Sim |
| `docs/04-module-specs/` (todas) | `module-specs` | ✅ Sim |
| `docs/05-lesson-specs/` (todas) | `lesson-specs` | ✅ Sim |

> A ausência de qualquer um destes é, por si só, motivo de **REPROVAÇÃO**.

---

## 6. Outputs gerados

| Output | Caminho | Formato |
|---|---|---|
| Checklist de prontidão + veredito | `docs/06-course-readiness-checklist.md` | Markdown |

---

## 7. Processo interno

1. **Inventariar** os artefatos: existem todos? Quais módulos e aulas a
   arquitetura prevê vs. quais specs existem de fato?
2. **Checar completude** de cada artefato contra os critérios obrigatórios (seção 11).
3. **Checar coerência cruzada** (a parte mais importante):
   - todo objetivo do PRD é coberto por módulo(s) (arquitetura)?
   - todo módulo da arquitetura tem uma module spec?
   - todo tópico de cada module spec aparece em ≥1 lesson spec?
   - toda lesson spec tem objetivo e critério de aceite?
   - o projeto final está definido e conectado aos módulos?
   - escopo/fora de escopo, exercícios, tecnologias e riscos estão definidos?
4. **Classificar achados** em: bloqueadores (reprovam) e avisos (não reprovam,
   mas recomendam correção).
5. **Emitir veredito:** APROVADO somente se **zero bloqueadores**.
6. **Gerar** `docs/06-course-readiness-checklist.md` com o checklist marcado, a
   lista de correções e o veredito.
7. **Entregar:** se APROVADO, liberar para a CLI; se REPROVADO, listar as skills a
   re-rodar e as correções necessárias.

---

## 8. Perguntas que deve fazer ao usuário

A readiness é uma **auditoria**, não uma entrevista — ela não inventa nem decide
conteúdo. Pergunte só em casos pontuais:

- Se um item for ambíguo entre "ausente" e "intencionalmente fora de escopo":
  *"O módulo 05 não tem lesson specs. É um módulo opcional/bônus que decidimos
  não detalhar, ou faltou especificar?"*
- Se o usuário quiser **forçar** aprovação com avisos pendentes: *"Há 0
  bloqueadores e 3 avisos. Aprovo mesmo assim e registro os avisos como dívida,
  ou prefere corrigir antes?"* (bloqueadores nunca podem ser ignorados).

A skill **nunca** preenche um artefato faltante por conta própria; ela só audita.

---

## 9. Regras de execução

- **NÃO** APROVE se faltar qualquer artefato obrigatório.
- **APROVADO** exige **zero bloqueadores**. Bloqueadores não são negociáveis.
- **NÃO** edite nem complete os artefatos auditados — apenas relate.
- **SEMPRE** que reprovar, liste: o que falta, em qual arquivo, e qual skill re-rodar.
- **Separe** bloqueadores de avisos com clareza.
- **Cobertura é checagem central:** PRD↔arquitetura↔módulos↔aulas devem fechar.
- O veredito deve ser **explícito e binário** no topo do documento.
- **NÃO** produza conteúdo de aula/exercício/slide (nem aqui, nem nunca nesta fase).
- Escreva na mesma língua dos artefatos.

---

## 10. Template do arquivo gerado

```markdown
# Course Readiness Checklist — {{Nome do curso}}

> Documento gerado pela skill `course-readiness`.
> Data: {{AAAA-MM-DD}}

## Veredito
**[ APROVADO ✅ | REPROVADO ❌ ]**
> {{Frase de resumo: pronto para a CLI produzir, ou não — e por quê.}}
> Bloqueadores: {{N}} · Avisos: {{M}}

## 1. Inventário de artefatos
| Artefato | Existe? | Observação |
|---|---|---|
| 00-course-discovery.md | {{✅/❌}} | {{...}} |
| 01-course-prd.md | {{✅/❌}} | {{...}} |
| 02-market-research.md | {{✅/❌}} | {{...}} |
| 03-learning-architecture.md | {{✅/❌}} | {{...}} |
| 04-module-specs/ ({{X}}/{{Y}} módulos) | {{✅/❌}} | {{...}} |
| 05-lesson-specs/ ({{A}}/{{B}} aulas) | {{✅/❌}} | {{...}} |

## 2. Checklist de critérios obrigatórios
- [ ] Público-alvo definido
- [ ] Objetivos mensuráveis
- [ ] Módulos definidos
- [ ] Aulas definidas
- [ ] Cada aula tem objetivo claro
- [ ] Cada aula tem critério de aceite
- [ ] Projeto final definido
- [ ] Exercícios planejados
- [ ] Tecnologias definidas
- [ ] Escopo e fora de escopo definidos
- [ ] Riscos identificados
- [ ] Curso pronto para execução pela CLI

## 3. Coerência cruzada
| Verificação | Status | Detalhe |
|---|---|---|
| Todo objetivo do PRD → módulo(s) | {{✅/❌}} | {{...}} |
| Todo módulo da arquitetura → module spec | {{✅/❌}} | {{...}} |
| Todo tópico de module spec → ≥1 lesson spec | {{✅/❌}} | {{...}} |
| Toda lesson spec tem objetivo + critério de aceite | {{✅/❌}} | {{...}} |
| Projeto final conectado aos módulos | {{✅/❌}} | {{...}} |
| Carga horária coerente (PRD ↔ arquitetura ↔ aulas) | {{✅/❌}} | {{...}} |

## 4. Bloqueadores (reprovam o curso)
> Se a lista estiver vazia e não faltar artefato, o curso pode ser aprovado.
1. **[{{artefato}}]** {{o que falta}} → **Correção:** {{ação}} (re-rodar skill `{{...}}`)

## 5. Avisos (não reprovam, mas recomenda-se corrigir)
- {{...}}

## 6. Plano de correção (se REPROVADO)
| Ordem | Skill a re-rodar | Motivo |
|---|---|---|
| 1 | {{...}} | {{...}} |

## 7. Liberação para produção
- **Status:** {{LIBERADO para a CLI | BLOQUEADO}}
- {{Se liberado: o que a CLI pode começar a produzir e em que ordem.}}
```

---

## 11. Critérios de validação

Estes são os **critérios obrigatórios** que a skill audita. Cada um deve estar
satisfeito por algum artefato; qualquer falha é **bloqueador**:

- [ ] **Público-alvo definido** (discovery + PRD).
- [ ] **Objetivos mensuráveis** (PRD).
- [ ] **Módulos definidos** (arquitetura).
- [ ] **Aulas definidas** (lesson specs).
- [ ] **Cada aula tem objetivo claro** (lesson specs).
- [ ] **Cada aula tem critério de aceite** (lesson specs).
- [ ] **Projeto final definido** (PRD/arquitetura).
- [ ] **Exercícios planejados** (module/lesson specs).
- [ ] **Tecnologias definidas** (discovery/PRD/mercado).
- [ ] **Escopo e fora de escopo definidos** (PRD + cada spec).
- [ ] **Riscos identificados** (PRD).
- [ ] **Curso pronto para execução pela CLI** (todos os acima + coerência cruzada).

E a própria auditoria é válida se:
- [ ] O veredito é **binário e explícito** no topo.
- [ ] Cada bloqueador tem **arquivo + correção + skill a re-rodar**.
- [ ] Bloqueadores e avisos estão **separados**.
- [ ] Nenhum artefato foi editado pela skill.

---

## 12. Erros comuns

- ❌ Aprovar com artefato faltando.
- ❌ Tratar um bloqueador como simples aviso.
- ❌ Reprovar sem dizer o que corrigir e onde.
- ❌ Editar/completar os artefatos em vez de só auditar.
- ❌ Verificar completude isolada e esquecer a **coerência cruzada**.
- ❌ Veredito ambíguo ("quase pronto", "talvez").
- ❌ Não checar se cada tópico de módulo virou aula.
- ❌ Ignorar divergência de carga horária entre artefatos.

---

## 13. Exemplo de uso

**Usuário:** "O curso está pronto pra CLI produzir?"

**Agente (com a skill):**
> Auditei os 6 conjuntos de artefatos. Veredito: **REPROVADO ❌** — 2 bloqueadores:
> 1. **[05-lesson-specs/module-04]** as aulas 03 e 04 não têm critério de aceite
>    → re-rodar `lesson-specs` para o módulo 04.
> 2. **[03-learning-architecture]** o objetivo "implantar o agente em produção"
>    do PRD não está coberto por nenhum módulo → re-rodar `learning-architecture`.
>
> Há ainda 1 aviso (carga do módulo 02 ficou 1h acima do previsto). Gerei
> `docs/06-course-readiness-checklist.md` com o plano de correção. Quando os 2
> bloqueadores forem resolvidos, rode a readiness de novo para liberar a CLI.

---

## 14. Prompt completo da skill

```
Você é a skill "course-readiness" do sistema CourseGen.

CONTEXTO DO SISTEMA
Fluxo: ideia → discovery → PRD → mercado → arquitetura → specs de módulos →
specs de aulas → READINESS → produção pela CLI. Princípio: o agente DEFINE, a CLI
ESCALA. Você é o portão entre as duas fases. Você NÃO produz conteúdo e NÃO
edita os artefatos — você AUDITA e emite um veredito.

SEU PAPEL
Verificar se o planejamento está pronto para a CLI produzir o curso, e gerar
docs/06-course-readiness-checklist.md com checklist, achados e veredito binário.

ENTRADA
Leia: docs/00-course-discovery.md, docs/01-course-prd.md, docs/02-market-research.md,
docs/03-learning-architecture.md, todas as specs em docs/04-module-specs/ e todas
as specs em docs/05-lesson-specs/. A ausência de qualquer um é, por si só,
motivo de REPROVAÇÃO.

CRITÉRIOS OBRIGATÓRIOS (cada falha é bloqueador)
público-alvo definido; objetivos mensuráveis; módulos definidos; aulas definidas;
cada aula com objetivo claro; cada aula com critério de aceite; projeto final
definido; exercícios planejados; tecnologias definidas; escopo e fora de escopo
definidos; riscos identificados; curso pronto para execução pela CLI.

COERÊNCIA CRUZADA (também bloqueia se falhar)
- todo objetivo do PRD coberto por módulo(s) da arquitetura;
- todo módulo da arquitetura tem module spec;
- todo tópico de module spec aparece em ≥1 lesson spec;
- toda lesson spec tem objetivo e critério de aceite;
- projeto final conectado aos módulos;
- carga horária coerente entre PRD, arquitetura e aulas.

REGRAS INEGOCIÁVEIS
1. APROVADO exige ZERO bloqueadores. Falta de artefato = REPROVADO.
2. Veredito binário e explícito no topo do documento.
3. Cada bloqueador deve dizer: arquivo, o que falta, correção, skill a re-rodar.
4. Separe bloqueadores de avisos. Bloqueadores nunca viram avisos.
5. Não edite nem complete os artefatos. Apenas relate.
6. Não produza conteúdo de aula/exercício/slide.
7. Escreva na mesma língua dos artefatos.

PROCESSO
1. Inventarie os artefatos (existem todos? módulos/aulas previstos vs. existentes?).
2. Cheque completude de cada um contra os critérios obrigatórios.
3. Cheque a coerência cruzada.
4. Classifique achados em bloqueadores e avisos.
5. Emita o veredito (APROVADO só com zero bloqueadores).
6. Gere docs/06-course-readiness-checklist.md no template oficial.
7. Se APROVADO, libere para a CLI e indique o que produzir e em que ordem. Se
   REPROVADO, entregue o plano de correção (skills a re-rodar).

SAÍDA
Apenas docs/06-course-readiness-checklist.md no template, mais uma mensagem curta
com o veredito e o próximo passo. Confirme antes de sobrescrever.
```
