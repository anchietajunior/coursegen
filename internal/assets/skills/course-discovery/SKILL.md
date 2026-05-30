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
3. **Entrevistar UMA pergunta por vez** (ver seção 8), seguindo a ordem temática.
   - Faça **uma única pergunta** e então **pare e espere** a resposta. Nunca
     envie a próxima pergunta na mesma mensagem nem despeje um bloco inteiro.
   - **Ofereça opções (a, b, c, …)** sempre que a pergunta admitir alternativas
     razoáveis, mais uma opção aberta ("outro — descreva"). Isso facilita o input.
   - Para cada resposta vaga, refine com **um** follow-up (também uma pergunta só)
     antes de avançar.
   - Nunca escolha por ele em decisões estratégicas (público, objetivo,
     transformação) — as opções servem para guiar, não para decidir no lugar dele.
   - Indique o progresso (ex.: "Pergunta 3 de ~17") para o usuário se situar.
4. **Resumir e confirmar.** Antes de gerar, apresente um resumo das decisões e
   peça confirmação ou ajustes ("Confirma estes pontos antes de eu gerar o doc?").
5. **Detectar lacunas.** Se algum dos 17 campos obrigatórios ficou sem resposta,
   volte e pergunte. Não preencha com suposição silenciosa.
6. **Gerar o documento** usando o template da seção 10.
7. **Validar** contra a seção 11. Se algo falhar, corrija antes de entregar.
8. **Entregar e indicar o próximo passo:** sugerir rodar `course-prd`.

---

## 8. Perguntas que deve fazer ao usuário

Faça **UMA pergunta por vez**, na ordem abaixo, e **espere a resposta** antes de
seguir para a próxima. Os blocos A–F são apenas o **agrupamento temático e a
ordem** — **não** são lotes para enviar de uma vez. Sempre que houver
alternativas razoáveis, ofereça **opções (a, b, c, …)** + uma opção aberta, para
o usuário responder rápido (pode responder só a letra). As opções abaixo são
sugestões de partida; adapte-as ao tema do curso.

### Bloco A — Aluno e mercado

**1. Público-alvo** — quem é o aluno (cargo, momento de carreira, contexto)?
- a) Iniciante em transição para a área (primeiro emprego / bootcamp)
- b) Dev júnior/pleno querendo subir de nível
- c) Dev sênior/especialista entrando numa stack ou tema novo
- d) Profissional não-dev (PM, analista, gestor) que precisa entender o tema
- e) Outro — descreva

**2. Nível técnico de entrada** — o que ele já sabe ao começar?
- a) Iniciante (lógica básica, pouca ou nenhuma experiência na stack)
- b) Intermediário (já programa, conhece o básico do ecossistema)
- c) Avançado (experiente, busca tópicos específicos/profundos)
- (e diga, em 1 linha, o que ele JÁ domina ao entrar)

**3. Problemas reais** — que dor do dia a dia deste aluno o curso resolve?
(Ex.: "perde tempo com X", "não consegue Y", "trava em Z".) Liste 1 a 3.

**4. Mercado-alvo** — geografia, idioma e tipo de empresa.
- a) Brasil / pt-BR
- b) Global / inglês
- c) Nicho específico (segmento, porte de empresa — descreva)
- d) Outro — descreva

### Bloco B — Resultado e transformação

**5. Objetivo principal** em UMA frase. (Ex.: "Capacitar o aluno a fazer X de
forma autônoma.") Se quiser, posso sugerir 2–3 formulações para você escolher.

**6. Transformação esperada** no formato antes → depois:
- **Antes:** o que o aluno NÃO consegue fazer hoje?
- **Depois:** o que ele conseguirá ao concluir?

**7. Diferencial** — por que este curso e não os que já existem? Ângulos comuns:
- a) Mais prático / mão na massa que os concorrentes
- b) Atualizado com o que o mercado usa hoje
- c) Foco num nicho/stack específico mal atendido
- d) Didática ou projeto final únicos
- e) Outro — descreva

### Bloco C — Forma e duração

**8. Duração total** desejada.
- a) Curto (até ~10h / curso objetivo)
- b) Médio (~10–30h)
- c) Longo (30h+ / bootcamp)
- d) Outro — descreva (em horas ou semanas)

**9. Formato** de entrega.
- a) Vídeo assíncrono (gravado)
- b) Hands-on / ao vivo
- c) Misto (vídeo + prática guiada)
- d) Mentoria / acompanhamento
- e) Outro — descreva

**10. Profundidade teórica.**
- a) Conceitual leve (o suficiente para aplicar)
- b) Aplicada (teoria sempre amarrada à prática)
- c) Acadêmica / aprofundada (fundamentos, por baixo do capô)

**11. Profundidade prática.**
- a) Demos e exemplos guiados
- b) Projetos guiados (passo a passo)
- c) Projeto real do zero (o aluno constrói algo próprio)

### Bloco D — Tecnologias

**12. Tecnologias obrigatórias** — quais DEVEM aparecer no curso (linguagens,
frameworks, ferramentas)? Liste por nome.

**13. Tecnologias opcionais / bônus** — o que entra "se der tempo"? (ou "nenhuma")

