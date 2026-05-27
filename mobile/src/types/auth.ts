import { z } from 'zod';
import { StudentSchema } from './student';

export const SessionSchema = z.object({
  accessToken: z.string(),
  refreshToken: z.string(),
  accessTokenExpiresAt: z.string().datetime().optional(),
  expiresIn: z.number().int().optional(),
  studentId: z.string().uuid().optional(),
  student: StudentSchema.optional(),
});

export type Session = z.infer<typeof SessionSchema>;

export const RegisterRequestSchema = z.object({
  email: z.string().email(),
  password: z
    .string()
    .min(8, 'A senha precisa ter pelo menos 8 caracteres.')
    .max(128, 'A senha é grande demais.')
    .regex(/[A-Za-z]/, 'A senha precisa de pelo menos uma letra.')
    .regex(/\d/, 'A senha precisa de pelo menos um número.'),
  displayName: z.string().min(2).max(64),
});

export const LoginRequestSchema = z.object({
  email: z.string().email(),
  password: z.string().min(1),
});

export const VerifyEmailRequestSchema = z.object({
  email: z.string().email(),
  otp: z.string().regex(/^\d{6}$/, 'O código tem 6 dígitos.'),
});

export const ResendVerifyRequestSchema = z.object({
  email: z.string().email(),
});

export const OtpLoginRequestSchema = z.object({ email: z.string().email() });
export const OtpLoginVerifySchema = VerifyEmailRequestSchema;

export const PasswordResetRequestSchema = z.object({ email: z.string().email() });
export const PasswordResetConfirmSchema = z.object({
  email: z.string().email(),
  otp: z.string().regex(/^\d{6}$/),
  newPassword: z
    .string()
    .min(8)
    .max(128)
    .regex(/[A-Za-z]/)
    .regex(/\d/),
});

export type RegisterRequest = z.infer<typeof RegisterRequestSchema>;
export type LoginRequest = z.infer<typeof LoginRequestSchema>;
export type VerifyEmailRequest = z.infer<typeof VerifyEmailRequestSchema>;
