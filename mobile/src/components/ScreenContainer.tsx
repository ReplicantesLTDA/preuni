import { KeyboardAvoidingView, Platform, StyleSheet, View, type ViewStyle } from 'react-native';
import { SafeAreaView, type Edge } from 'react-native-safe-area-context';
import { useTheme } from '@/theme';

export interface ScreenContainerProps {
  children: React.ReactNode;
  edges?: Edge[];
  avoidKeyboard?: boolean;
  style?: ViewStyle;
}

export function ScreenContainer({
  children,
  edges = ['top', 'bottom', 'left', 'right'],
  avoidKeyboard = true,
  style,
}: ScreenContainerProps) {
  const { color } = useTheme();
  const body = (
    <View style={[styles.flex, { backgroundColor: color.paper[0] }, style]}>{children}</View>
  );

  return (
    <SafeAreaView
      edges={edges}
      style={[styles.flex, { backgroundColor: color.paper[0] }]}
    >
      {avoidKeyboard ? (
        <KeyboardAvoidingView
          behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
          style={styles.flex}
        >
          {body}
        </KeyboardAvoidingView>
      ) : (
        body
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({ flex: { flex: 1 } });
