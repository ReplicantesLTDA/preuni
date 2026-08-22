import { Redirect, Stack } from 'expo-router';
import { useSessionStore } from '@/stores/sessionStore';
import { ScreenLoader } from '@/components/ScreenLoader';

export default function AuthLayout() {
  const status = useSessionStore((s) => s.status);
  if (status === 'loading') return <ScreenLoader />;
  if (status === 'authed') return <Redirect href="/(tabs)/trilha" />;
  return <Stack screenOptions={{ headerShown: false }} />;
}
