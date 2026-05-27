import { Image } from 'expo-image';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { useTheme } from '@/theme';

export interface TopStatusBarProps {
  streak: number;
  xp: number;
  avatarUrl?: string | null;
  onAvatarPress?: () => void;
  placeholder?: boolean;
}

export function TopStatusBar(props: TopStatusBarProps) {
  const { streak, xp, avatarUrl, onAvatarPress, placeholder = false } = props;
  const { color, space, font, size, radius } = useTheme();
  const numberColor = placeholder ? color.ink.muted : color.ink[0];

  return (
    <View
      accessibilityRole="header"
      style={[
        styles.container,
        {
          backgroundColor: color.paper[0],
          paddingHorizontal: space[4],
          paddingVertical: space[3],
          borderBottomColor: color.paper[1],
        },
      ]}
    >
      <View style={styles.metrics}>
        <View style={[styles.chip, { borderColor: color.ink.muted, borderRadius: radius.pill, paddingHorizontal: space[3], paddingVertical: space[1] }]}>
          <Text style={{ color: color.accent, fontFamily: font.numericBold, fontSize: size.md }}>🔥</Text>
          <Text
            accessibilityLabel={`${streak} dias de ofensiva`}
            style={{ color: numberColor, fontFamily: font.numericBold, fontSize: size.md, marginLeft: space[1] }}
          >
            {streak}
          </Text>
        </View>
        <View style={[styles.chip, { borderColor: color.ink.muted, borderRadius: radius.pill, paddingHorizontal: space[3], paddingVertical: space[1], marginLeft: space[2] }]}>
          <Text style={{ color: color.ink[1], fontFamily: font.heading, fontSize: size.sm }}>XP</Text>
          <Text
            accessibilityLabel={`${xp} XP`}
            style={{ color: numberColor, fontFamily: font.numericBold, fontSize: size.md, marginLeft: space[1] }}
          >
            {xp}
          </Text>
        </View>
      </View>

      <Pressable
        accessibilityRole="button"
        accessibilityLabel="Abrir perfil"
        onPress={onAvatarPress}
        style={[styles.avatar, { width: 36, height: 36, borderRadius: 18, backgroundColor: color.paper[1] }]}
      >
        {avatarUrl ? (
          <Image source={{ uri: avatarUrl }} style={{ width: 36, height: 36, borderRadius: 18 }} contentFit="cover" />
        ) : (
          <Text style={{ color: color.ink.muted, fontFamily: font.heading, fontSize: size.md }}>👤</Text>
        )}
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    borderBottomWidth: 1,
  },
  metrics: { flexDirection: 'row', alignItems: 'center' },
  chip: { borderWidth: 1, flexDirection: 'row', alignItems: 'center' },
  avatar: { alignItems: 'center', justifyContent: 'center' },
});
