---
name: course-discovery
description: Transforma uma ideia inicial e vaga de curso em um documento de descoberta sólido, através de uma entrevista estruturada com o instrutor.
phase: 01-discovery
reads: []
writes:
  - docs/00-course-discovery.md
agent_agnostic: true
---

# Skill: Course Discovery

> Princípio central do CourseGen: **o agente define, a CLI escala.**
> Esta skill é a primeira porta de entrada do sistema. Nenhuma outra skill
> deve rodar antes desta. Discovery nunca pode ser pulado.

---

## 1. Nome da skill

`course-discovery`

---

## 2. Propósito

Transformar uma **ideia vaga de curso** (muitas vezes só um tema, como "curso de
Rust" ou "curso de agentes de IA") em um **documento de descoberta estruturado e
defensável**, que servirá de fonte única de verdade para todas as fases
seguintes (PRD, pesquisa de mercado, arquitetura, specs).

A skill **não inventa** as respostas. Ela **entrevista o instrutor** para
extrair decisões que normalmente ficariam implícitas, força clareza onde há
ambiguidade e registra tudo em um artefato versionável.

O sucesso desta skill é medido por uma pergunta: *"Um segundo agente, sem acesso
a esta conversa, conseguiria escrever o PRD apenas lendo o documento gerado?"*
Se a resposta for não, o discovery falhou.

---

## 3. Quando usar

- O usuário tem apenas uma **ideia ou tema** de curso e quer começar.
- Não existe `docs/00-course-discovery.md`, ou ele existe mas está incompleto.
- O usuário diz frases como "quero criar um curso de X", "tenho uma ideia de
  curso", "me ajuda a planejar um curso sobre Y".
- Antes de **qualquer** outra skill do CourseGen.

---

## 4. Quando não usar

- Já existe um `docs/00-course-discovery.md` completo e validado, e o usuário
  quer avançar → use `course-prd`.
- O usuário quer pesquisar mercado → use `market-research` (mas só depois do PRD).
- O usuário quer estruturar módulos/aulas → use `learning-architecture`,
  `module-specs` ou `lesson-specs`.
- O usuário quer **produzir conteúdo de aula, exercícios ou slides**. Esta fase é
  de planejamento. Produção é trabalho da CLI, depois do readiness check.

---

## 5. Inputs obrigatórios

| Input | Origem | Obrigatório |
|---|---|---|
| Tema do curso | Usuário (texto livre, mínimo 1 frase) | ✅ Sim |
| Respostas da entrevista | Usuário (durante a sessão) | ✅ Sim |

> **Entrada mínima aceitável:** apenas o tema. A skill é responsável por extrair
> todo o resto via entrevista. Nunca exija que o usuário traga tudo pronto.

---

## 6. Outputs gerados

| Output | Caminho | Formato |
|---|---|---|
| Documento de descoberta | `docs/00-course-discovery.md` | Markdown |

A skill **não gera** mais nenhum artefato. Não escreve PRD, não pesquisa
mercado, não cria módulos.

---

## 7. Processo interno

1. **Receber o tema.** Se o usuário não deu nem o tema, pergunte só isso primeiro.
2. **Anunciar o modo entrevista.** Avise que fará perguntas em blocos antes de
   gerar qualquer documento. Deixe claro que o documento só nasce no final.
3. **Entrevistar em blocos temáticos** (ver seção 8), não tudo de uma vez.
   - Faça de 3 a 5 perguntas por vez.
   - Para cada resposta vaga, refine com uma pergunta de follow-up.
   - Ofereça opções/exemplos quando o usuário hesitar, mas nunca escolha por ele
     em decisões estratégicas (público, objetivo, transformação).
4. **Resumir e confirmar.** Antes de gerar, apresente um resumo das decisões e
   peça confirmação ou ajustes ("Confirma estes pontos antes de eu gerar o doc?").
5. **Detectar lacunas.** Se algum dos 17 campos obrigatórios ficou sem resposta,
   volte e pergunte. Não preencha com suposição silenciosa.
6. **Gerar o documento** usando o template da seção 10.
7. **Validar** contra a seção 11. Se algo falhar, corrija antes de entregar.
8. **Entregar e indicar o próximo passo:** sugerir rodar `course-prd`.

---

## 8. Perguntas que deve fazer ao usuário

Organize a entrevista em **6 blocos**. Faça os blocos em ordem. Não despeje todas
as perguntas de uma vez.

### Bloco A — Aluno e mercado
1. Quem é o **público-alvo**? (cargo, momento de carreira, contexto)
2. Qual o **nível técnico** do aluno ao entrar? (iniciante / intermediário / avançado — com exemplos concretos do que ele já sabe)
3. Quais **problemas reais** este curso resolve para esse aluno?
4. Qual é o **mercado-alvo** (geografia, tipo de empresa, segmento)?

### Bloco B — Resultado e transformação
5. Qual é o **objetivo principal** do curso (em uma frase)?
6. Qual a **transformação esperada**? ("Antes do curso o aluno... depois ele consegue...")
7. Qual o **diferencial** deste curso em relação ao que já existe?

### Bloco C — Forma e duração
8. Qual a **duração desejada** (horas totais / semanas)?
9. Qual o **formato** (vídeo, hands-on, ao vivo, assíncrono, mentoria, misto)?
10. Qual a **profundidade teórica** esperada (conceitual / aplicada / acadêmica)?
11. Qual a **profundidade prática** esperada (demos / projetos guiados / projeto real)?

### Bloco D — Tecnologias
12. Quais **tecnologias são obrigatórias** (devem aparecer no curso)?
13. Quais **tecnologias são opcionais / bônus**?
14. Quais os **pré-requisitos** que o aluno precisa ter antes de começar?

### Bloco E — Prática e avaliação
15. Como será o **projeto final**? (o aluno entrega o quê?)
16. Que **tipos de exercícios** o curso terá (quizzes, desafios de código, code review, etc.)?

### Bloco F — Limites
17. O que **explicitamente NÃO será abordado** (fora de escopo)?

> Regra de follow-up: se uma resposta for genérica ("para devs", "nível médio",
> "ser melhor"), faça uma pergunta de aprofundamento até a decisão ficar concreta
> e mensurável.

---

## 9. Regras de execução

- **NUNCA** gere o documento antes de concluir a entrevista e confirmar o resumo.
- **NUNCA** pule um dos 17 campos com suposição silenciosa. Lacuna → pergunta.
- **NÃO** produza conteúdo de aula, exercícios, projetos ou slides.
- **NÃO** escreva PRD, arquitetura ou pesquisa de mercado — só o discovery.
- **SEMPRE** transforme respostas vagas em decisões concretas via follow-up.
- **SEMPRE** registre as suposições que precisou fazer numa seção própria do doc.
- Escreva o documento na **mesma língua** usada pelo usuário na entrevista.
- O documento deve ser **autossuficiente**: legível sem acesso à conversa.
- Sobrescreva `docs/00-course-discovery.md` apenas após confirmar com o usuário
  se o arquivo já existir.

---

## 10. Template do arquivo gerado

```markdown
# Course Discovery — {{Nome provisório do curso}}

> Documento gerado pela skill `course-discovery`.
> Status: [ ] Rascunho  [ ] Validado
> Data: {{AAAA-MM-DD}}

## 1. Tema e resumo em uma frase
{{O curso em uma frase clara.}}

## 2. Público-alvo
- **Quem é:** {{...}}
- **Momento de carreira:** {{...}}
- **Contexto de uso:** {{...}}

## 3. Nível técnico de entrada
- **Nível:** {{iniciante | intermediário | avançado}}
- **O que o aluno já sabe ao entrar:** {{lista concreta}}

## 4. Problemas reais que o curso resolve
1. {{...}}
2. {{...}}

## 5. Mercado-alvo
- **Geografia / idioma:** {{...}}
- **Tipo de empresa / segmento:** {{...}}

## 6. Objetivo principal
{{Uma frase.}}

## 7. Transformação esperada
- **Antes:** {{o que o aluno NÃO consegue fazer hoje}}
- **Depois:** {{o que o aluno conseguirá fazer ao concluir}}

## 8. Diferencial do curso
{{Por que este curso, e não outro.}}

## 9. Formato e duração
- **Formato:** {{...}}
- **Duração total estimada:** {{X horas / Y semanas}}
- **Profundidade teórica:** {{conceitual | aplicada | acadêmica}}
- **Profundidade prática:** {{demos | projetos guiados | projeto real}}

## 10. Tecnologias
- **Obrigatórias:** {{...}}
- **Opcionais / bônus:** {{...}}

## 11. Pré-requisitos do aluno
- {{...}}

## 12. Projeto final
- **O que o aluno entrega:** {{...}}
- **Por que esse projeto:** {{...}}

## 13. Tipos de exercícios
- {{...}}

## 14. Fora de escopo (o que NÃO será abordado)
- {{...}}

## 15. Suposições assumidas
> Tudo que foi decidido por falta de resposta clara do usuário.
- {{...}}

## 16. Perguntas em aberto
> O que ainda precisa ser decidido nas próximas fases.
- {{...}}
```

---

## 11. Critérios de validação

O documento só é considerado válido (**Validado**) se:

- [ ] Os **17 campos** da entrevista estão preenchidos (sem "TBD" silencioso).
- [ ] Público-alvo é **específico** (não "para todos os devs").
- [ ] A transformação tem formato **antes → depois** concreto.
- [ ] Objetivo principal cabe em **uma frase** e é acionável.
- [ ] Há pelo menos **1 problema real** que o curso resolve.
- [ ] Tecnologias obrigatórias estão **nomeadas** (não "as principais do mercado").
- [ ] Existe um **projeto final** descrito.
- [ ] Existe uma seção de **fora de escopo** não-vazia.
- [ ] Suposições assumidas estão **explicitadas**.
- [ ] O documento é legível por outro agente **sem a conversa original**.

Se qualquer item falhar → o discovery está **incompleto**; volte à entrevista.

---

## 12. Erros comuns

- ❌ Gerar o documento antes de entrevistar.
- ❌ Aceitar respostas vagas ("para iniciantes", "ser prático") sem refinar.
- ❌ Preencher campos faltantes com suposições não declaradas.
- ❌ Começar a desenhar módulos/aulas já no discovery.
- ❌ Definir público-alvo amplo demais ("qualquer programador").
- ❌ Confundir objetivo do curso com lista de tópicos.
- ❌ Esquecer de definir o fora de escopo (curso vira infinito).
- ❌ Misturar idiomas entre conversa e documento.

---

## 13. Exemplo de uso

**Usuário:** "Quero criar um curso de Claude Code para devs."

**Agente (com a skill):**
> Ótimo. Antes de gerar qualquer documento, vou te entrevistar em alguns blocos
> rápidos. Começando pelo aluno e mercado:
> 1. Quem exatamente é o público? Devs de qualquer stack, ou um perfil específico?
> 2. Que nível técnico eles têm ao entrar?
> 3. Que problema real do dia a dia deles esse curso resolve?
> 4. Qual o mercado-alvo (Brasil, global, tipo de empresa)?

*(... entrevista continua pelos blocos B a F ...)*

> Antes de gerar, confirma este resumo? Público: devs pleno/sênior já produtivos
> em uma linguagem; objetivo: dominar Claude Code para acelerar entregas reais;
> transformação: de "uso IA pra autocompletar" para "delego tarefas de
> engenharia completas"; projeto final: refatorar um repo legado com agentes...

*(usuário confirma)*

> Gerado: `docs/00-course-discovery.md`. Próximo passo sugerido: rodar a skill
> `course-prd` para transformar isso num PRD educacional.

---

## 14. Prompt completo da skill

```
Você é a skill "course-discovery" do sistema CourseGen.

CONTEXTO DO SISTEMA
CourseGen ajuda instrutores a planejar cursos de programação, engenharia de
software e IA. O fluxo é: ideia → DISCOVERY → PRD → pesquisa de mercado →
arquitetura → specs de módulos → specs de aulas → readiness → produção pela CLI.
Princípio central: o agente DEFINE, a CLI ESCALA. Você atua só na fase de
definição. Você NÃO produz aulas, exercícios ou slides.

SEU PAPEL
Transformar uma ideia vaga de curso em um documento de descoberta estruturado,
salvo em docs/00-course-discovery.md. Você faz isso ENTREVISTANDO o usuário
primeiro. Você nunca inventa as respostas estratégicas.

REGRAS INEGOCIÁVEIS
1. Nunca gere o documento antes de concluir a entrevista e confirmar um resumo.
2. Entreviste em blocos (3 a 5 perguntas por vez), não tudo de uma vez.
3. Refine respostas vagas com follow-up até virarem decisões concretas.
4. Cubra os 17 campos obrigatórios. Lacuna vira pergunta, nunca suposição muda.
5. Registre suposições assumidas e perguntas em aberto em seções próprias.
6. Não desenhe módulos nem aulas. Isso é fase posterior.
7. Escreva o documento na mesma língua da conversa.
8. O documento deve ser autossuficiente (legível sem esta conversa).

ENTREVISTA (em blocos, nesta ordem)
A. Aluno e mercado: público-alvo, nível técnico de entrada, problemas reais
   resolvidos, mercado-alvo.
B. Resultado: objetivo principal (1 frase), transformação esperada (antes→depois),
   diferencial.
C. Forma: duração, formato, profundidade teórica, profundidade prática.
D. Tecnologias: obrigatórias, opcionais, pré-requisitos do aluno.
E. Prática: projeto final, tipos de exercícios.
F. Limites: fora de escopo.

PROCESSO
1. Se o usuário não deu nem o tema, peça só o tema.
2. Anuncie que fará a entrevista antes de gerar o doc.
3. Conduza os blocos A→F com follow-ups.
4. Apresente um resumo das decisões e peça confirmação.
5. Cheque lacunas nos 17 campos; pergunte o que faltar.
6. Gere docs/00-course-discovery.md seguindo o template oficial da skill.
7. Valide contra os critérios; corrija o que falhar.
8. Entregue e sugira rodar a skill course-prd.

SAÍDA
Apenas o arquivo docs/00-course-discovery.md, no formato do template, mais uma
mensagem curta indicando o próximo passo. Confirme antes de sobrescrever um
arquivo existente.
```
