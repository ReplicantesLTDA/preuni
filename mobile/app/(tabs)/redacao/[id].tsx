import { ActivityIndicator, ScrollView, Text, View } from 'react-native';
import { useLocalSearchParams } from 'expo-router';
import { Card } from '@/components/Card';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useTheme } from '@/theme';
import { useEssay } from '@/features/essay/hooks';
import { t } from '@/lib/i18n/pt-BR';

export default function EssayStatusScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const { color, space, font, size } = useTheme();
  const { data: essay, isLoading } = useEssay(id);

  return (
    <ScreenContainer>
      <ScrollView contentContainerStyle={{ padding: space[5], paddingBottom: space[8] }}>
        {isLoading || !essay ? (
          <ActivityIndicator color={color.accent} />
        ) : essay.status === 'pending' ? (
          <View style={{ alignItems: 'center', paddingVertical: space[8] }}>
            <ActivityIndicator color={color.accent} />
            <Text
              style={{
                color: color.ink[2],
                fontFamily: font.body,
                fontSize: size.md,
                marginTop: space[3],
              }}
            >
              {t.redacao.statusPending}
            </Text>
          </View>
        ) : essay.status === 'failed' ? (
          <Card>
            <Text style={{ color: color.danger, fontFamily: font.heading, fontSize: size.lg }}>
              {t.redacao.statusFailed}
            </Text>
          </Card>
        ) : (
          <>
            <Card>
              <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['3xl'] }}>
                {essay.grade?.overallScore}
                <Text style={{ fontSize: size.md, fontFamily: font.body, color: color.ink[2] }}> / 1000</Text>
              </Text>
              <Text style={{ color: color.ink[2], fontFamily: font.body, fontSize: size.sm }}>
                {t.redacao.overallScore}
              </Text>
            </Card>
            <View style={{ height: space[4] }} />
            <View style={{ gap: space[3] }}>
              {essay.grade?.competencies.map((c) => (
                <Card key={c.competency}>
                  <Text style={{ color: color.accent, fontFamily: font.heading, fontSize: size.xs, marginBottom: space[1] }}>
                    Competência {c.competency} — {c.score}/200
                  </Text>
                  <Text style={{ color: color.ink[1], fontFamily: font.body, fontSize: size.sm, marginBottom: space[1] }}>
                    {c.justificationPtBr}
                  </Text>
                  <Text style={{ color: color.ink[2], fontFamily: font.body, fontSize: size.xs, fontStyle: 'italic' }}>
                    “{c.excerpt}”
                  </Text>
                </Card>
              ))}
            </View>
          </>
        )}
      </ScrollView>
    </ScreenContainer>
  );
}
