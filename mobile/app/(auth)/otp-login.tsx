import { useState } from 'react';
import { ScrollView, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { FormField } from '@/components/FormField';
import { OtpInput } from '@/components/OtpInput';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import { useOtpLoginRequest, useOtpLoginVerify } from '@/features/auth/hooks';
import { authErrorMessage } from '@/features/auth/errorMessages';
import { OtpFormSchema, ResetRequestFormSchema } from '@/features/auth/validation';

type Step = 'email' | 'otp';

export default function OtpLoginScreen() {
  const router = useRouter();
  const { color, space, font, size } = useTheme();
  const toast = useToast();
  const request = useOtpLoginRequest();
  const verify = useOtpLoginVerify();
  const [step, setStep] = useState<Step>('email');
  const [email, setEmail] = useState('');
  const [otp, setOtp] = useState('');
  const [emailError, setEmailError] = useState<string | null>(null);
  const [otpError, setOtpError] = useState<string | null>(null);

  function submitEmail() {
    const parsed = ResetRequestFormSchema.safeParse({ email });
    if (!parsed.success) {
      setEmailError(parsed.error.issues[0]?.message ?? 'E-mail inválido');
      return;
    }
    setEmailError(null);
    request.mutate(parsed.data, {
      onSuccess: () => {
        setStep('otp');
        toast.show('Enviamos um código para seu e-mail.', 'info');
      },
      onError: (err) => toast.show(authErrorMessage(err), 'error'),
    });
  }

  function submitOtp() {
    const parsed = OtpFormSchema.safeParse({ otp });
    if (!parsed.success) {
      setOtpError(parsed.error.issues[0]?.message ?? 'Código inválido');
      return;
    }
    setOtpError(null);
    verify.mutate({ email, otp: parsed.data.otp }, {
      onSuccess: () => router.replace('/'),
      onError: (err) => setOtpError(authErrorMessage(err)),
    });
  }

  return (
    <ScreenContainer>
      <ScrollView
        contentContainerStyle={{ padding: space[5], flexGrow: 1, justifyContent: 'center' }}
        keyboardShouldPersistTaps="handled"
      >
      <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['2xl'], marginBottom: space[4] }}>
        Entrar com código
      </Text>

      {step === 'email' ? (
        <>
          <FormField
            label="E-mail"
            value={email}
            onChangeText={setEmail}
            keyboardType="email-address"
            autoCapitalize="none"
            autoComplete="email"
            error={emailError ?? undefined}
          />
          <Button label="Enviar código" onPress={submitEmail} loading={request.isPending} fullWidth />
        </>
      ) : (
        <>
          <Text style={{ color: color.ink[2], fontFamily: font.body, fontSize: size.md, marginBottom: space[4] }}>
            Enviamos um código para {email}.
          </Text>
          <OtpInput value={otp} onChange={setOtp} autoFocus error={otpError ?? undefined} />
          <View style={{ height: space[4] }} />
          <Button label="Entrar" onPress={submitOtp} loading={verify.isPending} fullWidth />
        </>
      )}
      </ScrollView>
    </ScreenContainer>
  );
}
