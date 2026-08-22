import { useState } from 'react';
import { ScrollView, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { FormField } from '@/components/FormField';
import { OtpInput } from '@/components/OtpInput';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import { usePasswordResetConfirm, usePasswordResetRequest } from '@/features/auth/hooks';
import { ResetConfirmFormSchema, ResetRequestFormSchema } from '@/features/auth/validation';
import { authErrorMessage } from '@/features/auth/errorMessages';

type Step = 'request' | 'confirm';

export default function PasswordResetScreen() {
  const router = useRouter();
  const { color, space, font, size } = useTheme();
  const toast = useToast();
  const request = usePasswordResetRequest();
  const confirm = usePasswordResetConfirm();
  const [step, setStep] = useState<Step>('request');
  const [email, setEmail] = useState('');
  const [otp, setOtp] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [error, setError] = useState<string | null>(null);

  function submitRequest() {
    const parsed = ResetRequestFormSchema.safeParse({ email });
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? 'E-mail inválido');
      return;
    }
    setError(null);
    request.mutate(parsed.data, {
      onSuccess: () => {
        setStep('confirm');
        toast.show('Enviamos um código para seu e-mail.', 'info');
      },
      onError: (err) => toast.show(authErrorMessage(err), 'error'),
    });
  }

  function submitConfirm() {
    const parsed = ResetConfirmFormSchema.safeParse({ email, otp, newPassword });
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? 'Dados inválidos');
      return;
    }
    setError(null);
    confirm.mutate(parsed.data, {
      onSuccess: () => {
        toast.show('Senha atualizada. Faça login.', 'success');
        router.replace('/(auth)/login');
      },
      onError: (err) => setError(authErrorMessage(err)),
    });
  }

  return (
    <ScreenContainer>
      <ScrollView
        contentContainerStyle={{ padding: space[5], flexGrow: 1, justifyContent: 'center' }}
        keyboardShouldPersistTaps="handled"
      >
      <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['2xl'], marginBottom: space[4] }}>
        Redefinir senha
      </Text>

      {step === 'request' ? (
        <>
          <FormField
            label="E-mail"
            value={email}
            onChangeText={setEmail}
            keyboardType="email-address"
            autoCapitalize="none"
            autoComplete="email"
            error={error ?? undefined}
          />
          <Button label="Enviar código" onPress={submitRequest} loading={request.isPending} fullWidth />
        </>
      ) : (
        <>
          <Text style={{ color: color.ink[2], fontFamily: font.body, fontSize: size.md, marginBottom: space[3] }}>
            Código enviado para {email}.
          </Text>
          <OtpInput value={otp} onChange={setOtp} autoFocus />
          <View style={{ height: space[3] }} />
          <FormField
            label="Nova senha"
            value={newPassword}
            onChangeText={setNewPassword}
            secureTextEntry
            autoComplete="new-password"
            error={error ?? undefined}
            helper="Mín. 8 caracteres, 1 letra e 1 número."
          />
          <Button label="Atualizar senha" onPress={submitConfirm} loading={confirm.isPending} fullWidth />
        </>
      )}
      </ScrollView>
    </ScreenContainer>
  );
}
