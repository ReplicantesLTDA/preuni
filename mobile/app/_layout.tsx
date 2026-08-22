import { useEffect } from 'react';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { Slot } from 'expo-router';
import { QueryClientProvider } from '@tanstack/react-query';
import { StatusBar } from 'expo-status-bar';
import * as SplashScreen from 'expo-splash-screen';
import { ThemeProvider, useAppFonts } from '@/theme';
import { ApiProvider, useApi } from '@/lib/api/context';
import { createQueryClient } from '@/lib/query/client';
import { bootstrapSession } from '@/lib/auth/bootstrap';
import { ToastProvider } from '@/components/Toast';
import { ScreenLoader } from '@/components/ScreenLoader';

void SplashScreen.preventAutoHideAsync();

const queryClient = createQueryClient();
const apiBaseUrl = process.env.EXPO_PUBLIC_API_BASE_URL ?? 'http://localhost:8080';

function Bootstrap({ children }: { children: React.ReactNode }) {
  const api = useApi();
  useEffect(() => {
    void bootstrapSession(api).finally(() => {
      void SplashScreen.hideAsync();
    });
  }, [api]);
  return <>{children}</>;
}

export default function RootLayout() {
  const { fontsReady } = useAppFonts();

  if (!fontsReady) {
    return (
      <ThemeProvider>
        <ScreenLoader />
      </ThemeProvider>
    );
  }

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <SafeAreaProvider>
        <ThemeProvider>
          <QueryClientProvider client={queryClient}>
            <ApiProvider baseUrl={apiBaseUrl}>
              <ToastProvider>
                <StatusBar style="dark" />
                <Bootstrap>
                  <Slot />
                </Bootstrap>
              </ToastProvider>
            </ApiProvider>
          </QueryClientProvider>
        </ThemeProvider>
      </SafeAreaProvider>
    </GestureHandlerRootView>
  );
}