**14. Pré-requisitos** — o que o aluno precisa ter/saber ANTES de começar?
- a) Nenhum (o curso parte do zero)
- b) Lógica de programação básica
- c) Uma linguagem específica (qual?)
- d) Outro — descreva

### Bloco E — Prática e avaliação

**15. Projeto final** — o que o aluno entrega ao fim do curso? (Ex.: uma API, um
app, um agente, um pipeline de dados.) Descreva o entregável.

**16. Tipos de exercício** (pode escolher vários):
- a) Quizzes / múltipla escolha
- b) Desafios de código
- c) Code review / análise de código
- d) Projeto incremental por módulo
- e) Outro — descreva

### Bloco F — Limites

**17. Fora de escopo** — o que o curso explicitamente NÃO vai abordar? (evita que
o curso vire infinito). Liste o que fica de fora.

> Regra de follow-up: se uma resposta for genérica ("para devs", "nível médio",
> "ser melhor"), faça **um** follow-up (uma pergunta só) até a decisão ficar
> concreta e mensurável — depois siga para a próxima pergunta.

---

## 9. Regras de execução

- **NUNCA** gere o documento antes de concluir a entrevista e confirmar o resumo.
- **SEMPRE** faça **uma pergunta por vez** e **espere a resposta** antes de enviar
  a próxima. Nunca agrupe várias perguntas numa única mensagem.
- **SEMPRE** ofereça **opções (a, b, c, …)** quando a pergunta admitir
  alternativas, incluindo uma opção aberta ("outro — descreva"). O usuário pode
  responder apenas com a letra.
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
> Ótimo! Vou te entrevistar **uma pergunta por vez** antes de gerar qualquer
> documento — o doc só nasce no final. São ~17 perguntas curtas.
>
> **Pergunta 1 de ~17 — Público-alvo:** quem é o aluno?
> - a) Iniciante em transição para a área
> - b) Dev júnior/pleno querendo subir de nível
> - c) Dev sênior/especialista entrando num tema novo
> - d) Profissional não-dev que precisa entender o tema
> - e) Outro — descreva

*(espera a resposta)*

**Usuário:** "b"

**Agente:**
> Anotado: dev júnior/pleno. **Pergunta 2 de ~17 — Nível técnico de entrada:**
> - a) Iniciante  b) Intermediário  c) Avançado
>
> E, em 1 linha, o que ele já domina ao entrar?

*(... a entrevista segue UMA pergunta por vez, com opções, até a 17 ...)*

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
2. Faça UMA pergunta por vez e PARE para esperar a resposta. Nunca agrupe várias
   perguntas na mesma mensagem nem despeje um bloco inteiro.
3. Sempre que a pergunta admitir alternativas, ofereça opções rotuladas (a, b,
   c, …) mais uma opção aberta ("outro — descreva"); o usuário pode responder só
   com a letra. As opções guiam, não decidem por ele em escolhas estratégicas.
4. Refine respostas vagas com UM follow-up (uma pergunta só) até virarem decisões
   concretas, e então siga para a próxima.
5. Cubra os 17 campos obrigatórios. Lacuna vira pergunta, nunca suposição muda.
6. Registre suposições assumidas e perguntas em aberto em seções próprias.
7. Não desenhe módulos nem aulas. Isso é fase posterior.
8. Escreva o documento na mesma língua da conversa.
9. O documento deve ser autossuficiente (legível sem esta conversa).

ENTREVISTA (UMA pergunta por vez, nesta ordem; os blocos são só o tema/ordem)
A. Aluno e mercado: 1) público-alvo, 2) nível técnico de entrada, 3) problemas
   reais resolvidos, 4) mercado-alvo.
B. Resultado: 5) objetivo principal (1 frase), 6) transformação (antes→depois),
   7) diferencial.
C. Forma: 8) duração, 9) formato, 10) profundidade teórica, 11) profundidade
   prática.
D. Tecnologias: 12) obrigatórias, 13) opcionais, 14) pré-requisitos do aluno.
E. Prática: 15) projeto final, 16) tipos de exercícios.
F. Limites: 17) fora de escopo.
Ofereça opções (a, b, c…) em perguntas com alternativas (público, nível, mercado,
diferencial, duração, formato, profundidades, pré-requisitos, tipos de exercício).

PROCESSO
1. Se o usuário não deu nem o tema, peça só o tema.
2. Anuncie que fará a entrevista UMA pergunta por vez antes de gerar o doc.
3. Conduza as perguntas 1→17 em ordem, uma por vez, com opções e follow-ups.
   Mostre o progresso (ex.: "Pergunta 3 de ~17").
4. Apresente um resumo das decisões e peça confirmação.
5. Cheque lacunas nos 17 campos; pergunte o que faltar (uma de cada vez).
6. Gere docs/00-course-discovery.md seguindo o template oficial da skill.
7. Valide contra os critérios; corrija o que falhar.
8. Entregue e sugira rodar a skill course-prd.

SAÍDA
Apenas o arquivo docs/00-course-discovery.md, no formato do template, mais uma
mensagem curta indicando o próximo passo. Confirme antes de sobrescrever um
arquivo existente.
```
