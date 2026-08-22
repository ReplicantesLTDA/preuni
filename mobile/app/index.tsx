import { Redirect } from 'expo-router';
import { ScreenLoader } from '@/components/ScreenLoader';
import { useSessionStore } from '@/stores/sessionStore';

export default function Index() {
  const status = useSessionStore((s) => s.status);
  const student = useSessionStore((s) => s.student);

  if (status === 'loading') return <ScreenLoader />;
  if (status === 'anon') return <Redirect href="/(auth)/welcome" />;
  if (!student?.onboardingCompleted) return <Redirect href="/(onboarding)/welcome" />;
  return <Redirect href="/(tabs)/trilha" />;
}
