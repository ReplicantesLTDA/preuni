import { z } from 'zod';
import type { ApiClient } from '@/lib/api/client';
import { StudentSchema } from '@/types/student';

export const CompleteOnboardingRequestSchema = z.object({
  interests: z.array(z.string()).default([]),
  displayName: z.string().min(2).max(64).optional(),
});

export type CompleteOnboardingRequest = z.infer<typeof CompleteOnboardingRequestSchema>;

export function makeOnboardingApi(api: ApiClient) {
  return {
    complete(input: CompleteOnboardingRequest) {
      return api.request({
        method: 'PATCH',
        path: '/v1/students/me/onboarding',
        body: input,
        schema: StudentSchema,
      });
    },
  };
}
