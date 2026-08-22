import { useState } from 'react';
import { ScrollView, Text } from 'react-native';
import { useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { FormField } from '@/components/FormField';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import { useRegister } from '@/features/auth/hooks';
import { RegisterFormSchema } from '@/features/auth/validation';
import { authErrorMessage } from '@/features/auth/errorMessages';
import { t } from '@/lib/i18n/pt-BR';
import type { AppError } from '@/lib/api/errors';

export default function RegisterScreen() {
  const router = useRouter();
  const { color, space, font, size } = useTheme();
  const toast = useToast();
  const register = useRegister();
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errs, setErrs] = useState<{ displayName?: string; email?: string; password?: string }>({});

  function onSubmit() {
    const parsed = RegisterFormSchema.safeParse({ displayName, email, password });
    if (!parsed.success) {
      const next: typeof errs = {};
      for (const issue of parsed.error.issues) {
        const k = issue.path[0] as keyof typeof errs;
        next[k] = issue.message;
      }
      setErrs(next);
      return;
    }
    setErrs({});
    register.mutate(parsed.data, {
      onSuccess: () => router.replace({ pathname: '/(auth)/verify-email', params: { email } }),
      onError: (err) => {
        const e = err as unknown as AppError;
        if (e.kind === 'Validation') {
          setErrs({ [e.field as 'email' | 'password' | 'displayName']: e.message });
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
        {t.auth.register}
      </Text>

      <FormField
        label={t.auth.displayName}
        value={displayName}
        onChangeText={setDisplayName}
        autoCapitalize="words"
        autoComplete="name"
        error={errs.displayName}
      />
      <FormField
        label={t.auth.email}
        value={email}
        onChangeText={setEmail}
        keyboardType="email-address"
        autoCapitalize="none"
        autoComplete="email"
        error={errs.email}
      />
      <FormField
        label={t.auth.password}
        value={password}
        onChangeText={setPassword}
        secureTextEntry
        autoComplete="new-password"
        error={errs.password}
        helper="Mín. 8 caracteres, 1 letra e 1 número."
      />

      <Button label={t.auth.submit} onPress={onSubmit} loading={register.isPending} fullWidth />
      </ScrollView>
    </ScreenContainer>
  );
}
