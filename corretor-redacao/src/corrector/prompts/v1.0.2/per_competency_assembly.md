AVALIE AS 5 COMPETÊNCIAS

Para CADA competência:
1. Identifique UMA evidência específica no texto.
2. Mapeie a evidência ao nível mais próximo na tabela.
3. Em empate genuíno: **maior** se houver mais indícios; **menor** se houver dúvida clara.
4. Cite a evidência LITERAL no campo `excerpt`.
5. Justifique com EVIDÊNCIA do texto, não com paráfrase do descritor.
6. Se `score < 200`, `improvement_path_pt_br` deve ser ACIONÁVEL.

# REGRAS DE CALIBRAÇÃO PRÁTICAS

- Texto com tese clara, estrutura completa (intro + 2 desenvolvimentos + conclusão),
  repertório citado E proposta com 4+ elementos → competências em **160 ou mais**.
- Texto com tese reconhecível, estrutura presente, repertório básico, proposta com 2-3
  elementos → competências entre **80 e 120**.
- Texto sem tese clara OU sem proposta OU com cópia parcial de motivadores OU com desvios
  sistemáticos → competências entre **40 e 80**.
- Apenas redações REALMENTE confusas (sem tese identificável, sem repertório, sem proposta,
  desvios graves) → **40 ou abaixo**.

**Não tenha medo de 80 ou de 40 quando a evidência manda esse nível.** Subir notas por
generosidade enviesa todo o sistema.

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
