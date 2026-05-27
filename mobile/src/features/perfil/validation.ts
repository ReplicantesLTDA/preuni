import { z } from 'zod';

export const EditProfileSchema = z.object({
  displayName: z.string().min(2, 'Mínimo 2 caracteres.').max(64, 'Máximo 64 caracteres.'),
});

export const ChangePasswordSchema = z
  .object({
    currentPassword: z.string().min(1, 'Informe sua senha atual.'),
    newPassword: z
      .string()
      .min(8, 'A senha precisa ter pelo menos 8 caracteres.')
      .max(128, 'A senha é grande demais.')
      .regex(/[A-Za-z]/, 'A senha precisa de pelo menos uma letra.')
      .regex(/\d/, 'A senha precisa de pelo menos um número.'),
    confirmPassword: z.string(),
  })
  .refine((d) => d.newPassword === d.confirmPassword, {
    path: ['confirmPassword'],
    message: 'As senhas não coincidem.',
  });

export const ChangeEmailRequestSchema = z.object({
  newEmail: z.string().email('Informe um e-mail válido.'),
});

export const ChangeEmailConfirmSchema = z.object({
  newEmail: z.string().email(),
  otp: z.string().regex(/^\d{6}$/, 'O código tem 6 dígitos.'),
});
