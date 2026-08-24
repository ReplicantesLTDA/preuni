AVALIE AS 5 COMPETÊNCIAS

Para CADA competência:
1. Identifique UMA evidência específica no texto (frase, expressão, ou ausência observável).
2. Mapeie essa evidência ao nível mais próximo nos descritores abaixo.
3. Se a evidência está entre dois níveis, escolha o MAIOR (a Matriz pede um nível, não um meio
   ponto). Em caso de empate genuíno: escolha o nível superior se tiver mais indícios; o
   inferior se tiver dúvida real.
4. Cite a evidência LITERAL no campo `excerpt`. Comente especificamente no
   `justification_pt_br` (NÃO repita o descritor — cite a evidência).
5. Se score < 200, o `improvement_path_pt_br` deve ser ACIONÁVEL (algo que o aluno consegue
   praticar), não genérico.

# REGRAS DE CALIBRAÇÃO PRÁTICAS

- Texto com estrutura clara (intro + 2 desenvolvimentos + conclusão), repertório citado,
  proposta com 3+ elementos → todas as competências em **120 ou mais**.
- Texto com tese reconhecível mas argumentação fraca → competências entre **80 e 120**.
- Apenas redações realmente confusas, sem tese identificável OU com desvios gravíssimos
  em quase todos os períodos → **40 ou abaixo**.
- **Nota 0 em alguma competência só com evidência forte** (apoio textual flagrante).

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
- `excerpt` = substring LITERAL da redação enviada (mantenha acentuação e pontuação).
- `improvement_path_pt_br` obrigatório quando `score < 200`; AUSENTE quando `score == 200`.
- Lista de eliminatórios vazia por padrão; aplique apenas conforme `eliminatory_check`.
- Retorne SOMENTE o objeto JSON especificado. Sem texto fora do JSON.
