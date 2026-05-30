---
name: lesson-specs
description: Gera a especificação de cada aula de cada módulo, com objetivo, conceitos, demonstrações, exemplos de código esperados, exercícios, dúvidas prováveis, critérios de aceite e arquivos de saída.
phase: 06-lesson-specs
reads:
  - docs/01-course-prd.md
  - docs/03-learning-architecture.md
  - docs/04-module-specs/
writes:
  - docs/05-lesson-specs/module-XX/lesson-XX-YY.md
agent_agnostic: true
---

# Skill: Lesson Specs

> Princípio central do CourseGen: **o agente define, a CLI escala.**
> Esta é a última skill de definição antes do readiness. Lesson specs são o
> contrato mais fino do curso: é a partir delas que a CLI produzirá as aulas.

---

## 1. Nome da skill

`lesson-specs`

---

## 2. Propósito

Gerar a **especificação de cada aula** de cada módulo. Uma lesson spec descreve
exatamente o que a aula precisa entregar — objetivo, tempo, conceitos,
demonstrações, exemplo de código esperado, exercícios, dúvidas prováveis dos
alunos, erros comuns, boas práticas, critérios de aceite e os arquivos de saída.

Esta skill **não escreve a aula**. Ela escreve o que torna a aula produzível em
escala: um spec tão claro que a CLA (ou outro agente) consegue gerar a aula
completa sem reabrir nenhuma decisão. Aqui é a fronteira final entre *definir* e
*produzir*.

---

## 3. Quando usar

- Existem PRD, arquitetura e **todas** as module specs válidas.
- O usuário quer descer ao nível de aula antes do readiness check.
- Imediatamente antes de `course-readiness`.

---

## 4. Quando não usar

- Faltam module specs → rode `module-specs`.
- A arquitetura mudou e as module specs ainda não foram atualizadas → atualize primeiro.
- O usuário quer **produzir a aula completa** (texto, vídeo, código final) → isso
  é trabalho da CLI, depois do readiness. Esta skill só especifica.
- O usuário quer gerar slides ou exercícios prontos → fora desta fase.

---

## 5. Inputs obrigatórios

| Input | Origem | Obrigatório |
|---|---|---|
| `docs/01-course-prd.md` | Skill `course-prd` | ✅ Sim |
| `docs/03-learning-architecture.md` | Skill `learning-architecture` | ✅ Sim |
| `docs/04-module-specs/` (todas as specs) | Skill `module-specs` | ✅ Sim |

---

## 6. Outputs gerados

| Output | Caminho | Formato |
|---|---|---|
| Spec por aula | `docs/05-lesson-specs/module-XX/lesson-XX-YY.md` | Markdown |

> `XX` = número do módulo; `YY` = número da aula dentro do módulo (ambos dois
> dígitos). Ex.: a 3ª aula do módulo 02 → `docs/05-lesson-specs/module-02/lesson-02-03.md`.
> Uma pasta por módulo; um arquivo por aula.

---

## 7. Processo interno

1. **Ler** PRD, arquitetura e todas as module specs.
2. **Para cada módulo**, decompor os tópicos da module spec em uma **sequência de aulas**:
   - cada aula cobre um pedaço coeso de tópicos/competências;
   - a soma do tempo das aulas respeita a carga horária do módulo;
   - manter ordem lógica e progressão de dificuldade dentro do módulo.
3. **Para cada aula**, produzir a spec:
   - título e objetivo (verbo observável);
   - tempo estimado;
   - conceitos principais a ensinar;
   - demonstrações práticas previstas;
   - **descrição** do exemplo de código esperado (linguagem, o que demonstra) — não o código final;
   - exercícios esperados (tipo e intenção);
   - dúvidas prováveis dos alunos;
   - erros comuns relacionados;
   - boas práticas relacionadas;
   - critérios de aceite da aula (verificáveis);
   - arquivos de saída esperados (ex.: `aula.md`, `exemplo.py`, `exercicio.md`, `slides.md`).
4. **Verificar cobertura:** todos os tópicos do módulo viram aula; todas as competências do módulo são endereçadas por alguma aula.
5. **Gerar** os arquivos em `docs/05-lesson-specs/module-XX/`.
6. **Validar** contra a seção 11.
7. **Entregar** um resumo e sugerir `course-readiness`.

