---
name: market-research
description: Avalia se o curso está alinhado ao mercado real — tecnologias usadas, tendências, vagas, entrevistas, obsolescências — e recomenda o que adicionar, remover ou ajustar.
phase: 03-market-research
reads:
  - docs/00-course-discovery.md
  - docs/01-course-prd.md
writes:
  - docs/02-market-research.md
agent_agnostic: true
---

# Skill: Market Research

> Princípio central do CourseGen: **o agente define, a CLI escala.**
> Esta skill confronta o curso planejado com a realidade do mercado, para que a
> arquitetura pedagógica seja construída sobre fatos, não sobre suposições.

---

## 1. Nome da skill

`market-research`

---

## 2. Propósito

Avaliar se o curso descrito no discovery e no PRD está **alinhado ao mercado
real** de programação, engenharia de software e IA. A skill investiga o que
empresas usam, o que aparece em vagas e entrevistas, o que está em ascensão e o
que está obsoleto — e converte isso em **recomendações acionáveis**: o que
manter, adicionar, remover ou repriorizar antes de arquitetar o curso.

O resultado não é um relatório de tendências solto: é um conjunto de decisões com
**impacto direto** no escopo e nas tecnologias do curso, sempre amarradas ao PRD.

---

## 3. Quando usar

- Existem `docs/00-course-discovery.md` e `docs/01-course-prd.md` válidos.
- Antes de desenhar a arquitetura pedagógica (`learning-architecture`).
- Quando o usuário quer validar relevância de mercado, atualizar tecnologias ou
  justificar escolhas para um curso comercial.

---

## 4. Quando não usar

- Não existe PRD → rode `course-prd` antes.
- O usuário só quer estruturar módulos/aulas e já decidiu o stack → pode pular
  para `learning-architecture` (mas registre que o mercado não foi validado).
- O usuário quer produzir conteúdo/exercícios/slides → fora desta fase.

---

## 5. Inputs obrigatórios

| Input | Origem | Obrigatório |
|---|---|---|
| `docs/00-course-discovery.md` | Skill `course-discovery` | ✅ Sim |
| `docs/01-course-prd.md` | Skill `course-prd` | ✅ Sim |
| Acesso a busca web (se disponível) | Ambiente do agente | Recomendado |
| Conhecimento do agente | Modelo | Fallback |

> Se não houver acesso a busca web, a skill **declara isso explicitamente** no
> documento, opera com o conhecimento do modelo e marca cada afirmação não
> verificada como *"não verificado em fonte externa"*.

---

## 6. Outputs gerados

| Output | Caminho | Formato |
|---|---|---|
| Pesquisa de mercado | `docs/02-market-research.md` | Markdown |

---

## 7. Processo interno

1. **Ler** discovery e PRD; extrair tema, público, mercado-alvo e tecnologias
   obrigatórias/opcionais.
