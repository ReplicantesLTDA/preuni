import { useState } from 'react';
import { ScrollView, Text } from 'react-native';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { Button } from '@/components/Button';
import { FormField } from '@/components/FormField';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import { useSubmitEssay } from '@/features/essay/hooks';
import { SubmitEssaySchema } from '@/features/essay/validation';
import { t } from '@/lib/i18n/pt-BR';

export default function WriteEssayScreen() {
  const router = useRouter();
  const toast = useToast();
  const { color, space, font, size } = useTheme();
  const mutate = useSubmitEssay();
  const params = useLocalSearchParams<{ themeTitle?: string }>();

  const [promptThemeTitle, setThemeTitle] = useState(params.themeTitle ?? '');
  const [promptThemeContext, setThemeContext] = useState('');
  const [essayText, setEssayText] = useState('');
  const [errs, setErrs] = useState<{ promptThemeTitle?: string; promptThemeContext?: string; essayText?: string }>({});

  function onSubmit() {
    const parsed = SubmitEssaySchema.safeParse({ promptThemeTitle, promptThemeContext, essayText });
    if (!parsed.success) {
      const next: typeof errs = {};
      for (const issue of parsed.error.issues) {
        const k = issue.path[0] as keyof typeof errs;
        next[k] = issue.message;
      }
      setErrs(next);
      return;
    }
    setErrs({});
    mutate.mutate(parsed.data, {
      onSuccess: (submission) => {
        toast.show(t.redacao.submitted, 'success');
        router.replace(`/redacao/${submission.id}`);
      },
      onError: (err) => {
        const isQuota = typeof err === 'object' && err !== null && 'kind' in err && (err as { kind: string }).kind === 'RateLimited';
        toast.show(isQuota ? t.redacao.quotaExceeded : 'Não foi possível enviar sua redação.', 'error');
      },
    });
  }

  return (
    <ScreenContainer>
      <ScrollView
        contentContainerStyle={{ padding: space[5], flexGrow: 1 }}
        keyboardShouldPersistTaps="handled"
      >
        <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['2xl'], marginBottom: space[4] }}>
          {t.redacao.writeNow}
        </Text>
        <FormField
          label={t.redacao.themeTitle}
          value={promptThemeTitle}
          onChangeText={setThemeTitle}
          error={errs.promptThemeTitle}
        />
        <FormField
          label={t.redacao.themeContext}
          value={promptThemeContext}
          onChangeText={setThemeContext}
          multiline
          error={errs.promptThemeContext}
        />
        <FormField
          label={t.redacao.essayText}
          value={essayText}
          onChangeText={setEssayText}
          multiline
          numberOfLines={10}
          error={errs.essayText}
        />
        <Button label={t.redacao.submit} onPress={onSubmit} loading={mutate.isPending} fullWidth />
      </ScrollView>
    </ScreenContainer>
  );
}
