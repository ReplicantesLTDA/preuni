import { Redirect, Stack } from 'expo-router';
import { useSessionStore } from '@/stores/sessionStore';
import { ScreenLoader } from '@/components/ScreenLoader';

export default function OnboardingLayout() {
  const status = useSessionStore((s) => s.status);
  const student = useSessionStore((s) => s.student);

  if (status === 'loading') return <ScreenLoader />;
  if (status === 'anon') return <Redirect href="/(auth)/welcome" />;
  if (student?.onboardingCompleted) return <Redirect href="/(tabs)/trilha" />;
  return <Stack screenOptions={{ headerShown: false }} />;
}