---

## 8. Perguntas que deve fazer ao usuário

- Sobre granularidade: *"Prefere aulas curtas (10–15 min, mais aulas) ou aulas
  mais longas (30–45 min, menos aulas)? Isso muda a contagem por módulo."*
- Sobre cobertura de um módulo grande: *"O módulo 04 tem muitos tópicos; posso
  quebrar em {{N}} aulas. Esse número parece bom?"*
- Sobre artefatos de saída: *"Que arquivos cada aula deve produzir na fase de
  produção (ex.: aula.md, exemplo de código, exercício, slides)? Uso esse
  conjunto como padrão?"*
- Sobre volume: *"Gero as specs de todas as aulas de todos os módulos agora, ou
  módulo por módulo para revisão?"*

---

## 9. Regras de execução

- **NÃO** gere lesson specs sem PRD, arquitetura e module specs válidas.
- **Estrutura de pastas obrigatória:** `docs/05-lesson-specs/module-XX/lesson-XX-YY.md`.
- **Numeração** de módulo e aula em dois dígitos, consistente com as fases anteriores.
- **Cobertura total:** todo tópico da module spec aparece em ≥1 aula; toda
  competência do módulo é endereçada.
- **Tempo das aulas** soma dentro da carga horária do módulo.
- **Exemplo de código** é **descrito** (intenção, linguagem, o que demonstra),
  **nunca escrito por completo** aqui.
- **Critérios de aceite** verificáveis (não "aula boa").
- **Arquivos de saída esperados** sempre listados — é o que a CLI vai gerar.
- **NÃO** escreva o conteúdo da aula, exercícios completos ou slides.
- Escreva na mesma língua dos documentos de origem.

---

## 10. Template do arquivo gerado

```markdown
# Lesson Spec — {{XX-YY}}: {{Título da aula}}

> Documento gerado pela skill `lesson-specs`.
> Módulo: {{XX}} — {{Nome do módulo}}
> Fontes: docs/01-course-prd.md, docs/03-learning-architecture.md, docs/04-module-specs/module-XX.md
> Status: [ ] Rascunho  [ ] Validado
> Data: {{AAAA-MM-DD}}

## 1. Título da aula
{{...}}

## 2. Objetivo da aula
{{O que o aluno consegue fazer ao final (verbo observável).}}

## 3. Tempo estimado
{{X min}}

## 4. Conceitos principais
- {{conceito}}

## 5. Demonstrações práticas
- {{o que será demonstrado ao vivo / em vídeo}}

## 6. Exemplo de código esperado
> Descrição, não o código final.
- **Linguagem/stack:** {{...}}
- **O que o exemplo demonstra:** {{...}}
- **Forma esperada:** {{snippet | arquivo executável | repositório mínimo}}

## 7. Exercícios esperados
> Tipo e intenção, não enunciado completo.
- {{tipo — competência avaliada}}

## 8. Dúvidas prováveis dos alunos
- {{pergunta provável — onde a aula deve antecipá-la}}

## 9. Erros comuns relacionados
- {{erro — sintoma — como a aula previne}}

## 10. Boas práticas relacionadas
- {{...}}

## 11. Critérios de aceite da aula
> Verificáveis. Quando a aula está "pronta o suficiente".
- [ ] {{...}}

## 12. Arquivos de saída esperados (para a produção pela CLI)
- `aula.md`
- `{{exemplo.ext}}`
- `exercicio.md`
- `slides.md`
- {{outros}}

## 13. Rastreabilidade
- **Tópicos da module spec cobertos:** {{...}}
- **Competência(s) endereçada(s):** {{...}}
- **Suposições / questões em aberto:** {{...}}
```

---

## 11. Critérios de validação

Por **cada** lesson spec:

- [ ] Caminho `docs/05-lesson-specs/module-XX/lesson-XX-YY.md` correto.
- [ ] Objetivo com verbo observável.
- [ ] Tempo estimado presente.
- [ ] Conceitos, demonstrações, exercícios, dúvidas, erros e boas práticas preenchidos.
- [ ] Exemplo de código **descrito** (não escrito por completo).
- [ ] Critérios de aceite **verificáveis**.
- [ ] Lista de **arquivos de saída esperados** não-vazia.
- [ ] Rastreabilidade à module spec preenchida.

