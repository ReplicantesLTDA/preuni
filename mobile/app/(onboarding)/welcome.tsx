import { View } from 'react-native';
import { useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { EmptyState } from '@/components/EmptyState';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useTheme } from '@/theme';
import { t } from '@/lib/i18n/pt-BR';

export default function OnboardingWelcome() {
  const router = useRouter();
  const { space } = useTheme();
  return (
    <ScreenContainer>
      <View style={{ flex: 1, padding: space[5], justifyContent: 'center' }}>
        <EmptyState title={t.onboarding.welcome} mascot="cheering" />
        <View style={{ marginTop: space[6] }}>
          <Button
            label={t.common.continue}
            onPress={() => router.push('/(onboarding)/profile')}
            fullWidth
          />
        </View>
      </View>
    </ScreenContainer>
  );
}
