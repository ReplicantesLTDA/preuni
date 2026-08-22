import { Pressable, ScrollView, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { Card } from '@/components/Card';
import { useTheme } from '@/theme';
import { useSessionStore } from '@/stores/sessionStore';
import { useLogout } from '@/features/auth/hooks';
import { t } from '@/lib/i18n/pt-BR';

interface MenuLinkProps {
  label: string;
  onPress: () => void;
  destructive?: boolean;
}

function MenuLink({ label, onPress, destructive }: MenuLinkProps) {
  const { color, font, size, space } = useTheme();
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={label}
      onPress={onPress}
      style={({ pressed }) => ({
        paddingVertical: space[3],
        opacity: pressed ? 0.7 : 1,
      })}
    >
      <Text
        style={{
          color: destructive ? color.danger : color.ink[1],
          fontFamily: font.heading,
          fontSize: size.md,
        }}
      >
        {label}
      </Text>
    </Pressable>
  );
}

function Divider() {
  const { color } = useTheme();
  return <View style={{ height: 1, backgroundColor: color.paper[0] }} />;
}

export default function PerfilHome() {
  const router = useRouter();
  const { color, space, font, size } = useTheme();
  const student = useSessionStore((s) => s.student);
  const logout = useLogout();

  return (
    <ScrollView
      contentContainerStyle={{ padding: space[5], paddingBottom: space[8], backgroundColor: color.paper[0] }}
    >
      <Text
        style={{
          color: color.ink[1],
          fontFamily: font.display,
          fontSize: size['3xl'],
          marginBottom: space[4],
        }}
      >
        {t.perfil.title}
      </Text>

      <Card>
        <Text
          style={{ color: color.ink[1], fontFamily: font.heading, fontSize: size.lg }}
          accessibilityLabel="Nome"
        >
          {student?.displayName ?? '—'}
        </Text>
        <Text
          style={{
            color: color.ink[2],
            fontFamily: font.body,
            fontSize: size.sm,
            marginTop: space[1],
          }}
        >
          {student?.email ?? '—'}
        </Text>
        <View style={{ flexDirection: 'row', gap: space[4], marginTop: space[4] }}>
          <Stat label="XP" value={student?.xpTotal ?? 0} />
          <Stat label="Ofensiva" value={student?.streakCount ?? 0} />
          <Stat label="Prontidão" value={Math.round(student?.readinessScore ?? 0)} />
        </View>
      </Card>

      <View style={{ height: space[5] }} />

      <Card>
        <MenuLink label={t.social.title} onPress={() => router.push('/(tabs)/perfil/friends')} />
        <Divider />
        <MenuLink label={t.gamification.title} onPress={() => router.push('/(tabs)/perfil/ranking')} />
        <Divider />
        <MenuLink label="Editar perfil" onPress={() => router.push('/(tabs)/perfil/edit')} />
        <Divider />
        <MenuLink label={t.perfil.changeEmail} onPress={() => router.push('/(tabs)/perfil/change-email')} />
        <Divider />
        <MenuLink label={t.perfil.changePassword} onPress={() => router.push('/(tabs)/perfil/change-password')} />
        <Divider />
        <MenuLink label={t.perfil.dataExport} onPress={() => router.push('/(tabs)/perfil/data-export')} />
        <Divider />
        <MenuLink
          label={t.perfil.deleteAccount}
          destructive
          onPress={() => router.push('/(tabs)/perfil/delete-account')}
        />
      </Card>

      <View style={{ height: space[6] }} />

      <Button
        label={t.perfil.signOut}
        variant="danger"
        onPress={() => logout.mutate()}
        loading={logout.isPending}
        fullWidth
      />
    </ScrollView>
  );
}

function Stat({ label, value }: { label: string; value: number }) {
  const { color, font, size, space } = useTheme();
  return (
    <View>
      <Text
        style={{ color: color.ink.muted, fontFamily: font.heading, fontSize: size.xs }}
      >
        {label}
      </Text>
      <Text
        style={{
          color: color.ink[0],
          fontFamily: font.numericBold,
          fontSize: size.xl,
          marginTop: space[1],
        }}
      >
        {value}
      </Text>
    </View>
  );
}
