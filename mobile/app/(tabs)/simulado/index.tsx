import { ScrollView, Text, View } from 'react-native';
import { Button } from '@/components/Button';
import { Card } from '@/components/Card';
import { EmptyState } from '@/components/EmptyState';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import { t } from '@/lib/i18n/pt-BR';

const SIMULATIONS = [
  { id: 'mini', label: 'Simulado rápido', duration: '15 min', questions: 10 },
  { id: 'matematica', label: 'Matemática focada', duration: '45 min', questions: 25 },
  { id: 'enem-full', label: 'ENEM completo', duration: '5h30', questions: 180 },
];

export default function SimuladoHome() {
  const { color, space, font, size } = useTheme();
  const toast = useToast();

  return (
    <ScrollView
      contentContainerStyle={{ padding: space[5], paddingBottom: space[8], backgroundColor: color.paper[0] }}
    >
      <Text
        style={{
          color: color.ink[1],
          fontFamily: font.display,
          fontSize: size['3xl'],
          marginBottom: space[1],
        }}
      >
        {t.simulado.title}
      </Text>
      <Text
        style={{
          color: color.ink[2],
          fontFamily: font.body,
          fontSize: size.md,
          marginBottom: space[4],
        }}
      >
        Treine sob pressão. Escolha um modo abaixo.
      </Text>

      <View style={{ gap: space[3] }}>
        {SIMULATIONS.map((s) => (
          <Card key={s.id}>
            <Text
              style={{
                color: color.ink[1],
                fontFamily: font.heading,
                fontSize: size.lg,
              }}
            >
              {s.label}
            </Text>
            <Text
              style={{
                color: color.ink[2],
                fontFamily: font.body,
                fontSize: size.sm,
                marginTop: space[1],
              }}
            >
              {s.questions} questões · {s.duration}
            </Text>
            <View style={{ height: space[3] }} />
            <Button
              label="Começar"
              onPress={() => toast.show(`${s.label}: em construção.`, 'info')}
              fullWidth
            />
          </Card>
        ))}
      </View>

      <View style={{ height: space[6] }} />

      <Text
        style={{
          color: color.ink[1],
          fontFamily: font.heading,
          fontSize: size.lg,
          marginBottom: space[2],
        }}
      >
        Seu histórico
      </Text>

      <EmptyState
        title={t.simulado.emptyTitle}
        body={t.simulado.emptyBody}
        mascot="thinking"
      />
    </ScrollView>
  );
}
