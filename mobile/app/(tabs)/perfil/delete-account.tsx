import { useState } from 'react';
import { ScrollView, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { Card } from '@/components/Card';
import { FormField } from '@/components/FormField';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import { useDeleteMe } from '@/features/perfil/hooks';
import { authErrorMessage } from '@/features/auth/errorMessages';
import { t } from '@/lib/i18n/pt-BR';

const CONFIRM_WORD = 'EXCLUIR';

export default function DeleteAccountScreen() {
  const router = useRouter();
  const toast = useToast();
  const { color, space, font, size } = useTheme();
  const del = useDeleteMe();
  const [confirmation, setConfirmation] = useState('');
  const [error, setError] = useState<string | null>(null);

  function onDelete() {
    if (confirmation !== CONFIRM_WORD) {
      setError(`Digite ${CONFIRM_WORD} para confirmar.`);
      return;
    }
    setError(null);
    del.mutate('DELETE', {
      onSuccess: () => {
        toast.show('Conta excluída.', 'info');
        router.replace('/(auth)/welcome');
      },
      onError: (err) => toast.show(authErrorMessage(err), 'error'),
    });
  }

  return (
    <ScreenContainer>
      <ScrollView
        contentContainerStyle={{ padding: space[5], flexGrow: 1 }}
        keyboardShouldPersistTaps="handled"
      >
        <Text style={{ color: color.danger, fontFamily: font.display, fontSize: size['2xl'], marginBottom: space[4] }}>
          {t.perfil.deleteAccount}
        </Text>
        <Card>
          <Text style={{ color: color.ink[1], fontFamily: font.body, fontSize: size.md }}>
            Esta ação é irreversível. Toda sua progressão, redações e simulados serão removidos.
          </Text>
          <View style={{ height: space[4] }} />
          <FormField
            label={`Digite "${CONFIRM_WORD}" para confirmar`}
            value={confirmation}
            onChangeText={setConfirmation}
            autoCapitalize="characters"
            error={error ?? undefined}
          />
          <Button
            label="Excluir minha conta"
            variant="danger"
            onPress={onDelete}
            loading={del.isPending}
            fullWidth
          />
        </Card>
      </ScrollView>
    </ScreenContainer>
  );
}
