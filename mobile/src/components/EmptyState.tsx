import { View, Text, StyleSheet } from 'react-native';
import { Button } from './Button';
import { MascotPlaceholder, type MascotVariant } from './MascotPlaceholder';
import { useTheme } from '@/theme';

export interface EmptyStateProps {
  title: string;
  body?: string;
  mascot?: MascotVariant;
  cta?: { label: string; onPress: () => void };
}

export function EmptyState({ title, body, mascot = 'thinking', cta }: EmptyStateProps) {
  const { color, space, font, size } = useTheme();
  return (
    <View style={[styles.container, { padding: space[5] }]}>
      <MascotPlaceholder variant={mascot} />
      <Text
        style={{
          color: color.ink[1],
          fontFamily: font.display,
          fontSize: size['2xl'],
          marginTop: space[4],
          textAlign: 'center',
        }}
      >
        {title}
      </Text>
      {body ? (
        <Text
          style={{
            color: color.ink[2],
            fontFamily: font.body,
            fontSize: size.md,
            marginTop: space[2],
            textAlign: 'center',
          }}
        >
          {body}
        </Text>
      ) : null}
      {cta ? (
        <View style={{ marginTop: space[5] }}>
          <Button label={cta.label} onPress={cta.onPress} />
        </View>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { alignItems: 'center', justifyContent: 'center' },
});
