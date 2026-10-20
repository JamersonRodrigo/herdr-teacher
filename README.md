# herdr-teacher 🎓

> Framework universal de mentoria e ensino multiagente para o [Herdr](https://github.com/jamersonrodrigo111).
> Aprenda qualquer matéria, tecnologia, teoria ou projeto partindo do **absoluto zero**, com didática leve, exercícios práticos, recursos visuais e revisão adversária anti-viés.

---

## 🌟 O Que É o `herdr-teacher`?

Enquanto o `herdr-agents` (ou `herdr-soho`) é um time de agentes voltado a programar e implementar features em software, o **`herdr-teacher`** é um time de agentes pedagógicos desenhado para **ensinar qualquer disciplina ou projeto para você**:
- Economia & Finanças
- Desenvolvimento de Software & TI
- Relações Internacionais & Geopolítica
- História, Filosofia e Ciências
- Projetos e Arquiteturas Complexas

A skill assume por padrão que o aluno começa do **absoluto zero** e constrói o conhecimento de forma progressiva e sólida.

---

## 👥 A Equipe de Agentes

| Papel | Agente | Responsabilidade |
|---|---|---|
| **Professor (Orquestrador)** | Claude / Agy | Conduz a mentoria com didática leve e descontraída, mostra relações de causa e efeito ("como X afeta Y"), avalia a maturidade e fecha lacunas conceituais. |
| **Pesquisador** | Agy (Gemini) | Pesquisa fontes canônicas, documentações oficiais e realiza fact-checking em tempo real. |
| **Criador de Conteúdo** | Agy (Gemini) | **Papel exclusivo para criação de arquivos:** gera pastas por tópicos, tabelas personalizadas, diagramas Mermaid, Excalidraw, slides e exercícios. Não audita o próprio trabalho. |
| **Revisor Adversário** | Claude (Sonnet) | **Isolamento Total:** NUNCA roda no mesmo agente que o criador ou pesquisador. Caça falhas, erros técnicos, código quebrado e vieses. Se não achar falhas, deve formalmente admitir que falhou em encontrar erros e aprovar. |
| **Advogado do Diabo** | Claude | Apresenta contrapontos, escolas concorrentes e caça falácias no raciocínio. |
| **Avaliador Pedagógico** | Claude | Cria listas de exercícios progressivos, desafios práticos e aplica o Método Feynman. |
| **Analista de Cenários** | Agy | Modela estudos de caso práticos, simulações e matrizes de trade-offs do mundo real. |

---

## 📂 Organização dos Materiais

Todos os materiais gerados durante as aulas são salvos organizados por tópicos em pastas na raiz do repositório onde a sessão estiver ativa:

```
./aulas/ (ou ./lessons/)
  ├── 01-introducao-e-fundamentos/
  │   ├── README.md                 # Explicação didática da aula
  │   ├── tabela-comparativa.md     # Matriz de conceitos e trade-offs
  │   ├── fluxo-conceito.mermaid    # Diagrama de fluxo e relações de causa/efeito
  │   ├── arquitetura.excalidraw    # Diagrama visual conceitual
  │   ├── exercicios.md             # Desafios práticos e atividades
  │   └── apresentacao.marp.md      # Slides da aula
  └── 02-aprofundamento-e-pratica/
```

---

## 🚀 Como Usar

No terminal com Herdr ativo (`HERDR_ENV=1`), execute:

```bash
/herdr-teacher "Quero aprender Rust do zero"
# ou
/herdr-teacher "Quero aulas de Economia e Geopolítica"
# ou
/herdr-teacher "Quero entender a arquitetura deste repositório"
```

---

## ⚖️ Licença

MIT License.
