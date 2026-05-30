---
name: course-prd
description: Transforma o documento de descoberta em um PRD educacional defensável, com objetivos mensuráveis, escopo, critérios de sucesso e riscos.
phase: 02-prd
reads:
  - docs/00-course-discovery.md
writes:
  - docs/01-course-prd.md
agent_agnostic: true
---

# Skill: Course PRD

> Princípio central do CourseGen: **o agente define, a CLI escala.**
> O PRD é o contrato pedagógico do curso. Tudo que vier depois (mercado,
> arquitetura, módulos, aulas) precisa ser rastreável até este documento.

---

## 1. Nome da skill

`course-prd`

---

## 2. Propósito

Transformar o `docs/00-course-discovery.md` em um **PRD educacional** — um
Product Requirements Document adaptado para cursos. Ele converte as decisões
abertas do discovery em **requisitos verificáveis**: objetivos mensuráveis,
escopo explícito, critérios de sucesso, critérios de conclusão e riscos.

O PRD é a **fonte de verdade** que orienta as próximas fases. Se o discovery diz
"o quê" de forma exploratória, o PRD diz "o quê" de forma comprometida e medível.

---

## 3. Quando usar

- Existe um `docs/00-course-discovery.md` **completo e validado**.
- O usuário quer formalizar o curso antes de pesquisar mercado ou arquitetar.
- O usuário pede "criar o PRD", "formalizar o curso", "definir objetivos e escopo".

---

## 4. Quando não usar

- Não existe discovery, ou ele está incompleto → use `course-discovery` primeiro.
- O usuário quer validar o curso contra o mercado → use `market-research`
  (depois que o PRD existir).
- O usuário quer desenhar a sequência de módulos → use `learning-architecture`.
- O usuário quer produzir conteúdo de aula/exercícios/slides → fora desta fase.

---

## 5. Inputs obrigatórios

| Input | Origem | Obrigatório |
|---|---|---|
| `docs/00-course-discovery.md` | Skill `course-discovery` | ✅ Sim |
| Respostas a perguntas de lacuna | Usuário (se o discovery tiver buracos) | Condicional |

> Se o discovery estiver incompleto ou inválido, **não gere o PRD**. Liste o que
> falta e peça para rodar/rever o `course-discovery`.

---

## 6. Outputs gerados

| Output | Caminho | Formato |
|---|---|---|
| PRD educacional | `docs/01-course-prd.md` | Markdown |

---

## 7. Processo interno

1. **Ler** `docs/00-course-discovery.md` por completo.
2. **Validar a entrada:** todos os 17 campos do discovery presentes? Se não,
   pare e peça correção.
