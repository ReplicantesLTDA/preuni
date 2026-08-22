import { RefreshControl, ScrollView, Text, View } from 'react-native';
import { useTheme } from '@/theme';
import { ErrorState } from '@/components/ErrorState';
import { Skeleton } from '@/components/Skeleton';
import { useToast } from '@/components/Toast';
import { useTrilhaHome, useRefreshTrilha } from '@/features/trilha/hooks';
import { NextActivityCard } from '@/features/trilha/components/NextActivityCard';
import { ReadinessGauge } from '@/features/trilha/components/ReadinessGauge';
import { SubjectGrid } from '@/features/trilha/components/SubjectGrid';
import { t } from '@/lib/i18n/pt-BR';

export default function TrilhaHome() {
  const { color, space, font, size } = useTheme();
  const toast = useToast();
  const query = useTrilhaHome();
  const refresh = useRefreshTrilha();

  if (query.isLoading && !query.data) {
    return (
      <ScrollView
        contentContainerStyle={{ padding: space[5], backgroundColor: color.paper[0] }}
      >
        <Skeleton height={32} width="60%" style={{ marginBottom: space[4] }} />
        <Skeleton height={140} style={{ marginBottom: space[4] }} />
        <Skeleton height={80} style={{ marginBottom: space[4] }} />
        <Skeleton height={200} />
      </ScrollView>
    );
  }

  if (query.isError && !query.data) {
    return (
      <View style={{ flex: 1, backgroundColor: color.paper[0] }}>
        <ErrorState
          title="Sem conexão"
          body="Não conseguimos carregar sua trilha. Verifique sua internet."
          onRetry={() => query.refetch()}
        />
      </View>
    );
  }

  const student = query.data;
  if (!student) return null;

  const greeting = computeGreeting(student.displayName);

  return (
    <ScrollView
      contentContainerStyle={{ padding: space[5], paddingBottom: space[8], backgroundColor: color.paper[0] }}
      refreshControl={
        <RefreshControl
          refreshing={query.isFetching && !query.isLoading}
          onRefresh={() => refresh()}
          tintColor={color.accent}
        />
      }
    >
      <Text
        style={{
          color: color.ink[1],
          fontFamily: font.display,
          fontSize: size['3xl'],
          marginBottom: space[1],
        }}
      >
        {greeting}
      </Text>
      <Text
        style={{
          color: color.ink[2],
          fontFamily: font.body,
          fontSize: size.md,
          marginBottom: space[4],
        }}
      >
        {t.trilha.streak(student.streakCount)} · {t.trilha.xp(student.xpTotal)}
      </Text>

      <NextActivityCard
        title={t.trilha.nextStep}
        body="Continue de onde parou ou escolha uma matéria abaixo."
        ctaLabel="Em breve"
        onPress={() => toast.show('Trilha de atividades em construção.', 'info')}
        mascot="cheering"
      />

      <View style={{ height: space[4] }} />

      <ReadinessGauge score={student.readinessScore} />

      <View style={{ height: space[5] }} />

      <Text
        style={{
          color: color.ink[1],
          fontFamily: font.heading,
          fontSize: size.lg,
          marginBottom: space[2],
        }}
      >
        Matérias
      </Text>
      <SubjectGrid
        onSelect={(s) =>
          toast.show(`${s.label}: conteúdo em breve.`, 'info')
        }
      />
    </ScrollView>
  );
}

function computeGreeting(name: string): string {
  const hour = new Date().getHours();
  const first = name.split(' ')[0] ?? name;
  if (hour < 12) return `Bom dia, ${first}!`;
  if (hour < 18) return `Boa tarde, ${first}!`;
  return `Boa noite, ${first}!`;
}
