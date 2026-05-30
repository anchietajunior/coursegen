---
name: module-specs
description: Gera uma especificação detalhada para cada módulo do curso, a partir do PRD, da pesquisa de mercado e da arquitetura pedagógica.
phase: 05-module-specs
reads:
  - docs/01-course-prd.md
  - docs/02-market-research.md
  - docs/03-learning-architecture.md
writes:
  - docs/04-module-specs/module-XX.md
agent_agnostic: true
---

# Skill: Module Specs

> Princípio central do CourseGen: **o agente define, a CLI escala.**
> Esta skill detalha cada módulo definido pela arquitetura. Um módulo bem
> especificado é o que permite a CLI produzir o conteúdo depois, em escala.

---

## 1. Nome da skill

`module-specs`

---

## 2. Propósito

Gerar uma **especificação completa para cada módulo** previsto na arquitetura
pedagógica. Cada module spec descreve o que o módulo entrega: objetivo,
pré-requisitos, competências, tópicos dentro e fora de escopo, boas práticas,
erros comuns, exercícios esperados, projeto associado e critérios de conclusão.

Essas specs são o nível intermediário entre a arquitetura (alto nível) e as
lesson specs (detalhe fino). São o contrato de cada módulo: claro o bastante para
que outra pessoa — ou a CLI — produza o conteúdo sem precisar adivinhar intenção.

---

## 3. Quando usar

- Existe `docs/03-learning-architecture.md` válido, com módulos numerados.
- O usuário quer detalhar cada módulo antes de descer às aulas.
- Antes de `lesson-specs`.

---

## 4. Quando não usar

- Falta a arquitetura → rode `learning-architecture`.
- A arquitetura ainda tem trade-offs em aberto (carga, ordem, projeto) → resolva.
- O usuário quer detalhar aulas → use `lesson-specs`.
- O usuário quer produzir conteúdo/exercícios/slides → fora desta fase.

---

## 5. Inputs obrigatórios

| Input | Origem | Obrigatório |
|---|---|---|
| `docs/01-course-prd.md` | Skill `course-prd` | ✅ Sim |
| `docs/02-market-research.md` | Skill `market-research` | ✅ Sim |
| `docs/03-learning-architecture.md` | Skill `learning-architecture` | ✅ Sim |

---

## 6. Outputs gerados

| Output | Caminho | Formato |
|---|---|---|
| Spec por módulo | `docs/04-module-specs/module-XX.md` (um arquivo por módulo) | Markdown |

> `XX` = número do módulo em dois dígitos, exatamente como na arquitetura
> (`module-01.md`, `module-02.md`, ...). Um arquivo por módulo; nunca um arquivo
> único com todos.

---

## 7. Processo interno

1. **Ler** PRD, pesquisa de mercado e arquitetura.
2. **Extrair a lista de módulos** da arquitetura (número, nome, competências, carga, dependências).
3. **Para cada módulo**, produzir uma spec:
   - copiar objetivo e competências da arquitetura, refinando;
   - resolver pré-requisitos a partir das dependências declaradas;
   - listar **tópicos** que entregam as competências;
   - definir **tópicos fora de escopo** (o que o módulo deliberadamente não cobre);
   - extrair **boas práticas** e **erros comuns** relevantes (usando a pesquisa de mercado);
   - definir **exercícios esperados** (tipos, não os enunciados completos);
   - amarrar ao **projeto associado** (parte do projeto final, conforme a arquitetura);
   - escrever **critérios de conclusão** verificáveis do módulo.
4. **Garantir consistência** entre módulos: nenhum tópico de um módulo depende de algo só ensinado depois.
5. **Verificar cobertura:** as competências da arquitetura para o módulo estão todas endereçadas.
6. **Gerar** um arquivo por módulo em `docs/04-module-specs/`.
7. **Validar** cada spec contra a seção 11.
8. **Entregar** um resumo da geração e sugerir `lesson-specs`.

---

## 8. Perguntas que deve fazer ao usuário

