import { useState } from 'react';
import { Pressable, ScrollView, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { Image } from 'expo-image';
import * as ImagePicker from 'expo-image-picker';
import { Button } from '@/components/Button';
import { FormField } from '@/components/FormField';
import { useToast } from '@/components/Toast';
import { MascotPlaceholder } from '@/components/MascotPlaceholder';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useTheme } from '@/theme';
import { useSessionStore } from '@/stores/sessionStore';
import { t } from '@/lib/i18n/pt-BR';

export default function OnboardingProfileScreen() {
  const router = useRouter();
  const student = useSessionStore((s) => s.student);
  const patchStudent = useSessionStore((s) => s.patchStudent);
  const { color, space, font, size, radius } = useTheme();
  const toast = useToast();

  const [displayName, setDisplayName] = useState(student?.displayName ?? '');
  const [avatarUri, setAvatarUri] = useState<string | null>(student?.avatarUrl ?? null);
  const [error, setError] = useState<string | null>(null);

  async function pickAvatar() {
    const perm = await ImagePicker.requestMediaLibraryPermissionsAsync();
    if (!perm.granted) {
      toast.show('Permissão de fotos negada.', 'error');
      return;
    }
    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ['images'],
      allowsEditing: true,
      aspect: [1, 1],
      quality: 0.8,
    });
    if (result.canceled || !result.assets[0]) return;
    setAvatarUri(result.assets[0].uri);
  }

  function onContinue() {
    const trimmed = displayName.trim();
    if (trimmed.length < 2) {
      setError('Mínimo de 2 caracteres.');
      return;
    }
    if (trimmed.length > 64) {
      setError('Máximo de 64 caracteres.');
      return;
    }
    setError(null);
    // Apply locally; remote sync happens during onboarding-complete or avatar upload mutation.
    patchStudent({ displayName: trimmed, avatarUrl: avatarUri });
    router.push('/(onboarding)/interests');
  }

  return (
    <ScreenContainer>
      <ScrollView
        contentContainerStyle={{ padding: space[5], flexGrow: 1 }}
        keyboardShouldPersistTaps="handled"
      >
      <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['2xl'], marginBottom: space[2] }}>
        Vamos te conhecer
      </Text>
      <Text style={{ color: color.ink[2], fontFamily: font.body, fontSize: size.md, marginBottom: space[5] }}>
        Confirme seu nome e escolha um avatar opcional.
      </Text>

      <View style={{ alignItems: 'center', marginBottom: space[5] }}>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="Escolher avatar"
          onPress={pickAvatar}
          style={{
            width: 120,
            height: 120,
            borderRadius: radius.pill,
            backgroundColor: color.paper[1],
            alignItems: 'center',
            justifyContent: 'center',
            overflow: 'hidden',
          }}
        >
          {avatarUri ? (
            <Image source={{ uri: avatarUri }} style={{ width: 120, height: 120 }} contentFit="cover" />
          ) : (
            <MascotPlaceholder variant="cheering" size={120} label="Avatar" />
          )}
        </Pressable>
        <Pressable onPress={pickAvatar} style={{ marginTop: space[2] }}>
          <Text style={{ color: color.accent, fontFamily: font.heading, fontSize: size.sm }}>
            {avatarUri ? 'Trocar foto' : 'Adicionar foto'}
          </Text>
        </Pressable>
      </View>

      <FormField
        label={t.auth.displayName}
        value={displayName}
        onChangeText={setDisplayName}
        autoCapitalize="words"
        autoComplete="name"
        error={error ?? undefined}
      />

      <Button label={t.common.continue} onPress={onContinue} fullWidth />
      </ScrollView>
    </ScreenContainer>
  );
}
