import { ActivityIndicator, View } from 'react-native';
import { useTheme } from '@/theme';

export function ScreenLoader() {
  const { color } = useTheme();
  return (
    <View
      accessibilityRole="progressbar"
      accessibilityLabel="Carregando"
      style={{ flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: color.paper[0] }}
    >
      <ActivityIndicator color={color.accent} size="large" />
    </View>
  );
}
