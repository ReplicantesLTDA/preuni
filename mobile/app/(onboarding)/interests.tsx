import { useState } from 'react';
import { Pressable, ScrollView, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import { useCompleteOnboarding } from '@/features/onboarding/hooks';
import { t } from '@/lib/i18n/pt-BR';
import { authErrorMessage } from '@/features/auth/errorMessages';

const SUBJECTS = [
  { id: 'matematica', label: 'Matemática' },
  { id: 'portugues', label: 'Português' },
  { id: 'redacao', label: 'Redação' },
  { id: 'ciencias-natureza', label: 'Ciências da Natureza' },
  { id: 'ciencias-humanas', label: 'Ciências Humanas' },
  { id: 'linguagens', label: 'Linguagens' },
];

export default function InterestsScreen() {
  const router = useRouter();
  const { color, space, font, size, radius } = useTheme();
  const toast = useToast();
  const complete = useCompleteOnboarding();
  const [selected, setSelected] = useState<Set<string>>(new Set());

  function toggle(id: string) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function onFinish() {
    complete.mutate(
      { interests: Array.from(selected) },
      {
        onSuccess: () => router.replace('/(tabs)/trilha'),
        onError: (err) => toast.show(authErrorMessage(err), 'error'),
      },
    );
  }

  return (
    <ScreenContainer>
      <ScrollView
        contentContainerStyle={{ padding: space[5], flexGrow: 1 }}
        keyboardShouldPersistTaps="handled"
      >
      <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['2xl'], marginBottom: space[4] }}>
        {t.onboarding.interests}
      </Text>

      <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: space[2] }}>
        {SUBJECTS.map((s) => {
          const active = selected.has(s.id);
          return (
            <Pressable
              key={s.id}
              accessibilityRole="checkbox"
              accessibilityState={{ checked: active }}
              onPress={() => toggle(s.id)}
              style={{
                borderWidth: 1,
                borderColor: active ? color.accent : color.ink.muted,
                backgroundColor: active ? color.accent : color.paper[1],
                borderRadius: radius.pill,
                paddingHorizontal: space[4],
                paddingVertical: space[2],
              }}
            >
              <Text style={{ color: active ? color.ink[0] : color.ink[1], fontFamily: font.heading, fontSize: size.sm }}>
                {s.label}
              </Text>
            </Pressable>
          );
        })}
      </View>

      <View style={{ marginTop: space[6] }}>
        <Button label={t.onboarding.finish} onPress={onFinish} loading={complete.isPending} fullWidth />
      </View>
      </ScrollView>
    </ScreenContainer>
  );
}
