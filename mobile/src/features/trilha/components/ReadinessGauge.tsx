import { StyleSheet, Text, View } from 'react-native';
import { Card } from '@/components/Card';
import { useTheme } from '@/theme';

export interface ReadinessGaugeProps {
  score: number; // 0..100
  label?: string;
}

export function ReadinessGauge({ score, label = 'Prontidão' }: ReadinessGaugeProps) {
  const { color, space, font, size, radius } = useTheme();
  const clamped = Math.max(0, Math.min(100, Math.round(score)));
  const fill = clamped / 100;

  return (
    <Card>
      <Text style={{ color: color.ink[1], fontFamily: font.heading, fontSize: size.md }}>
        {label}
      </Text>
      <View
        accessibilityRole="progressbar"
        accessibilityValue={{ min: 0, max: 100, now: clamped }}
        style={[
          styles.track,
          {
            backgroundColor: color.paper[0],
            borderColor: color.ink.muted,
            borderRadius: radius.pill,
            marginTop: space[2],
          },
        ]}
      >
        <View
          style={[
            styles.fill,
            {
              backgroundColor: color.accent,
              width: `${fill * 100}%`,
              borderRadius: radius.pill,
            },
          ]}
        />
      </View>
      <Text
        style={{
          color: color.ink[2],
          fontFamily: font.numericBold,
          fontSize: size.lg,
          marginTop: space[2],
        }}
      >
        {clamped}
        <Text style={{ color: color.ink.muted, fontFamily: font.body, fontSize: size.sm }}>
          {' / 100'}
        </Text>
      </Text>
    </Card>
  );
}

const styles = StyleSheet.create({
  track: { height: 12, borderWidth: 1, overflow: 'hidden' },
  fill: { height: '100%' },
});
