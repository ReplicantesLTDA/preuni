import { z } from 'zod';

export const SubmitEssaySchema = z.object({
  promptThemeTitle: z.string().min(1, 'Informe o título do tema.'),
  promptThemeContext: z.string().min(1, 'Informe o contexto do tema.'),
  essayText: z.string().min(1, 'Escreva sua redação antes de enviar.'),
});
export type SubmitEssayInput = z.infer<typeof SubmitEssaySchema>;