A skill é majoritariamente derivacional. Pergunte apenas para resolver ambiguidade que a arquitetura não fechou:

- Se um módulo tiver competências amplas demais para um escopo: *"O módulo 04
  ficou grande. Divido em 04a/04b ou mantenho com menos profundidade?"*
- Se faltar definição de exercícios para um módulo prático: *"Que tipo de
  exercício combina com o módulo X — desafio de código, code review, lab guiado?"*
- Se a contribuição do módulo ao projeto final não estiver clara: *"O módulo Y
  alimenta qual parte do projeto final? Quero amarrar explicitamente."*

Se preferir gerar todos de uma vez, confirme antes: *"Gero as specs de todos os
{{N}} módulos agora, ou um por vez para você revisar?"*

---

## 9. Regras de execução

- **NÃO** gere specs sem arquitetura válida.
- **Um arquivo por módulo**, nomeado `module-XX.md` com `XX` da arquitetura.
- **Não invente módulos** que não existem na arquitetura, nem omita módulos que existem.
- Cada spec deve **cobrir todas as competências** do módulo segundo a arquitetura.
- **Tópicos fora de escopo** são obrigatórios (evita módulo inchar).
- **Critérios de conclusão** devem ser verificáveis (não "entendeu o assunto").
- **Exercícios** aqui são **tipos e intenção**, não enunciados completos (isso é fase de produção).
- **NÃO** escreva o conteúdo das aulas (papel de `lesson-specs` e da produção).
- Respeite dependências: nada de pré-requisito ensinado depois.
- Escreva na mesma língua dos documentos de origem.

---

## 10. Template do arquivo gerado

```markdown
# Module Spec — Módulo {{XX}}: {{Nome do módulo}}

> Documento gerado pela skill `module-specs`.
> Fontes: docs/01-course-prd.md, docs/02-market-research.md, docs/03-learning-architecture.md
> Status: [ ] Rascunho  [ ] Validado
> Data: {{AAAA-MM-DD}}

## 1. Objetivo do módulo
{{O que o aluno consegue fazer ao final deste módulo (verbo observável).}}

## 2. Pré-requisitos
> Derivados das dependências na arquitetura.
- {{Módulo XX / conhecimento prévio}}

## 3. Competências desenvolvidas
- {{competência observável — espelha a arquitetura}}

## 4. Tópicos (dentro de escopo)
1. {{tópico}}
2. {{tópico}}

## 5. Tópicos fora de escopo
- {{o que este módulo NÃO cobre — e onde é coberto, se for}}

## 6. Boas práticas
> Ancoradas no que o mercado espera (ver market-research).
- {{...}}

## 7. Erros comuns dos alunos
- {{erro — por que acontece — como prevenir}}

## 8. Exercícios esperados
> Tipos e intenção, não enunciados completos.
- {{tipo de exercício — competência que avalia}}

## 9. Projeto associado
- **Contribuição para o projeto final:** {{qual parte este módulo habilita}}
- **Entregável do módulo (se houver):** {{...}}

## 10. Critérios de conclusão do módulo
> Verificáveis. Como saber que o aluno concluiu este módulo.
- [ ] {{...}}

## 11. Carga horária estimada
{{X h — conforme arquitetura}}

## 12. Rastreabilidade
- **Objetivo(s) do PRD atendido(s):** {{...}}
- **Competências da arquitetura cobertas:** {{...}}
- **Suposições / questões em aberto:** {{...}}
```

---

## 11. Critérios de validação

Por **cada** module spec:

- [ ] Arquivo nomeado `module-XX.md` com `XX` correspondente à arquitetura.
- [ ] Objetivo do módulo com verbo observável.
- [ ] Pré-requisitos consistentes com as dependências da arquitetura.
- [ ] **Todas** as competências do módulo (na arquitetura) estão cobertas.
- [ ] Tópicos dentro de escopo entregam as competências.
- [ ] Existe seção **fora de escopo** não-vazia.
- [ ] Boas práticas e erros comuns presentes.
- [ ] Exercícios esperados definidos como tipos/intenção.
- [ ] Projeto associado amarrado ao projeto final.
- [ ] Critérios de conclusão **verificáveis**.
- [ ] Rastreabilidade ao PRD preenchida.

