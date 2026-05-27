import { useState } from 'react';
import { ScrollView, Text } from 'react-native';
import { useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { FormField } from '@/components/FormField';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import { useChangePassword } from '@/features/perfil/hooks';
import { ChangePasswordSchema } from '@/features/perfil/validation';
import { authErrorMessage } from '@/features/auth/errorMessages';
import { t } from '@/lib/i18n/pt-BR';

export default function ChangePasswordScreen() {
  const router = useRouter();
  const toast = useToast();
  const { color, space, font, size } = useTheme();
  const mutate = useChangePassword();
  const [currentPassword, setCurrent] = useState('');
  const [newPassword, setNew] = useState('');
  const [confirmPassword, setConfirm] = useState('');
  const [errs, setErrs] = useState<{ currentPassword?: string; newPassword?: string; confirmPassword?: string }>({});

  function onSubmit() {
    const parsed = ChangePasswordSchema.safeParse({ currentPassword, newPassword, confirmPassword });
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
    mutate.mutate(
      { currentPassword: parsed.data.currentPassword, newPassword: parsed.data.newPassword },
      {
        onSuccess: () => {
          toast.show('Senha atualizada.', 'success');
          router.back();
        },
        onError: (err) => toast.show(authErrorMessage(err), 'error'),
      },
    );
  }

  return (
    <ScreenContainer>
      <ScrollView
        contentContainerStyle={{ padding: space[5], flexGrow: 1 }}
        keyboardShouldPersistTaps="handled"
      >
        <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['2xl'], marginBottom: space[4] }}>
          {t.perfil.changePassword}
        </Text>
        <FormField
          label="Senha atual"
          value={currentPassword}
          onChangeText={setCurrent}
          secureTextEntry
          autoComplete="current-password"
          error={errs.currentPassword}
        />
        <FormField
          label="Nova senha"
          value={newPassword}
          onChangeText={setNew}
          secureTextEntry
          autoComplete="new-password"
          error={errs.newPassword}
          helper="Mín. 8 caracteres, 1 letra e 1 número."
        />
        <FormField
          label="Confirmar nova senha"
          value={confirmPassword}
          onChangeText={setConfirm}
          secureTextEntry
          autoComplete="new-password"
          error={errs.confirmPassword}
        />
        <Button label="Atualizar senha" onPress={onSubmit} loading={mutate.isPending} fullWidth />
      </ScrollView>
    </ScreenContainer>
  );
}
