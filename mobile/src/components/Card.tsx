import { StyleSheet, View, type ViewProps } from 'react-native';
import { useTheme } from '@/theme';

export interface CardProps extends ViewProps {
  padded?: boolean;
}

export function Card({ children, style, padded = true, ...rest }: CardProps) {
  const { color, radius, space, shadow } = useTheme();
  return (
    <View
      {...rest}
      style={[
        styles.base,
        {
          backgroundColor: color.paper[1],
          borderRadius: radius.md,
          padding: padded ? space[4] : 0,
          ...shadow.card,
        },
        style,
      ]}
    >
      {children}
    </View>
  );
}

const styles = StyleSheet.create({ base: { width: '100%' } });