Global:
- [ ] Existe **uma spec por módulo** da arquitetura (nem a mais, nem a menos).
- [ ] Nenhum módulo usa pré-requisito ensinado só depois.

---

## 12. Erros comuns

- ❌ Gerar um único arquivo com todos os módulos.
- ❌ Inventar ou omitir módulos em relação à arquitetura.
- ❌ Esquecer a seção de fora de escopo.
- ❌ Critérios de conclusão vagos ("aluno entende X").
- ❌ Escrever enunciados completos de exercícios (isso é produção).
- ❌ Escrever o conteúdo da aula dentro da module spec.
- ❌ Numeração de arquivo divergente da arquitetura.
- ❌ Deixar competências da arquitetura sem cobertura na spec.

---

## 13. Exemplo de uso

**Usuário:** "Gera as specs dos módulos."

**Agente (com a skill):**
> Li a arquitetura: 6 módulos numerados. Gero as 6 specs de uma vez ou prefere
> revisar uma a uma? E confirme: o módulo 04 (orquestração) alimenta qual parte
> do projeto final — o pipeline ou o deploy?

*(usuário: "uma a uma; orquestração = pipeline")*

> Gerado `docs/04-module-specs/module-01.md`. Objetivo, competências, tópicos,
> fora de escopo, boas práticas, erros comuns, exercícios esperados, projeto
> associado e 3 critérios de conclusão verificáveis. Sigo para o module-02?

---

## 14. Prompt completo da skill

```
Você é a skill "module-specs" do sistema CourseGen.

CONTEXTO DO SISTEMA
Fluxo: ideia → discovery → PRD → mercado → arquitetura → SPECS DE MÓDULOS →
specs de aulas → readiness → produção pela CLI. Princípio: o agente DEFINE, a CLI
ESCALA. Você detalha cada módulo; você NÃO escreve o conteúdo das aulas nem
produz exercícios/slides completos.

SEU PAPEL
Gerar uma spec por módulo em docs/04-module-specs/module-XX.md, a partir de
docs/01-course-prd.md, docs/02-market-research.md e docs/03-learning-architecture.md.

ENTRADA
Leia os três documentos. Se a arquitetura estiver ausente ou com trade-offs em
aberto, pare e peça resolução. Extraia da arquitetura a lista de módulos
(número, nome, competências, carga, dependências).

REGRAS INEGOCIÁVEIS
1. Um arquivo por módulo, nomeado module-XX.md com XX exatamente como na
   arquitetura. Nunca um arquivo único.
2. Não invente nem omita módulos: a lista vem da arquitetura.
3. Cada spec cobre TODAS as competências do módulo segundo a arquitetura.
4. Seção "fora de escopo" é obrigatória.
5. Critérios de conclusão devem ser verificáveis.
6. Exercícios = tipos e intenção, não enunciados completos.
7. Não escreva conteúdo de aula.
8. Respeite dependências: nada de pré-requisito ensinado depois.
9. Escreva na mesma língua dos documentos de origem.

PROCESSO
1. Leia PRD, mercado e arquitetura.
2. Extraia a lista de módulos.
3. Pergunte se gera todos de uma vez ou um a um; resolva ambiguidades pontuais.
4. Para cada módulo, produza a spec no template oficial (objetivo, pré-requisitos,
   competências, tópicos, fora de escopo, boas práticas, erros comuns, exercícios
   esperados, projeto associado, critérios de conclusão, carga, rastreabilidade).
5. Garanta consistência de dependências entre módulos.
6. Verifique cobertura de competências.
7. Valide cada spec; corrija.
8. Entregue um resumo e sugira lesson-specs.

SAÍDA
Arquivos docs/04-module-specs/module-XX.md (um por módulo) no template, mais uma
mensagem curta com resumo e próximo passo. Confirme antes de sobrescrever.
```
