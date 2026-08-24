import { z } from 'zod';
import type { ApiClient } from '@/lib/api/client';
import { EssaySubmissionSchema, EssaySummarySchema, type EssaySubmission, type EssaySummary } from '@/types/essay';
import type { SubmitEssayInput } from './validation';

const EssaySummaryListSchema = z.array(EssaySummarySchema);

export function makeEssayApi(api: ApiClient) {
  return {
    submit(input: SubmitEssayInput): Promise<EssaySubmission> {
      return api.request({
        method: 'POST',
        path: '/v1/essays',
        body: input,
        schema: EssaySubmissionSchema,
      });
    },

    get(id: string): Promise<EssaySubmission> {
      return api.request({
        method: 'GET',
        path: `/v1/essays/${id}`,
        schema: EssaySubmissionSchema,
      });
    },

    list(): Promise<EssaySummary[]> {
      return api.request({
        method: 'GET',
        path: '/v1/essays',
        schema: EssaySummaryListSchema,
      });
    },
  };
}

export type EssayApi = ReturnType<typeof makeEssayApi>;
