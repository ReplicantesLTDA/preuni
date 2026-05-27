import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { useTheme } from '@/theme';

type ToastVariant = 'success' | 'error' | 'info';
interface ToastMessage {
  id: number;
  variant: ToastVariant;
  message: string;
}

interface ToastContextValue {
  show: (message: string, variant?: ToastVariant) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);
  const { color, radius, space, font, size } = useTheme();

  const show = useCallback((message: string, variant: ToastVariant = 'info') => {
    const id = Date.now() + Math.random();
    setToasts((prev) => [...prev, { id, variant, message }]);
    setTimeout(() => setToasts((prev) => prev.filter((t) => t.id !== id)), 4000);
  }, []);

  const value = useMemo(() => ({ show }), [show]);

  return (
    <ToastContext.Provider value={value}>
      {children}
      <View pointerEvents="none" style={[styles.host, { padding: space[4] }]}>
        {toasts.map((t) => (
          <View
            key={t.id}
            accessibilityRole="alert"
            style={[
              styles.toast,
              {
                backgroundColor:
                  t.variant === 'success'
                    ? color.success
                    : t.variant === 'error'
                      ? color.danger
                      : color.ink[1],
                borderRadius: radius.md,
                paddingHorizontal: space[4],
                paddingVertical: space[3],
                marginTop: space[2],
              },
            ]}
          >
            <Text style={{ color: color.paper[0], fontFamily: font.body, fontSize: size.md }}>
              {t.message}
            </Text>
          </View>
        ))}
      </View>
    </ToastContext.Provider>
  );
}

export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error('useToast must be used within <ToastProvider>');
  return ctx;
}

const styles = StyleSheet.create({
  host: { position: 'absolute', left: 0, right: 0, bottom: 0 },
  toast: { alignSelf: 'center', maxWidth: 480 },
});