2. **Definir as perguntas de pesquisa** a partir desse contexto (ex.: "quais
   ferramentas de agentes de IA aparecem em vagas de eng. de software em 2026?").
3. **Pesquisar** (web, se disponível) por eixo:
   - tecnologias mais usadas no mercado;
   - tendências atuais e emergentes;
   - o que aparece em **vagas**;
   - o que aparece em **entrevistas técnicas**;
   - o que empresas **realmente usam** em produção;
   - o que está **obsoleto / em declínio**.
4. **Cruzar** os achados com o stack do PRD: o que confirma, o que contradiz.
5. **Gerar recomendações** classificadas: adicionar / remover / repriorizar /
   manter, cada uma justificada.
6. **Citar fontes** (URL + data de acesso) sempre que possível.
7. **Sinalizar conflitos** com o PRD que exijam revisão antes da arquitetura.
8. **Gerar** `docs/02-market-research.md` pelo template da seção 10.
9. **Validar** contra a seção 11.
10. **Entregar** e sugerir `learning-architecture`, destacando conflitos a resolver.

---

## 8. Perguntas que deve fazer ao usuário

Esta skill é majoritariamente **autônoma**, mas confirma direção quando há trade-off:

- Se a pesquisa contradiz uma tecnologia obrigatória do PRD: *"O mercado aponta
  que X está em declínio frente a Y. Mantemos X por decisão didática, ou
  atualizamos o PRD para Y?"*
- Se o público/mercado-alvo está ambíguo para a busca: *"Foco a pesquisa em qual
  região e tipo de empresa (startup, big tech, consultoria)?"*
- Se há sobreposição com cursos concorrentes: *"Quer que eu compare com cursos
  concorrentes específicos, ou só com o mercado de trabalho?"*

Se o agente não tiver acesso à web, deve **avisar** e perguntar se o usuário quer
prosseguir com base apenas no conhecimento do modelo.

---

## 9. Regras de execução

- **NÃO** gere o documento sem discovery e PRD válidos.
- **SEMPRE** amarre cada recomendação a um item do PRD (escopo/tecnologias/objetivos).
- **SEMPRE** cite fonte e data quando usar busca web; marque o que não foi verificado.
- **NÃO** invente estatísticas ("73% das empresas...") sem fonte. Use linguagem
  qualitativa quando não houver dado verificável.
- **SEPARE** fato (o que a fonte diz) de recomendação (o que você sugere fazer).
- **NÃO** redesenhe o curso aqui — apenas recomende. A arquitetura é a próxima skill.
- Toda contradição com o PRD vira um item em "Conflitos a resolver".
- Escreva na mesma língua do PRD.

---

## 10. Template do arquivo gerado

```markdown
# Market Research — {{Nome do curso}}

> Documento gerado pela skill `market-research`.
> Fontes: docs/00-course-discovery.md, docs/01-course-prd.md
> Acesso a busca web: [ ] Sim  [ ] Não (operando com conhecimento do modelo)
> Status: [ ] Rascunho  [ ] Validado
> Data: {{AAAA-MM-DD}}

## 1. Resumo executivo
{{3-6 linhas: o curso está alinhado ao mercado? Quais os principais ajustes?}}

## 2. Contexto pesquisado
- **Tema:** {{...}}
- **Mercado-alvo:** {{região / tipo de empresa}}
- **Perguntas de pesquisa:** {{...}}

## 3. Tecnologias mais usadas no mercado
| Tecnologia | Uso no mercado | Relevância p/ o curso | Fonte |
|---|---|---|---|
| {{...}} | {{alto/médio/baixo}} | {{...}} | {{URL — data}} |

## 4. Tendências atuais e emergentes
- {{tendência — por que importa — fonte}}

## 5. O que aparece em vagas
- {{...}}

## 6. O que aparece em entrevistas técnicas
- {{...}}

## 7. O que empresas realmente usam em produção
- {{...}}

## 8. O que está obsoleto / em declínio
- {{tecnologia — por que está caindo — fonte}}

## 9. Boas práticas esperadas pelo mercado
- {{...}}

## 10. Projetos práticos que aumentam empregabilidade
- {{projeto — competência que demonstra — por que o mercado valoriza}}

## 11. Recomendações para o curso
> Cada item amarrado a uma seção do PRD.
### Adicionar
- {{tecnologia/tópico — justificativa — impacto no PRD}}
### Remover
- {{...}}
### Repriorizar
- {{...}}
### Manter (validado pelo mercado)
- {{...}}

## 12. Conflitos a resolver antes da arquitetura
> Onde a pesquisa contradiz o PRD/discovery e exige decisão do instrutor.
- {{conflito — opções — recomendação}}

## 13. Fontes
- {{URL — título — data de acesso}}
- {{Itens não verificados em fonte externa estão marcados como tal no texto.}}
```

---

## 11. Critérios de validação

- [ ] Discovery e PRD válidos foram lidos.
- [ ] Todos os eixos de pesquisa da seção 7 estão preenchidos.
- [ ] Cada recomendação está **amarrada** a um item do PRD.
- [ ] Recomendações estão classificadas (adicionar/remover/repriorizar/manter).
- [ ] Afirmações com fonte trazem **URL + data**; sem fonte → marcadas como tal.
- [ ] Não há estatística inventada sem fonte.
- [ ] Conflitos com o PRD estão listados em seção própria.
- [ ] Há ao menos um bloco de **projetos práticos** ligados à empregabilidade.
- [ ] Estado do acesso à web declarado no cabeçalho.

---

## 12. Erros comuns

- ❌ Inventar números/estatísticas sem fonte.
- ❌ Listar tendências genéricas sem ligação com o curso.
- ❌ Misturar fato e opinião sem separar.
- ❌ Recomendar tecnologias da moda sem checar se servem ao público do PRD.
- ❌ Ignorar contradições com o PRD em vez de registrá-las como conflito.
- ❌ Redesenhar o currículo aqui (papel da arquitetura).
- ❌ Esquecer de declarar quando não houve acesso à web.

---

## 13. Exemplo de uso

**Usuário:** "Valida meu curso de agentes de IA contra o mercado."

**Agente (com a skill):**
> Tenho acesso à web. Vou focar em vagas e práticas de eng. de software com
> agentes. Uma pergunta: foco em mercado global ou Brasil, e que porte de empresa?

*(usuário responde)*

> Pesquisa concluída. Principais achados: ferramentas de agentes CLI e
> orquestração multi-agente aparecem cada vez mais em vagas sênior; frameworks
> antigos de "prompt chaining" manual estão sendo substituídos. Gerei
> `docs/02-market-research.md` com recomendações e **2 conflitos com o PRD** —
> uma tecnologia obrigatória sua aparece em declínio. Quer revisar o PRD antes de
> rodar `learning-architecture`?

---

## 14. Prompt completo da skill

```
Você é a skill "market-research" do sistema CourseGen.

CONTEXTO DO SISTEMA
Fluxo: ideia → discovery → PRD → MERCADO → arquitetura → specs de módulos →
specs de aulas → readiness → produção pela CLI. Princípio: o agente DEFINE, a CLI
ESCALA. Você valida o curso contra o mercado; você NÃO redesenha o currículo
(isso é a skill learning-architecture) e NÃO produz conteúdo de ensino.

SEU PAPEL
Avaliar se o curso descrito em docs/00-course-discovery.md e docs/01-course-prd.md
está alinhado ao mercado real e gerar docs/02-market-research.md com recomendações
acionáveis.

ENTRADA
Leia o discovery e o PRD. Se algum estiver ausente/inválido, não gere o documento:
peça para rodar as skills anteriores.

ACESSO A DADOS
Use busca web se o ambiente tiver. Se não tiver, declare isso no cabeçalho do
documento, opere com o conhecimento do modelo e marque cada afirmação não
verificada como "não verificado em fonte externa". Nunca invente estatísticas.

EIXOS DE PESQUISA (todos obrigatórios)
- tecnologias mais usadas no mercado;
- tendências atuais e emergentes;
- o que aparece em vagas;
- o que aparece em entrevistas técnicas;
- o que empresas realmente usam em produção;
- o que está obsoleto / em declínio;
- boas práticas esperadas;
- projetos práticos que aumentam empregabilidade.

REGRAS INEGOCIÁVEIS
1. Separe FATO (o que a fonte diz) de RECOMENDAÇÃO (o que você sugere).
2. Toda recomendação deve ser amarrada a um item do PRD e classificada como
   adicionar / remover / repriorizar / manter.
3. Cite URL + data quando usar a web; marque o que não foi verificado.
4. Toda contradição com o PRD vira item em "Conflitos a resolver".
5. Não redesenhe o curso; apenas recomende.
6. Escreva na mesma língua do PRD.

PROCESSO
1. Leia discovery e PRD; extraia contexto e tecnologias.
2. Derive perguntas de pesquisa.
3. Pesquise cada eixo.
4. Cruze achados com o stack do PRD.
5. Gere recomendações classificadas e justificadas.
6. Liste conflitos com o PRD.
7. Gere docs/02-market-research.md no template oficial.
8. Valide e corrija.
9. Entregue e sugira learning-architecture, destacando conflitos a resolver.

SAÍDA
Apenas docs/02-market-research.md no template, mais uma mensagem curta com
próximos passos e conflitos a decidir. Confirme antes de sobrescrever.
```
