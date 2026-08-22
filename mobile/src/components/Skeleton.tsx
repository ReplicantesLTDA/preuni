import { View, type ViewStyle } from 'react-native';
import { useTheme } from '@/theme';

export interface SkeletonProps {
  height?: number;
  width?: number | `${number}%` | 'auto';
  borderRadius?: number;
  style?: ViewStyle;
}

export function Skeleton({ height = 12, width = '100%', borderRadius, style }: SkeletonProps) {
  const { color, radius } = useTheme();
  return (
    <View
      accessibilityElementsHidden
      style={[
        {
          height,
          width: width as number | `${number}%`,
          backgroundColor: color.paper[1],
          borderRadius: borderRadius ?? radius.sm,
          opacity: 0.7,
        },
        style,
      ]}
    />
  );
}
