import { ScrollView, Text } from 'react-native';
import { Button } from '@/components/Button';
import { Card } from '@/components/Card';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import { useRequestDataExport } from '@/features/perfil/hooks';
import { authErrorMessage } from '@/features/auth/errorMessages';
import { t } from '@/lib/i18n/pt-BR';

export default function DataExportScreen() {
  const toast = useToast();
  const { color, space, font, size } = useTheme();
  const exportData = useRequestDataExport();

  function onExport() {
    exportData.mutate(undefined, {
      onSuccess: () => toast.show('Exportação solicitada. Verifique seu e-mail.', 'success'),
      onError: (err) => toast.show(authErrorMessage(err), 'error'),
    });
  }

  return (
    <ScreenContainer>
      <ScrollView
        contentContainerStyle={{ padding: space[5], flexGrow: 1 }}
        keyboardShouldPersistTaps="handled"
      >
        <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['2xl'], marginBottom: space[4] }}>
          {t.perfil.dataExport}
        </Text>
        <Card>
          <Text style={{ color: color.ink[2], fontFamily: font.body, fontSize: size.md, marginBottom: space[3] }}>
            Solicitar uma cópia dos seus dados. Enviaremos um e-mail com o arquivo quando estiver pronto.
          </Text>
          <Button label="Solicitar exportação" onPress={onExport} loading={exportData.isPending} fullWidth />
        </Card>
      </ScrollView>
    </ScreenContainer>
  );
}
