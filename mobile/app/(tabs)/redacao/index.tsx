import { ScrollView, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { Card } from '@/components/Card';
import { EmptyState } from '@/components/EmptyState';
import { useTheme } from '@/theme';
import { t } from '@/lib/i18n/pt-BR';
import { useEssayList } from '@/features/essay/hooks';
import { useStreak } from '@/features/streak/hooks';

const SAMPLE_PROMPTS = [
  {
    id: 'enem-2024',
    title: 'Desafios para a valorização dos povos indígenas',
    year: 'ENEM 2024',
  },
  {
    id: 'enem-2023',
    title: 'Trabalho de cuidado realizado pela mulher no Brasil',
    year: 'ENEM 2023',
  },
  {
    id: 'enem-2022',
    title: 'Desafios para o enfrentamento da invisibilidade do trabalho de cuidado',
    year: 'ENEM 2022',
  },
];

export default function RedacaoHome() {
  const { color, space, font, size } = useTheme();
  const router = useRouter();
  const { data: streak } = useStreak();
  const { data: essays } = useEssayList();

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
        {t.redacao.title}
      </Text>
      <Text
        style={{
          color: color.ink[2],
          fontFamily: font.body,
          fontSize: size.md,
          marginBottom: space[4],
        }}
      >
        {streak ? t.trilha.streak(streak.currentStreak) : 'Pratique com temas do ENEM e receba feedback.'}
      </Text>

      <Card>
        <Text
          style={{
            color: color.ink[1],
            fontFamily: font.heading,
            fontSize: size.lg,
            marginBottom: space[1],
          }}
        >
          Sua próxima redação
        </Text>
        <Text
          style={{
            color: color.ink[2],
            fontFamily: font.body,
            fontSize: size.sm,
            marginBottom: space[3],
          }}
        >
          Escolha um tema abaixo e comece quando quiser.
        </Text>
        <Button label={t.redacao.writeNow} onPress={() => router.push('/redacao/write')} fullWidth />
      </Card>

      <View style={{ height: space[5] }} />

      <Text
        style={{
          color: color.ink[1],
          fontFamily: font.heading,
          fontSize: size.lg,
          marginBottom: space[2],
        }}
      >
        Temas sugeridos
      </Text>

      <View style={{ gap: space[3] }}>
        {SAMPLE_PROMPTS.map((p) => (
          <Card key={p.id}>
            <Text
              style={{
                color: color.accent,
                fontFamily: font.heading,
                fontSize: size.xs,
                marginBottom: space[1],
              }}
            >
              {p.year}
            </Text>
            <Text
              style={{
                color: color.ink[1],
                fontFamily: font.body,
                fontSize: size.md,
              }}
            >
              {p.title}
            </Text>
            <View style={{ height: space[3] }} />
            <Button
              label="Começar"
              variant="secondary"
              onPress={() => router.push({ pathname: '/redacao/write', params: { themeTitle: p.title, themeYear: p.year } })}
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
        Suas redações
      </Text>

      {essays && essays.length > 0 ? (
        <View style={{ gap: space[3] }}>
          {essays.map((e) => (
            <Card key={e.id}>
              <Button
                label={
                  e.status === 'pending'
                    ? t.redacao.statusPending
                    : e.status === 'failed'
                      ? t.redacao.statusFailed
                      : `Ver correção — ${new Date(e.submittedAt).toLocaleDateString('pt-BR')}`
                }
                variant="secondary"
                onPress={() => router.push(`/redacao/${e.id}`)}
                fullWidth
              />
            </Card>
          ))}
        </View>
      ) : (
        <EmptyState
          title={t.redacao.emptyTitle}
          body={t.redacao.emptyBody}
          mascot="reading"
        />
      )}
    </ScrollView>
  );
}
