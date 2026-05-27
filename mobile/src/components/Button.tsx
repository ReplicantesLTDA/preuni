import { ActivityIndicator, Pressable, StyleSheet, Text, View } from 'react-native';
import { useTheme } from '@/theme';

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';
export type ButtonSize = 'sm' | 'md' | 'lg';

export interface ButtonProps {
  label: string;
  onPress: () => void;
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
  disabled?: boolean;
  leftIcon?: React.ReactNode;
  fullWidth?: boolean;
  accessibilityLabel?: string;
  testID?: string;
}

export function Button(props: ButtonProps) {
  const {
    label,
    onPress,
    variant = 'primary',
    size = 'md',
    loading = false,
    disabled = false,
    leftIcon,
    fullWidth = false,
    accessibilityLabel,
    testID,
  } = props;
  const { color, radius, font, size: fontSize, space } = useTheme();

  const palette: Record<ButtonVariant, { bg: string; fg: string; border: string }> = {
    primary: { bg: color.accent, fg: color.ink[0], border: color.accent },
    secondary: { bg: color.paper[1], fg: color.ink[1], border: color.ink.muted },
    ghost: { bg: color.transparent, fg: color.ink[1], border: color.transparent },
    danger: { bg: color.danger, fg: color.paper[0], border: color.danger },
  };
  const sizing: Record<ButtonSize, { padV: number; padH: number; font: number }> = {
    sm: { padV: space[2], padH: space[3], font: fontSize.sm },
    md: { padV: space[3], padH: space[4], font: fontSize.md },
    lg: { padV: space[4], padH: space[5], font: fontSize.lg },
  };
  const { bg, fg, border } = palette[variant];
  const { padV, padH, font: fz } = sizing[size];

  const isDisabled = disabled || loading;
  const opacity = isDisabled ? 0.5 : 1;
  const handlePress = () => {
    if (isDisabled) return;
    onPress();
  };

  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={accessibilityLabel ?? label}
      accessibilityState={{ disabled: isDisabled, busy: loading }}
      disabled={isDisabled}
      onPress={handlePress}
      testID={testID}
      style={({ pressed }) => [
        styles.base,
        {
          backgroundColor: bg,
          borderColor: border,
          borderRadius: radius.md,
          paddingVertical: padV,
          paddingHorizontal: padH,
          opacity: pressed && !isDisabled ? 0.85 : opacity,
          alignSelf: fullWidth ? 'stretch' : 'auto',
        },
      ]}
    >
      <View style={styles.row}>
        {loading ? (
          <ActivityIndicator color={fg} />
        ) : (
          <>
            {leftIcon ? <View style={{ marginRight: space[2] }}>{leftIcon}</View> : null}
            <Text style={{ color: fg, fontFamily: font.heading, fontSize: fz }}>{label}</Text>
          </>
        )}
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  base: { borderWidth: 1, justifyContent: 'center', alignItems: 'center' },
  row: { flexDirection: 'row', alignItems: 'center', justifyContent: 'center' },
});
