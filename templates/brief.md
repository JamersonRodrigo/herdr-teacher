# Brief — [tópico] · [papel] · [etapa]

Papel: [researcher|content-creator|reviewer|debater|evaluator|scenario-analyst] · Idioma do relatório: pt-BR
Type: [docs|review]

<!-- O motor herdr-soho faz lint deste brief no dispatch. Seções obrigatórias:
Objetivo (Goal), Resultado esperado (Expected result) ou Aceite, Proibido
(Forbidden), Relatório (Report), uma linha de "sem commit" e — para o
content-creator, o único papel que edita — Arquivos próprios (Owned files).
Por isso essas seções ficam com o cabeçalho em inglês, que o lint reconhece
sem configuração; o conteúdo e as outras seções são em português. Releia o brief como o worker: todo ponto em que ele
teria de escolher é uma lacuna sua. Dois ou três itens por brief. -->

## Goal

[Um parágrafo: o que precisa ser verdade quando o worker terminar.]

## Contexto do aluno

- Matéria e objetivo do aluno: [...]
- Nível atual neste tópico (régua 0–5): [...]
- O que ele já viu: [aulas anteriores, com caminho]
- Dúvidas/erros recentes: [trechos da conversa, literais]

## Decisões já tomadas

[Tudo que o worker não deve escolher: nome da pasta (`aulas/NN-slug/`),
arquivos a gerar, profundidade, exemplos obrigatórios, posições que precisam
aparecer, tom.]

## Expected result

[O que existe e é observável no fim: arquivos, seções, tabelas, veredito.]

## Acceptance criteria

1. [critério] — provado por [`comando`] → [saída esperada] ou [conferência no arquivo]
2. [...]

## Owned files

<!-- só para o content-creator; papéis read-only podem omitir -->
- `aulas/NN-slug/README.md` — aula
- `aulas/NN-slug/exercicios.md` — exercícios
- [...]

## Fontes (ler antes)

- Dossiê do pesquisador: `[caminho do relatório]`
- Relatórios de apoio: `[debater/scenario-analyst/evaluator]`
- Templates: `[caminho da skill]/templates/[aula.md|exercicios.md|slides.marp.md|progresso.md]`
- Achados da revisão anterior (em correção): `[caminho do relatório do reviewer]`

## Quando o brief não decidir

Não escolha. Marque o item `[partial]`, liste a lacuna e as opções em
`pendencias` e siga com os outros itens. Nunca invente nomes, números, datas,
citações, URLs ou requisitos.

## Forbidden

- Qualquer arquivo fora de `Owned files` (papéis read-only: qualquer arquivo).
- Declarar o material aprovado (exceto o reviewer, no próprio relatório).
- No commit, push, tag, or PR. Sem commit: o professor é dono do git.

## Report

Escreva em Markdown no caminho do relatório que o dispatch informar, no
formato da seção `<report>` do seu papel, com `[done]` / `[partial]` /
`[skipped]` + motivo por item. Ao terminar, responda só com o caminho.
