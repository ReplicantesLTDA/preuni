import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Card } from '@/components/Card';
import { MascotPlaceholder, type MascotVariant } from '@/components/MascotPlaceholder';
import { useTheme } from '@/theme';

export interface NextActivityCardProps {
  title: string;
  body?: string;
  ctaLabel: string;
  onPress: () => void;
  mascot?: MascotVariant;
  testID?: string;
}

export function NextActivityCard({
  title,
  body,
  ctaLabel,
  onPress,
  mascot = 'cheering',
  testID,
}: NextActivityCardProps) {
  const { color, space, font, size, radius } = useTheme();

  return (
    <Card>
      <View style={styles.row}>
        <MascotPlaceholder variant={mascot} size={80} label="Mascote" />
        <View style={{ flex: 1, marginLeft: space[3] }}>
          <Text
            style={{
              color: color.ink[1],
              fontFamily: font.display,
              fontSize: size.xl,
            }}
          >
            {title}
          </Text>
          {body ? (
            <Text
              style={{
                color: color.ink[2],
                fontFamily: font.body,
                fontSize: size.sm,
                marginTop: space[1],
              }}
            >
              {body}
            </Text>
          ) : null}
        </View>
      </View>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel={ctaLabel}
        onPress={onPress}
        testID={testID}
        style={({ pressed }) => [
          styles.cta,
          {
            backgroundColor: color.accent,
            borderRadius: radius.md,
            paddingVertical: space[3],
            marginTop: space[4],
            opacity: pressed ? 0.85 : 1,
          },
        ]}
      >
        <Text
          style={{
            color: color.ink[0],
            fontFamily: font.heading,
            fontSize: size.md,
            textAlign: 'center',
          }}
        >
          {ctaLabel}
        </Text>
      </Pressable>
    </Card>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center' },
  cta: { width: '100%' },
});
