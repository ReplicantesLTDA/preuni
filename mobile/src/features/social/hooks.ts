import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useApi } from '@/lib/api/context';
import { qk } from '@/lib/query/keys';
import { makeSocialApi } from './api';

export function useSocialApi() {
  const api = useApi();
  return makeSocialApi(api);
}

export function useFriends() {
  const api = useSocialApi();
  return useQuery({
    queryKey: qk.friendsList(),
    queryFn: api.listFriends,
  });
}

export function useSendFriendRequest() {
  const api = useSocialApi();
  return useMutation({ mutationFn: api.sendRequest });
}

export function useAcceptFriendRequest() {
  const api = useSocialApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.acceptRequest,
    onSuccess: () => void qc.invalidateQueries({ queryKey: qk.friendsList() }),
  });
}

export function useRemoveFriend() {
  const api = useSocialApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.removeFriend,
    onSuccess: () => void qc.invalidateQueries({ queryKey: qk.friendsList() }),
  });
}
