import { z } from 'zod';
import type { ApiClient } from '@/lib/api/client';
import { StudentSchema, type Student } from '@/types/student';

const Empty = z.object({}).partial();
type Empty = z.infer<typeof Empty>;

const AvatarUploadUrlSchema = z.object({
  uploadUrl: z.string().url(),
  objectKey: z.string(),
});
export type AvatarUploadUrl = z.infer<typeof AvatarUploadUrlSchema>;

export interface UpdateStudentInput {
  displayName?: string;
  username?: string;
}

export interface ChangePasswordInput {
  currentPassword: string;
  newPassword: string;
}

export interface ChangeEmailRequestInput {
  newEmail: string;
}

export interface ChangeEmailConfirmInput {
  newEmail: string;
  otp: string;
}

export function makePerfilApi(api: ApiClient) {
  return {
    getMe(): Promise<Student> {
      return api.request({ method: 'GET', path: '/v1/students/me', schema: StudentSchema });
    },

    updateMe(input: UpdateStudentInput): Promise<Student> {
      return api.request({
        method: 'PATCH',
        path: '/v1/students/me',
        body: input,
        schema: StudentSchema,
      });
    },

    deleteMe(confirmation: string): Promise<Empty> {
      return api.request({
        method: 'DELETE',
        path: '/v1/students/me',
        body: { confirmation },
        schema: Empty,
      });
    },

    changePassword(input: ChangePasswordInput): Promise<Empty> {
      return api.request({
        method: 'POST',
        path: '/v1/auth/password/change',
        body: input,
        schema: Empty,
      });
    },

    changeEmailRequest(input: ChangeEmailRequestInput): Promise<Empty> {
      return api.request({
        method: 'POST',
        path: '/v1/auth/email/change/request',
        body: input,
        schema: Empty,
      });
    },

    changeEmailConfirm(input: ChangeEmailConfirmInput): Promise<Empty> {
      return api.request({
        method: 'POST',
        path: '/v1/auth/email/change/confirm',
        body: input,
        schema: Empty,
      });
    },

    getAvatarUploadUrl(): Promise<AvatarUploadUrl> {
      return api.request({
        method: 'PUT',
        path: '/v1/students/me/avatar',
        schema: AvatarUploadUrlSchema,
      });
    },

    confirmAvatar(objectKey: string): Promise<Student> {
      return api.request({
        method: 'POST',
        path: '/v1/students/me/avatar/confirm',
        body: { objectKey },
        schema: StudentSchema,
      });
    },

    requestDataExport(): Promise<Empty> {
      return api.request({
        method: 'GET',
        path: '/v1/students/me/data-export',
        schema: Empty,
      });
    },
  };
}

export type PerfilApi = ReturnType<typeof makePerfilApi>;
