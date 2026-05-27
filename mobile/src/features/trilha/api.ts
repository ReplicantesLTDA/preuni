import type { ApiClient } from '@/lib/api/client';
import { StudentSchema, type Student } from '@/types/student';

// Trilha-specific endpoints will land under /v1/students/me/trilha when the
// content/learning domain ships. For now the home composes itself from
// GET /v1/students/me alone (streak / XP / readiness / onboarding_completed).

export function makeTrilhaApi(api: ApiClient) {
  return {
    getMe(): Promise<Student> {
      return api.request({
        method: 'GET',
        path: '/v1/students/me',
        schema: StudentSchema,
      });
    },
  };
}

export type TrilhaApi = ReturnType<typeof makeTrilhaApi>;
