import { z } from 'zod';
import type { ApiClient } from '@/lib/api/client';
import { FriendSchema, FriendRequestSchema, type Friend, type FriendRequest } from '@/types/social';

const FriendListSchema = z.array(FriendSchema);
const Empty = z.object({}).partial();
type Empty = z.infer<typeof Empty>;

export function makeSocialApi(api: ApiClient) {
  return {
    listFriends(): Promise<Friend[]> {
      return api.request({ method: 'GET', path: '/v1/friends', schema: FriendListSchema });
    },

    sendRequest(addresseeId: string): Promise<FriendRequest> {
      return api.request({
        method: 'POST',
        path: '/v1/friends/requests',
        body: { addresseeId },
        schema: FriendRequestSchema,
      });
    },

    acceptRequest(id: string): Promise<FriendRequest> {
      return api.request({
        method: 'POST',
        path: `/v1/friends/requests/${id}/accept`,
        schema: FriendRequestSchema,
      });
    },

    removeFriend(id: string): Promise<Empty> {
      return api.request({
        method: 'DELETE',
        path: `/v1/friends/${id}`,
        schema: Empty,
      });
    },
  };
}

export type SocialApi = ReturnType<typeof makeSocialApi>;
