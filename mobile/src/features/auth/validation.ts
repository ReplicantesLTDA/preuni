import { z } from 'zod';

export const LoginFormSchema = z.object({
  email: z.string().email('Informe um e-mail válido.'),
  password: z.string().min(1, 'Informe sua senha.'),
});

export const RegisterFormSchema = z.object({
  displayName: z
    .string()
    .min(2, 'Mínimo de 2 caracteres.')
    .max(64, 'Máximo de 64 caracteres.'),
  email: z.string().email('Informe um e-mail válido.'),
  password: z
    .string()
    .min(8, 'A senha precisa ter pelo menos 8 caracteres.')
    .max(128, 'A senha é grande demais.')
    .regex(/[A-Za-z]/, 'A senha precisa de pelo menos uma letra.')
    .regex(/\d/, 'A senha precisa de pelo menos um número.'),
});

export const OtpFormSchema = z.object({
  otp: z.string().regex(/^\d{6}$/, 'O código tem 6 dígitos.'),
});

export const ResetRequestFormSchema = z.object({
  email: z.string().email('Informe um e-mail válido.'),
});

export const ResetConfirmFormSchema = z.object({
  email: z.string().email(),
  otp: z.string().regex(/^\d{6}$/),
  newPassword: RegisterFormSchema.shape.password,
});

export type LoginForm = z.infer<typeof LoginFormSchema>;
export type RegisterForm = z.infer<typeof RegisterFormSchema>;
