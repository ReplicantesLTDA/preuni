INSTRUÇÕES POR COMPETÊNCIA

Avalie cada competência seguindo rigorosamente a Matriz de Referência. Para CADA competência,
escolha o nível de pontuação cuja descrição melhor caracteriza a redação. Em caso de dúvida
entre dois níveis adjacentes, opte pelo inferior, justificando o motivo no campo
`justification_pt_br`. Sempre cite um trecho LITERAL da redação no campo `excerpt`.

Os fragmentos abaixo descrevem os níveis 0–200 para cada competência.

---

# Competência 1 — Domínio da modalidade escrita formal da língua portuguesa

{c1_fragment}

---

# Competência 2 — Compreensão da proposta e aplicação de conceitos das várias áreas de conhecimento para desenvolver o tema, dentro dos limites estruturais do texto dissertativo-argumentativo

{c2_fragment}

---

# Competência 3 — Capacidade de selecionar, relacionar, organizar e interpretar informações, fatos, opiniões e argumentos em defesa de um ponto de vista

{c3_fragment}

---

# Competência 4 — Demonstrar conhecimento dos mecanismos linguísticos necessários para a construção da argumentação

{c4_fragment}

---

# Competência 5 — Elaborar proposta de intervenção para o problema abordado, respeitando os direitos humanos

{c5_fragment}

---

LEMBRETES FINAIS

- Cada `score` deve ser exatamente 0, 40, 80, 120, 160 ou 200.
- `final_score` = soma dos cinco scores; deve estar em [0, 1000].
- `excerpt` deve ser uma substring literal da redação enviada (preserve acentuação e pontuação).
- `improvement_path_pt_br` obrigatório quando `score < 200`; ausente quando `score == 200`.
- Em presença de qualquer flag eliminatória, as competências afetadas recebem 0 e a soma reflete
  esses zeros.
- Retorne SOMENTE o objeto JSON especificado. Nenhum texto fora do JSON.
