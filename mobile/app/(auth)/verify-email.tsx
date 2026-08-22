import { useEffect, useState } from 'react';
import { ScrollView, Text, View } from 'react-native';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { OtpInput } from '@/components/OtpInput';
import { MascotPlaceholder } from '@/components/MascotPlaceholder';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import { useResendVerificationOtp, useVerifyEmail } from '@/features/auth/hooks';
import { OtpFormSchema } from '@/features/auth/validation';
import { authErrorMessage } from '@/features/auth/errorMessages';
import { t } from '@/lib/i18n/pt-BR';

const COOLDOWN_SECONDS = 60;

export default function VerifyEmailScreen() {
  const router = useRouter();
  const params = useLocalSearchParams<{ email?: string }>();
  const email = typeof params.email === 'string' ? params.email : '';
  const { color, space, font, size } = useTheme();
  const toast = useToast();
  const verify = useVerifyEmail();
  const resend = useResendVerificationOtp();
  const [otp, setOtp] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [cooldown, setCooldown] = useState(COOLDOWN_SECONDS);

  useEffect(() => {
    if (cooldown <= 0) return;
    const id = setInterval(() => setCooldown((c) => Math.max(0, c - 1)), 1000);
    return () => clearInterval(id);
  }, [cooldown]);

  function onSubmit() {
    const parsed = OtpFormSchema.safeParse({ otp });
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? 'Código inválido');
      return;
    }
    setError(null);
    verify.mutate({ email, otp: parsed.data.otp }, {
      onSuccess: () => {
        toast.show(t.auth.verified, 'success');
        router.replace('/');
      },
      onError: (err) => {
        setError(authErrorMessage(err));
      },
    });
  }

  function onResend() {
    if (cooldown > 0) return;
    resend.mutate({ email }, {
      onSuccess: () => {
        toast.show('Enviamos um novo código.', 'info');
        setCooldown(COOLDOWN_SECONDS);
      },
      onError: (err) => {
        toast.show(authErrorMessage(err), 'error');
      },
    });
  }

  return (
    <ScreenContainer>
      <ScrollView
        contentContainerStyle={{ padding: space[5], flexGrow: 1, justifyContent: 'center' }}
        keyboardShouldPersistTaps="handled"
      >
      <View style={{ alignItems: 'center', marginBottom: space[5] }}>
        <MascotPlaceholder variant="reading" size={96} />
      </View>
      <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['2xl'], textAlign: 'center', marginBottom: space[2] }}>
        {t.auth.otpTitle}
      </Text>
      <Text style={{ color: color.ink[2], fontFamily: font.body, fontSize: size.md, textAlign: 'center', marginBottom: space[5] }}>
        {t.auth.otpSubtitle} {email}
      </Text>

      <OtpInput value={otp} onChange={setOtp} error={error ?? undefined} autoFocus />

      <View style={{ height: space[5] }} />
      <Button label={t.auth.submit} onPress={onSubmit} loading={verify.isPending} fullWidth />

      <View style={{ height: space[3] }} />
      <Button
        label={cooldown > 0 ? t.auth.resendCooldown(cooldown) : t.auth.resend}
        variant="ghost"
        onPress={onResend}
        disabled={cooldown > 0 || resend.isPending}
        loading={resend.isPending}
        fullWidth
      />
      </ScrollView>
    </ScreenContainer>
  );
}