Por **módulo**:
- [ ] Todos os tópicos da module spec aparecem em ≥1 aula.
- [ ] Todas as competências do módulo são endereçadas.
- [ ] Soma dos tempos das aulas ≈ carga horária do módulo.

---

## 12. Erros comuns

- ❌ Escrever a aula completa em vez de especificá-la.
- ❌ Escrever o código de exemplo final dentro da spec.
- ❌ Numeração/pastas fora do padrão `module-XX/lesson-XX-YY.md`.
- ❌ Deixar tópicos do módulo sem aula correspondente.
- ❌ Soma dos tempos das aulas estourar a carga do módulo sem avisar.
- ❌ Critérios de aceite vagos.
- ❌ Esquecer de listar os arquivos de saída esperados (a CLI precisa deles).
- ❌ Gerar slides/exercícios prontos (isso é produção).

---

## 13. Exemplo de uso

**Usuário:** "Specs de aula do módulo 02, por favor."

**Agente (com a skill):**
> O módulo 02 tem 6 tópicos e 4h de carga. Sugiro 6 aulas de ~30–40 min. Padrão
> de saída por aula: `aula.md`, `exemplo.py`, `exercicio.md`, `slides.md`.
> Confirma a granularidade e o conjunto de arquivos?

*(usuário confirma)*

> Gerados em `docs/05-lesson-specs/module-02/`: `lesson-02-01.md` ...
> `lesson-02-06.md`. Cada um com objetivo, conceitos, demo, exemplo de código
> descrito, exercícios, dúvidas prováveis, critérios de aceite e arquivos de
> saída. Cobertura: 6/6 tópicos, todas as competências do módulo. Próximo:
> `course-readiness` para liberar a produção pela CLI.

---

## 14. Prompt completo da skill

```
Você é a skill "lesson-specs" do sistema CourseGen.

CONTEXTO DO SISTEMA
Fluxo: ideia → discovery → PRD → mercado → arquitetura → specs de módulos →
SPECS DE AULAS → readiness → produção pela CLI. Princípio: o agente DEFINE, a CLI
ESCALA. Esta é a fronteira final entre definir e produzir: você ESPECIFICA cada
aula; você NÃO escreve a aula, o código final, os exercícios prontos ou os slides.

SEU PAPEL
Gerar uma spec por aula em docs/05-lesson-specs/module-XX/lesson-XX-YY.md, a
partir de docs/01-course-prd.md, docs/03-learning-architecture.md e de todas as
specs em docs/04-module-specs/.

ENTRADA
Leia PRD, arquitetura e todas as module specs. Se faltar module spec, pare e peça
para rodar module-specs.

REGRAS INEGOCIÁVEIS
1. Estrutura: docs/05-lesson-specs/module-XX/lesson-XX-YY.md (dois dígitos).
2. Cobertura total: todo tópico da module spec vira ≥1 aula; toda competência do
   módulo é endereçada por alguma aula.
3. A soma dos tempos das aulas respeita a carga horária do módulo.
4. O exemplo de código é DESCRITO (linguagem, o que demonstra), nunca escrito por
   completo.
5. Critérios de aceite verificáveis.
6. Sempre liste os arquivos de saída esperados (o que a CLI vai gerar).
7. Não escreva conteúdo de aula, exercícios completos ou slides.
8. Escreva na mesma língua dos documentos de origem.

PROCESSO
1. Leia PRD, arquitetura e module specs.
2. Combine com o usuário a granularidade das aulas e o conjunto padrão de
   arquivos de saída; pergunte se gera tudo ou módulo a módulo.
3. Para cada módulo, decomponha os tópicos em uma sequência de aulas coerente.
4. Para cada aula, produza a spec no template oficial (título, objetivo, tempo,
   conceitos, demonstrações, exemplo de código descrito, exercícios, dúvidas
   prováveis, erros comuns, boas práticas, critérios de aceite, arquivos de
   saída, rastreabilidade).
5. Verifique cobertura de tópicos e competências por módulo.
6. Valide cada spec; corrija.
7. Entregue um resumo e sugira course-readiness.

SAÍDA
Arquivos docs/05-lesson-specs/module-XX/lesson-XX-YY.md no template, mais uma
mensagem curta com cobertura e próximo passo. Confirme antes de sobrescrever.
```
