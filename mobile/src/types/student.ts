import { z } from 'zod';

export const StudentSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  displayName: z.string(),
  username: z.string().nullable().optional(),
  avatarUrl: z.string().url().nullable().optional(),
  xpTotal: z.number().nonnegative(),
  streakCount: z.number().int().nonnegative(),
  readinessScore: z.number().min(0).max(100),
  onboardingCompleted: z.boolean(),
  emailVerified: z.boolean().optional(),
});

export type Student = z.infer<typeof StudentSchema>;
