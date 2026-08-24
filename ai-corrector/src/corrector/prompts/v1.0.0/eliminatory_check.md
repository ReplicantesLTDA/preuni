CRITÉRIOS ELIMINATÓRIOS — VERIFICAÇÃO OBRIGATÓRIA ANTES DA PONTUAÇÃO POR COMPETÊNCIA

Antes de pontuar cada competência, avalie se algum dos critérios eliminatórios abaixo se aplica.
Se sim, registre-o em `eliminatory_flags` e zere as competências afetadas conforme a matriz
oficial. Use exclusivamente as seguintes strings:

1. `off_topic` — Fuga total ao tema
   - O texto não dialoga, em nenhum momento, com o tema proposto.
   - Sinais: aborda apenas o eixo temático genérico (ex.: "tecnologia") sem o recorte específico
     do tema; copia o motivo do tema mas desenvolve outro assunto.
   - Conduta: zerar todas as cinco competências.

2. `annulled` — Anulação
   - Cópia integral ou predominante de trechos da proposta motivadora.
   - Texto ofensivo, com palavrões deliberados, ou desconectado do exercício.
   - Texto em forma de cópia pura sem produção autoral.
   - Inclua sempre `annulment_reason_pt_br` com a justificativa específica.
   - Conduta: zerar todas as cinco competências.

3. `insufficient_text` — Texto insuficiente
   - Geralmente menos de 7 linhas manuscritas (ou volume textual claramente abaixo do mínimo).
   - O conteúdo é tão curto que impede análise de qualquer competência.
   - Conduta: zerar todas as cinco competências.

4. `not_dissertative_argumentative` — Gênero textual incorreto
   - O texto é narrativo (conto, crônica), poético (poema), epistolar (carta pessoal), ou outro
     gênero que não respeita o dissertativo-argumentativo exigido pela banca.
   - Conduta: zerar as competências C1, C2, C3 e C5 (C4 pode receber pontuação parcial em casos
     limítrofes, mas, por padrão, considere todas zeradas).

Quando NENHUM critério eliminatório se aplica, `eliminatory_flags` é uma lista vazia `[]` e a
correção segue normalmente pelas 5 competências.

EXEMPLOS DE DECISÃO

- Tema: "Desafios para a valorização de comunidades e povos tradicionais no Brasil" — Texto
  trata genericamente de "preservação ambiental", sem menção a comunidades tradicionais:
  `off_topic`.

- Tema sobre "manipulação do comportamento via Internet" — Texto reproduz textualmente parte da
  motivadora e adiciona duas frases genéricas: `annulled` com
  `annulment_reason_pt_br: "Cópia integral de excertos da motivadora sem produção autoral
  significativa."`.

- Texto com 5 linhas manuscritas equivalentes a ~200 caracteres: `insufficient_text`.

- Texto bem escrito em forma de carta para o ministro da Educação: `not_dissertative_argumentative`.
