import { z } from 'zod';
import type { ApiClient } from '@/lib/api/client';
import { SessionSchema, type Session } from '@/types/auth';
import { StudentSchema, type Student } from '@/types/student';

const Empty = z.object({}).partial();
type Empty = z.infer<typeof Empty>;

const RegisterResponseSchema = SessionSchema.extend({ student: StudentSchema.optional() });
const LoginResponseSchema = SessionSchema.extend({ student: StudentSchema.optional() });

export function makeAuthApi(api: ApiClient) {
  return {
    register(input: { email: string; password: string; displayName: string }): Promise<Session> {
      return api.request({
        method: 'POST',
        path: '/v1/auth/register',
        body: input,
        schema: RegisterResponseSchema,
        auth: false,
      });
    },

    login(input: { email: string; password: string }): Promise<Session> {
      return api.request({
        method: 'POST',
        path: '/v1/auth/login',
        body: input,
        schema: LoginResponseSchema,
        auth: false,
      });
    },

    logout(): Promise<Empty> {
      return api.request({
        method: 'POST',
        path: '/v1/auth/logout',
        body: {},
        schema: Empty,
      });
    },

    verifyEmail(input: { email: string; otp: string }): Promise<Empty> {
      return api.request({
        method: 'POST',
        path: '/v1/auth/email/verify',
        body: input,
        schema: Empty,
        auth: false,
      });
    },

    resendVerificationOtp(input: { email: string }): Promise<Empty> {
      return api.request({
        method: 'POST',
        path: '/v1/auth/email/verify-resend',
        body: input,
        schema: Empty,
        auth: false,
      });
    },

    requestOtpLogin(input: { email: string }): Promise<Empty> {
      return api.request({
        method: 'POST',
        path: '/v1/auth/otp/request',
        body: input,
        schema: Empty,
        auth: false,
      });
    },

    verifyOtpLogin(input: { email: string; otp: string }): Promise<Session> {
      return api.request({
        method: 'POST',
        path: '/v1/auth/otp/verify',
        body: input,
        schema: LoginResponseSchema,
        auth: false,
      });
    },

    requestPasswordReset(input: { email: string }): Promise<Empty> {
      return api.request({
        method: 'POST',
        path: '/v1/auth/password/reset/request',
        body: input,
        schema: Empty,
        auth: false,
      });
    },

    confirmPasswordReset(input: {
      email: string;
      otp: string;
      newPassword: string;
    }): Promise<Empty> {
      return api.request({
        method: 'POST',
        path: '/v1/auth/password/reset/confirm',
        body: input,
        schema: Empty,
        auth: false,
      });
    },

    getMe(): Promise<Student> {
      return api.request({
        method: 'GET',
        path: '/v1/students/me',
        schema: StudentSchema,
      });
    },
  };
}

export type AuthApi = ReturnType<typeof makeAuthApi>;
