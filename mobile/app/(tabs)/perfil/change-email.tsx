import { useState } from 'react';
import { ScrollView, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { FormField } from '@/components/FormField';
import { OtpInput } from '@/components/OtpInput';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import {
  useChangeEmailConfirm,
  useChangeEmailRequest,
} from '@/features/perfil/hooks';
import { ChangeEmailConfirmSchema, ChangeEmailRequestSchema } from '@/features/perfil/validation';
import { authErrorMessage } from '@/features/auth/errorMessages';
import { t } from '@/lib/i18n/pt-BR';

type Step = 'request' | 'confirm';

export default function ChangeEmailScreen() {
  const router = useRouter();
  const toast = useToast();
  const { color, space, font, size } = useTheme();
  const request = useChangeEmailRequest();
  const confirm = useChangeEmailConfirm();
  const [step, setStep] = useState<Step>('request');
  const [newEmail, setNewEmail] = useState('');
  const [otp, setOtp] = useState('');
  const [error, setError] = useState<string | null>(null);

  function submitRequest() {
    const parsed = ChangeEmailRequestSchema.safeParse({ newEmail });
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? 'E-mail inválido');
      return;
    }
    setError(null);
    request.mutate(parsed.data, {
      onSuccess: () => {
        toast.show('Código enviado para o novo e-mail.', 'info');
        setStep('confirm');
      },
      onError: (err) => toast.show(authErrorMessage(err), 'error'),
    });
  }

  function submitConfirm() {
    const parsed = ChangeEmailConfirmSchema.safeParse({ newEmail, otp });
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? 'Código inválido');
      return;
    }
    setError(null);
    confirm.mutate(parsed.data, {
      onSuccess: () => {
        toast.show('E-mail atualizado.', 'success');
        router.back();
      },
      onError: (err) => setError(authErrorMessage(err)),
    });
  }

  return (
    <ScreenContainer>
      <ScrollView
        contentContainerStyle={{ padding: space[5], flexGrow: 1 }}
        keyboardShouldPersistTaps="handled"
      >
        <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['2xl'], marginBottom: space[4] }}>
          {t.perfil.changeEmail}
        </Text>

        {step === 'request' ? (
          <>
            <FormField
              label="Novo e-mail"
              value={newEmail}
              onChangeText={setNewEmail}
              keyboardType="email-address"
              autoCapitalize="none"
              autoComplete="email"
              error={error ?? undefined}
            />
            <Button
              label="Enviar código"
              onPress={submitRequest}
              loading={request.isPending}
              fullWidth
            />
          </>
        ) : (
          <>
            <Text style={{ color: color.ink[2], fontFamily: font.body, fontSize: size.md, marginBottom: space[3] }}>
              Enviamos um código para {newEmail}.
            </Text>
            <OtpInput value={otp} onChange={setOtp} autoFocus error={error ?? undefined} />
            <View style={{ height: space[4] }} />
            <Button label="Confirmar" onPress={submitConfirm} loading={confirm.isPending} fullWidth />
          </>
        )}
      </ScrollView>
    </ScreenContainer>
  );
}
