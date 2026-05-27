import { useMemo } from 'react';
import { View } from 'react-native';
import { Redirect, Tabs, usePathname, useRouter } from 'expo-router';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useSessionStore } from '@/stores/sessionStore';
import { ScreenLoader } from '@/components/ScreenLoader';
import { TopStatusBar } from '@/components/TopStatusBar';
import { BottomNavigation, type TabKey } from '@/components/BottomNavigation';
import { useTheme } from '@/theme';

function pathToTab(pathname: string): TabKey {
  if (pathname.startsWith('/redacao')) return 'redacao';
  if (pathname.startsWith('/simulado')) return 'simulado';
  if (pathname.startsWith('/perfil')) return 'perfil';
  return 'trilha';
}

export default function TabsLayout() {
  const status = useSessionStore((s) => s.status);
  const student = useSessionStore((s) => s.student);
  const pathname = usePathname();
  const router = useRouter();
  const { color } = useTheme();

  const current = useMemo(() => pathToTab(pathname), [pathname]);

  if (status === 'loading') return <ScreenLoader />;
  if (status === 'anon') return <Redirect href="/(auth)/welcome" />;
  if (!student?.onboardingCompleted) return <Redirect href="/(onboarding)/welcome" />;

  return (
    <SafeAreaView
      edges={['top', 'bottom', 'left', 'right']}
      style={{ flex: 1, backgroundColor: color.paper[0] }}
    >
      <TopStatusBar
        streak={student.streakCount}
        xp={student.xpTotal}
        avatarUrl={student.avatarUrl ?? null}
        onAvatarPress={() => router.navigate('/(tabs)/perfil')}
      />
      <View style={{ flex: 1 }}>
        <Tabs
          screenOptions={{ headerShown: false }}
          tabBar={() => (
            <BottomNavigation
              current={current}
              onSelect={(tab) => router.navigate(`/(tabs)/${tab}` as never)}
            />
          )}
        >
          <Tabs.Screen name="trilha" />
          <Tabs.Screen name="redacao" />
          <Tabs.Screen name="simulado" />
          <Tabs.Screen name="perfil" />
        </Tabs>
      </View>
    </SafeAreaView>
  );
}
