CRITÉRIOS ELIMINATÓRIOS — APLICAR APENAS COM EVIDÊNCIA INEQUÍVOCA

# Regra geral

**Default**: `eliminatory_flags = []` (vazio). Aplique uma flag SOMENTE quando outro avaliador
humano também aplicaria sem hesitação. Não aplique por desconforto, por divergência com a
posição do aluno, ou porque o texto parece confuso. Em caso de dúvida: **NÃO aplique**.

# Critérios e SEUS NÃO-APLICAR

## 1. `off_topic` — Fuga total ao tema

**Aplique APENAS SE**: o texto inteiro NÃO menciona o tema específico nem nenhum conceito
diretamente relacionado.

**NÃO aplique se**: o texto cita o tema central pelo menos uma vez OU menciona conceitos
conectados (mesmo que de forma rasa). Tangenciamento parcial = Competência 2 baixa, não
`off_topic`.

Exemplo NÃO-aplicar: Tema sobre "decreto de armas no Brasil", texto discute "violência" e
"educação" como solução para violência — está tangencial mas conectado. Não é `off_topic`.

## 2. `annulled` — Anulação

**Aplique APENAS SE**: o texto é majoritariamente cópia literal da motivadora (>60% do
texto) OU contém ofensa grave deliberada (palavrão, ataque a grupo) OU é uma assinatura/em
branco. Inclua obrigatoriamente `annulment_reason_pt_br` apontando o trecho problemático.

**NÃO aplique se**: o texto cita ou parafraseia trechos da motivadora mas tem produção
autoral própria (a citação é repertório, não cópia). NÃO aplique por uso de palavras
incomuns.

## 3. `insufficient_text` — Texto insuficiente

**Aplique APENAS SE**: o texto tem menos de 7 linhas manuscritas (≈ menos de 500 caracteres
contínuos) E o conteúdo é tão raso que impede análise de qualquer competência.

**NÃO aplique se**: o texto tem 7+ linhas (ou ≈ 500+ caracteres), mesmo que rasamente
desenvolvido. Pré-validação do sistema JÁ rejeita textos abaixo do mínimo — se o texto chegou
até você, ele atende ao mínimo de tamanho. Texto curto mas com 7+ linhas vira nota baixa em
todas as competências (40 ou 80), NÃO `insufficient_text`.

## 4. `not_dissertative_argumentative` — Gênero textual incorreto

**Aplique APENAS SE**: o texto é INTEIRAMENTE narrativo (conto, crônica com personagens e
enredo), poético (versos), epistolar (carta com vocativo + despedida + assinatura), ou
relato de experiência pessoal sem tese.

**NÃO aplique se**: o texto tem estrutura dissertativo-argumentativa (tese + argumentos +
conclusão) mesmo que tenha falhas — argumentos vagos ou conclusão fraca são problemas de
competência (C2/C3/C5), não de gênero. NÃO aplique apenas porque a tese é implícita.

# QUANDO A LISTA FICA VAZIA

Se nenhum dos 4 critérios bate INEQUIVOCAMENTE, `eliminatory_flags = []`. Avalie as 5
competências normalmente, mesmo que a redação seja fraca. Uma redação fraca pode receber
40 ou 80 em cada competência (total 200-400). Isso é menos punitivo e mais útil ao aluno do
que zerar.

# AUDITORIA

Antes de definir `eliminatory_flags`, pergunte-se:
- "Tenho citação literal do texto do aluno que comprova essa flag?"
- "Um avaliador humano hesitaria? Se sim, NÃO aplico."

Se não responder SIM à primeira e NÃO à segunda, mantenha `eliminatory_flags = []`.
