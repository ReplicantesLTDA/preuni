import { View, Text, StyleSheet } from 'react-native';
import Svg, { Circle, Path } from 'react-native-svg';
import { useTheme } from '@/theme';

export type MascotVariant = 'thinking' | 'reading' | 'cheering';

export interface MascotPlaceholderProps {
  variant?: MascotVariant;
  size?: number;
  label?: string;
}

export function MascotPlaceholder({
  variant = 'thinking',
  size = 96,
  label = 'Mascote',
}: MascotPlaceholderProps) {
  const { color } = useTheme();

  return (
    <View
      accessibilityRole="image"
      accessibilityLabel={`${label} — ${variant}`}
      style={[styles.container, { width: size, height: size }]}
    >
      <Svg viewBox="0 0 100 100" width={size} height={size}>
        <Circle cx="50" cy="50" r="44" fill={color.accent} opacity={0.18} />
        <Circle cx="50" cy="46" r="32" fill={color.paper[1]} stroke={color.ink[1]} strokeWidth={2} />
        <Circle cx="40" cy="44" r="3" fill={color.ink[0]} />
        <Circle cx="60" cy="44" r="3" fill={color.ink[0]} />
        {variant === 'cheering' ? (
          <Path
            d="M38 58 Q50 70 62 58"
            stroke={color.ink[0]}
            strokeWidth={2}
            fill="none"
            strokeLinecap="round"
          />
        ) : variant === 'reading' ? (
          <Path
            d="M40 58 H60"
            stroke={color.ink[0]}
            strokeWidth={2}
            fill="none"
            strokeLinecap="round"
          />
        ) : (
          <Path
            d="M40 58 Q50 56 60 58"
            stroke={color.ink[0]}
            strokeWidth={2}
            fill="none"
            strokeLinecap="round"
          />
        )}
      </Svg>
      <Text style={styles.hiddenLabel}>{label}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { alignItems: 'center', justifyContent: 'center' },
  hiddenLabel: { position: 'absolute', width: 1, height: 1, opacity: 0 },
});
