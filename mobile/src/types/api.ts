import { z } from 'zod';

export const ApiErrorBodySchema = z.object({
  code: z.string().optional(),
  field: z.string().optional(),
  message: z.string(),
});

export const ApiErrorEnvelopeSchema = z.object({
  error: ApiErrorBodySchema,
});

export type ApiErrorBody = z.infer<typeof ApiErrorBodySchema>;
export type ApiErrorEnvelope = z.infer<typeof ApiErrorEnvelopeSchema>;

export function parseApiErrorEnvelope(text: string): ApiErrorEnvelope | null {
  try {
    const parsed = JSON.parse(text) as unknown;
    const result = ApiErrorEnvelopeSchema.safeParse(parsed);
    return result.success ? result.data : null;
  } catch {
    return null;
  }
}
