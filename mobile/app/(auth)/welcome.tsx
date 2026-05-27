import { View } from 'react-native';
import { useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { MascotPlaceholder } from '@/components/MascotPlaceholder';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useTheme } from '@/theme';
import { t } from '@/lib/i18n/pt-BR';

export default function AuthWelcome() {
  const router = useRouter();
  const { space } = useTheme();
  return (
    <ScreenContainer>
      <View style={{ flex: 1, padding: space[5], alignItems: 'center', justifyContent: 'center' }}>
        <MascotPlaceholder size={140} label="Preuni" />
        <View style={{ height: space[5] }} />
        <View style={{ width: '100%', maxWidth: 320, gap: space[3] }}>
          <Button label={t.auth.login} onPress={() => router.push('/(auth)/login')} fullWidth />
          <Button
            label={t.auth.register}
            variant="secondary"
            onPress={() => router.push('/(auth)/register')}
            fullWidth
          />
        </View>
      </View>
    </ScreenContainer>
  );
}
