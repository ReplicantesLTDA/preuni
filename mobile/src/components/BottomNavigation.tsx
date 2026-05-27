import { Pressable, StyleSheet, Text, View } from 'react-native';
import { useTheme } from '@/theme';

export type TabKey = 'trilha' | 'redacao' | 'simulado' | 'perfil';

const LABELS: Record<TabKey, string> = {
  trilha: 'Trilha',
  redacao: 'Redação',
  simulado: 'Simulado',
  perfil: 'Perfil',
};
const ICONS: Record<TabKey, string> = {
  trilha: '🛤️',
  redacao: '✍️',
  simulado: '🧪',
  perfil: '👤',
};

export interface BottomNavigationProps {
  current: TabKey;
  onSelect: (tab: TabKey) => void;
}

export function BottomNavigation({ current, onSelect }: BottomNavigationProps) {
  const { color, space, font, size, radius } = useTheme();
  const tabs: TabKey[] = ['trilha', 'redacao', 'simulado', 'perfil'];

  return (
    <View
      accessibilityRole="tablist"
      style={[
        styles.bar,
        {
          backgroundColor: color.paper[0],
          borderTopColor: color.paper[1],
          paddingVertical: space[2],
        },
      ]}
    >
      {tabs.map((tab) => {
        const active = tab === current;
        return (
          <Pressable
            key={tab}
            accessibilityRole="tab"
            accessibilityState={{ selected: active }}
            accessibilityLabel={LABELS[tab]}
            onPress={() => onSelect(tab)}
            style={({ pressed }) => [
              styles.tab,
              {
                opacity: pressed ? 0.7 : 1,
                paddingHorizontal: space[3],
                paddingVertical: space[1],
                borderRadius: radius.sm,
              },
            ]}
          >
            <Text style={{ fontSize: size.xl }}>{ICONS[tab]}</Text>
            <Text
              style={{
                color: active ? color.accent : color.ink[2],
                fontFamily: active ? font.heading : font.body,
                fontSize: size.xs,
                marginTop: space[1],
              }}
            >
              {LABELS[tab]}
            </Text>
          </Pressable>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  bar: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    alignItems: 'center',
    borderTopWidth: 1,
  },
  tab: { alignItems: 'center', minWidth: 64 },
});
