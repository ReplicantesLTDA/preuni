import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useApi } from '@/lib/api/context';
import { useSessionStore } from '@/stores/sessionStore';
import { qk } from '@/lib/query/keys';
import { makePerfilApi } from './api';

export function usePerfilApi() {
  const api = useApi();
  return makePerfilApi(api);
}

export function useUpdateMe() {
  const api = usePerfilApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.updateMe,
    onSuccess: (student) => {
      useSessionStore.getState().setAuthed(student);
      void qc.invalidateQueries({ queryKey: qk.studentMe() });
    },
  });
}

export function useDeleteMe() {
  const api = usePerfilApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.deleteMe,
    onSuccess: () => {
      useSessionStore.getState().setAnon();
      qc.clear();
    },
  });
}

export function useChangePassword() {
  const api = usePerfilApi();
  return useMutation({ mutationFn: api.changePassword });
}

export function useChangeEmailRequest() {
  const api = usePerfilApi();
  return useMutation({ mutationFn: api.changeEmailRequest });
}

export function useChangeEmailConfirm() {
  const api = usePerfilApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.changeEmailConfirm,
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.studentMe() }),
  });
}

export function useRequestDataExport() {
  const api = usePerfilApi();
  return useMutation({ mutationFn: api.requestDataExport });
}

export function useUploadAvatar() {
  const api = usePerfilApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: { uri: string; mimeType?: string }) => {
      const { uploadUrl, objectKey } = await api.getAvatarUploadUrl();
      // Dev backend returns a stub presigned URL (`X-Amz-Credential=PRESIGNED`).
      // Real S3 PUT would fail because no signature is wired in v1; skip the
      // upload step and go straight to confirm. In prod the URL is a real
      // presigned URL and we PUT the bytes first.
      const isDevStub = /PRESIGNED|placeholder/i.test(uploadUrl);
      if (!isDevStub) {
        const resp = await fetch(input.uri);
        const blob = await resp.blob();
        const putResp = await fetch(uploadUrl, {
          method: 'PUT',
          headers: { 'Content-Type': input.mimeType ?? 'image/webp' },
          body: blob,
        });
        if (!putResp.ok) {
          throw {
            kind: 'Unknown' as const,
            message: 'Falha ao enviar a imagem. Tente novamente.',
          };
        }
      }
      return api.confirmAvatar(objectKey);
    },
    onSuccess: (student) => {
      useSessionStore.getState().setAuthed(student);
      void qc.invalidateQueries({ queryKey: qk.studentMe() });
    },
  });
}
