import { z } from 'zod';

export const FriendSchema = z.object({
  friendshipId: z.string().uuid(),
  userId: z.string().uuid(),
  displayName: z.string(),
  currentStreak: z.number().int().nonnegative(),
  latestOverallScore: z.number().int().nullable().optional(),
  latestGradedAt: z.string().nullable().optional(),
});
export type Friend = z.infer<typeof FriendSchema>;

export const FriendRequestSchema = z.object({
  id: z.string().uuid(),
  status: z.string(),
});
export type FriendRequest = z.infer<typeof FriendRequestSchema>;
