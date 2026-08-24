import { z } from 'zod';

export const CompetencySchema = z.object({
  competency: z.number().int().min(1).max(5),
  score: z.number().int().min(0).max(200),
  justificationPtBr: z.string(),
  excerpt: z.string(),
});
export type Competency = z.infer<typeof CompetencySchema>;

export const EssayGradeSchema = z.object({
  overallScore: z.number().int().min(0).max(1000),
  competencies: z.array(CompetencySchema),
  gradedAt: z.string(),
});
export type EssayGrade = z.infer<typeof EssayGradeSchema>;

export const EssaySubmissionStatusSchema = z.enum(['pending', 'graded', 'failed']);
export type EssaySubmissionStatus = z.infer<typeof EssaySubmissionStatusSchema>;

export const EssaySubmissionSchema = z.object({
  id: z.string().uuid(),
  status: EssaySubmissionStatusSchema,
  submittedAt: z.string(),
  grade: EssayGradeSchema.optional(),
});
export type EssaySubmission = z.infer<typeof EssaySubmissionSchema>;

export const EssaySummarySchema = z.object({
  id: z.string().uuid(),
  status: EssaySubmissionStatusSchema,
  submittedAt: z.string(),
});
export type EssaySummary = z.infer<typeof EssaySummarySchema>;

export const StreakSchema = z.object({
  currentStreak: z.number().int().nonnegative(),
  longestStreak: z.number().int().nonnegative(),
  lastActiveDay: z.string().nullable(),
});
export type Streak = z.infer<typeof StreakSchema>;
