import { useState } from 'react';
import { ScrollView, View } from 'react-native';
import { useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { FormField } from '@/components/FormField';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import { useLogin } from '@/features/auth/hooks';
import { LoginFormSchema } from '@/features/auth/validation';
import { authErrorMessage } from '@/features/auth/errorMessages';
import { t } from '@/lib/i18n/pt-BR';
import type { AppError } from '@/lib/api/errors';
import { Text } from 'react-native';

export default function LoginScreen() {
  const router = useRouter();
  const { color, space, font, size } = useTheme();
  const toast = useToast();
  const login = useLogin();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [fieldErrors, setFieldErrors] = useState<{ email?: string; password?: string }>({});

  function onSubmit() {
    const parsed = LoginFormSchema.safeParse({ email, password });
    if (!parsed.success) {
      const errs: typeof fieldErrors = {};
      for (const issue of parsed.error.issues) {
        const k = issue.path[0] as keyof typeof fieldErrors;
        errs[k] = issue.message;
      }
      setFieldErrors(errs);
      return;
    }
    setFieldErrors({});
    login.mutate(parsed.data, {
      onSuccess: () => router.replace('/(tabs)/trilha'),
      onError: (err) => {
        const e = err as unknown as AppError;
        if (e.kind === 'Validation') {
          setFieldErrors({ [e.field as 'email' | 'password']: e.message });
          return;
        }
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
      <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['3xl'], marginBottom: space[4] }}>
        {t.auth.login}
      </Text>

      <FormField
        label={t.auth.email}
        value={email}
        onChangeText={setEmail}
        keyboardType="email-address"
        autoCapitalize="none"
        autoComplete="email"
        error={fieldErrors.email}
      />
      <FormField
        label={t.auth.password}
        value={password}
        onChangeText={setPassword}
        secureTextEntry
        autoComplete="password"
        error={fieldErrors.password}
      />

      <Button label={t.auth.login} onPress={onSubmit} loading={login.isPending} fullWidth />
      <View style={{ height: space[3] }} />
      <Button
        label={t.auth.forgotPassword}
        variant="ghost"
        onPress={() => router.push('/(auth)/password-reset')}
        fullWidth
      />
      <Button
        label={t.auth.otpLogin}
        variant="ghost"
        onPress={() => router.push('/(auth)/otp-login')}
        fullWidth
      />
      <View style={{ height: space[4] }} />
      <Button
        label={t.auth.register}
        variant="secondary"
        onPress={() => router.push('/(auth)/register')}
        fullWidth
      />
      </ScrollView>
    </ScreenContainer>
  );
}
