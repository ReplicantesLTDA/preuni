AVALIE AS 5 COMPETÊNCIAS

Para CADA competência:
1. Olhe a tabela do fragmento da competência abaixo.
2. Identifique a UMA evidência específica no texto que define o nível.
3. Mapeie ao nível mais próximo. Em empate genuíno: escolha o **MAIOR**.
4. Cite a evidência LITERAL no campo `excerpt`.
5. Justifique com EVIDÊNCIA do texto (não com paráfrase do descritor).
6. Se `score < 200`, `improvement_path_pt_br` deve ser ACIONÁVEL.

# REGRA ANTI-MEDIANIZAÇÃO

- Se a competência mostra **excelência clara** (todos os sinais do nível 200), dê **200**.
- Se mostra **fragilidade clara** (sinais do nível 40 ou 80), dê **40 ou 80**.
- Os níveis 120 e 160 são para gradações reais, não para "evitar decisão".

# FRAGMENTOS POR COMPETÊNCIA

## Competência 1 — Domínio da modalidade escrita formal

{c1_fragment}

## Competência 2 — Compreensão da proposta + tipo dissertativo-argumentativo

{c2_fragment}

## Competência 3 — Seleção, organização e interpretação dos argumentos

{c3_fragment}

## Competência 4 — Mecanismos linguísticos para a argumentação

{c4_fragment}

## Competência 5 — Proposta de intervenção respeitando os direitos humanos

{c5_fragment}

# LEMBRETES FINAIS

- `score ∈ {0, 40, 80, 120, 160, 200}`. Nada intermediário.
- `final_score` = soma das 5; intervalo `[0, 1000]`. Calcule antes de responder.
- `excerpt` = substring LITERAL da redação (mantenha acentuação e pontuação).
- `improvement_path_pt_br` obrigatório quando `score < 200`; AUSENTE quando `score == 200`.
- `eliminatory_flags` vazia por padrão.
- **Retorne SOMENTE o objeto JSON. Sem markdown. Sem ```json. Sem comentários. Sem texto
  explicativo antes ou depois.** Apenas `{` no início e `}` no fim.