3. **Mapear** cada item do discovery para a seção correspondente do PRD.
4. **Transformar objetivos vagos em mensuráveis.** Cada objetivo do curso deve
   virar um objetivo com verbo de ação observável (ex.: "ao final, o aluno
   *configura* um pipeline de agentes que executa X com Y% de sucesso").
5. **Derivar critérios de sucesso e conclusão** a partir dos objetivos.
6. **Levantar riscos** pedagógicos e técnicos (ver seção 8).
7. **Fechar escopo e fora de escopo** com base no discovery, tornando-os
   explícitos e mutuamente exclusivos.
8. **Fazer perguntas de lacuna** somente onde o discovery deixou ambiguidade
   relevante para um requisito.
9. **Gerar** `docs/01-course-prd.md` pelo template da seção 10.
10. **Validar** contra a seção 11 e corrigir.
11. **Entregar** e sugerir `market-research`.

---

## 8. Perguntas que deve fazer ao usuário

O PRD **não reabre a entrevista**. Ele pergunta apenas onde há lacuna que impede
um requisito verificável. Faça no máximo o necessário:

- Se os objetivos do discovery não forem mensuráveis: *"Como saberemos,
  objetivamente, que o aluno atingiu o objetivo X? Que evidência ele produz?"*
- Se não houver critério de conclusão claro: *"O que define que um aluno
  CONCLUIU o curso — entregar o projeto? passar num desafio? que nota mínima?"*
- Se o público tiver mais de um perfil: *"Quero descrever 1 a 3 personas. Há um
  perfil secundário relevante além do principal?"*
- Se não houver métricas de sucesso de negócio/curso: *"Qual métrica diria que o
  curso teve sucesso (conclusão %, satisfação, empregabilidade, etc.)?"*

Se o usuário não souber responder, registre como **suposição** ou **risco**, não
trave o documento indefinidamente.

---

## 9. Regras de execução

- **NUNCA** gere o PRD se o discovery estiver ausente ou inválido.
- **TODO** objetivo deve ser mensurável (verbo observável + evidência).
- **Escopo** e **fora de escopo** devem ser explícitos e não se sobrepor.
- Não objetivos ≠ fora de escopo: *não objetivos* são metas que o curso
  conscientemente NÃO persegue; *fora de escopo* são tópicos/conteúdos excluídos.
- Toda decisão do PRD deve ser **rastreável** a um item do discovery ou marcada
  como nova decisão tomada nesta fase.
- **NÃO** desenhe módulos nem aulas (isso é arquitetura).
- **NÃO** produza conteúdo de ensino.
- Personas devem ser **concretas** (1 a 3), não demografia genérica.
- Escreva na mesma língua do discovery.

---

## 10. Template do arquivo gerado

```markdown
# Course PRD — {{Nome do curso}}

> Documento gerado pela skill `course-prd`.
> Fonte: docs/00-course-discovery.md
> Status: [ ] Rascunho  [ ] Validado
> Data: {{AAAA-MM-DD}}

## 1. Visão do curso
{{Parágrafo único: o que é o curso e por que ele existe.}}

## 2. Problema do aluno
{{A dor concreta que o aluno tem hoje e que motiva o curso.}}

## 3. Público-alvo
{{Descrição do público principal.}}

## 4. Personas
### Persona 1 — {{nome/rótulo}}
- **Perfil:** {{...}}
- **Objetivo ao fazer o curso:** {{...}}
- **Dor principal:** {{...}}
- **Nível de entrada:** {{...}}
*(Repetir para até 3 personas.)*

## 5. Objetivos mensuráveis
> Cada objetivo: verbo de ação observável + evidência verificável.
1. {{Ao final, o aluno CONSEGUE ... comprovado por ...}}
2. {{...}}

## 6. Não objetivos
> Metas que o curso conscientemente NÃO persegue.
- {{...}}

## 7. Escopo
> O que o curso COBRE.
- {{...}}

## 8. Fora de escopo
> Tópicos/conteúdos explicitamente EXCLUÍDOS.
- {{...}}

## 9. Resultado prometido
{{A promessa central feita ao aluno — a transformação garantida.}}

## 10. Critérios de sucesso (do curso)
> Como saberemos que o CURSO funcionou.
- {{Métrica + meta. Ex.: taxa de conclusão ≥ X%}}

## 11. Critérios de conclusão (do aluno)
> O que define que um ALUNO concluiu.
- {{...}}

## 12. Pré-requisitos
- {{...}}

## 13. Riscos pedagógicos
| Risco | Impacto | Mitigação |
|---|---|---|
| {{...}} | {{alto/médio/baixo}} | {{...}} |

## 14. Riscos técnicos
| Risco | Impacto | Mitigação |
|---|---|---|
| {{...}} | {{alto/médio/baixo}} | {{...}} |

## 15. Diferenciais de mercado
- {{...}}

## 16. Rastreabilidade e suposições
- **Decisões novas tomadas neste PRD:** {{...}}
- **Suposições assumidas:** {{...}}
- **Perguntas em aberto para o mercado/arquitetura:** {{...}}
```

---

## 11. Critérios de validação

- [ ] O PRD foi derivado de um discovery **válido**.
- [ ] Visão, problema e público estão preenchidos e coerentes com o discovery.
- [ ] Há **1 a 3 personas** concretas.
- [ ] **Todo** objetivo é mensurável (verbo observável + evidência).
- [ ] Existem **não objetivos** explícitos.
- [ ] **Escopo** e **fora de escopo** existem e não se sobrepõem.
- [ ] Há **critérios de sucesso do curso** com métrica.
- [ ] Há **critérios de conclusão do aluno**.
- [ ] Riscos pedagógicos e técnicos listados, cada um com mitigação.
- [ ] Diferenciais de mercado presentes.
- [ ] Nenhuma decisão importante ficou implícita.

Falha em qualquer item → PRD **incompleto**; corrija antes de seguir.

---

## 12. Erros comuns

- ❌ Escrever objetivos não mensuráveis ("entender", "conhecer", "dominar" sem evidência).
- ❌ Confundir não objetivos com fora de escopo.
- ❌ Copiar o discovery sem transformar em requisitos verificáveis.
- ❌ Personas genéricas ("desenvolvedor de 25 a 40 anos").
- ❌ Critérios de sucesso sem métrica.
- ❌ Listar riscos sem mitigação.
- ❌ Começar a planejar módulos/aulas dentro do PRD.
- ❌ Gerar o PRD com discovery incompleto.

---

## 13. Exemplo de uso

**Usuário:** "O discovery está pronto, gera o PRD."

**Agente (com a skill):**
> Li o `docs/00-course-discovery.md`. Dois pontos precisam virar mensuráveis
> antes do PRD:
> 1. O objetivo "dominar agentes de IA" — como provamos? Sugiro: "o aluno
>    constrói e implanta um agente que resolve a tarefa X com sucesso ≥ 90%".
> 2. Não há critério de conclusão. Proponho: "concluiu quem entrega o projeto
>    final aprovado no checklist Y." Confirma?

*(usuário confirma)*

> Gerado: `docs/01-course-prd.md` com 16 seções, objetivos mensuráveis, escopo,
> riscos e diferenciais. Próximo passo: rodar `market-research` para validar
> contra o mercado atual.

---

## 14. Prompt completo da skill

```
Você é a skill "course-prd" do sistema CourseGen.

CONTEXTO DO SISTEMA
Fluxo: ideia → discovery → PRD → mercado → arquitetura → specs de módulos →
specs de aulas → readiness → produção pela CLI. Princípio: o agente DEFINE, a CLI
ESCALA. Você atua na fase de definição. Você NÃO produz aulas, exercícios ou
slides, e NÃO desenha módulos/aulas (isso é arquitetura).

SEU PAPEL
Transformar docs/00-course-discovery.md em um PRD educacional salvo em
docs/01-course-prd.md.

ENTRADA
Leia docs/00-course-discovery.md inteiro. Se ele estiver ausente ou incompleto
(faltam campos obrigatórios do discovery), NÃO gere o PRD: liste o que falta e
peça para rodar/rever a skill course-discovery.

REGRAS INEGOCIÁVEIS
1. Todo objetivo do PRD deve ser MENSURÁVEL: verbo de ação observável +
   evidência verificável.
2. Escopo e fora de escopo devem existir e não se sobrepor.
3. Não objetivos (metas não perseguidas) são diferentes de fora de escopo
   (conteúdo excluído). Trate-os em seções separadas.
4. Toda decisão deve ser rastreável ao discovery ou marcada como decisão nova.
5. Personas concretas, de 1 a 3.
6. Critérios de sucesso do curso precisam de métrica; critérios de conclusão do
   aluno precisam ser verificáveis.
7. Riscos pedagógicos e técnicos sempre com mitigação.
8. Nenhuma decisão importante pode ficar implícita.
9. Escreva na mesma língua do discovery.

PROCESSO
1. Leia e valide o discovery.
2. Mapeie discovery → seções do PRD.
3. Converta objetivos vagos em mensuráveis.
4. Derive critérios de sucesso e de conclusão.
5. Levante riscos pedagógicos e técnicos com mitigações.
6. Feche escopo e fora de escopo explícitos.
7. Pergunte ao usuário SOMENTE onde houver lacuna que impeça um requisito
   verificável; o resto vira suposição/risco registrado.
8. Gere docs/01-course-prd.md no template oficial (16 seções).
9. Valide contra os critérios; corrija o que falhar.
10. Entregue e sugira rodar market-research.

SAÍDA
Apenas docs/01-course-prd.md no formato do template, mais uma mensagem curta com
o próximo passo. Confirme antes de sobrescrever arquivo existente.
```
