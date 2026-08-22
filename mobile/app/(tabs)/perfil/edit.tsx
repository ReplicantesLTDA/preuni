import { useState } from 'react';
import { Pressable, ScrollView, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { Image } from 'expo-image';
import * as ImagePicker from 'expo-image-picker';
import { Button } from '@/components/Button';
import { FormField } from '@/components/FormField';
import { MascotPlaceholder } from '@/components/MascotPlaceholder';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import { useSessionStore } from '@/stores/sessionStore';
import { useUpdateMe, useUploadAvatar } from '@/features/perfil/hooks';
import { EditProfileSchema } from '@/features/perfil/validation';
import { authErrorMessage } from '@/features/auth/errorMessages';

export default function EditProfileScreen() {
  const router = useRouter();
  const student = useSessionStore((s) => s.student);
  const update = useUpdateMe();
  const uploadAvatar = useUploadAvatar();
  const toast = useToast();
  const { color, space, font, size, radius } = useTheme();
  const [displayName, setDisplayName] = useState(student?.displayName ?? '');
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
    const asset = result.assets[0];
    uploadAvatar.mutate(
      { uri: asset.uri, mimeType: asset.mimeType ?? 'image/jpeg' },
      {
        onSuccess: () => toast.show('Avatar atualizado.', 'success'),
        onError: (err) => toast.show(authErrorMessage(err), 'error'),
      },
    );
  }

  function onSave() {
    const parsed = EditProfileSchema.safeParse({ displayName });
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? 'inválido');
      return;
    }
    setError(null);
    update.mutate(parsed.data, {
      onSuccess: () => {
        toast.show('Perfil atualizado.', 'success');
        router.back();
      },
      onError: (err) => toast.show(authErrorMessage(err), 'error'),
    });
  }

  const avatarUrl = student?.avatarUrl ?? null;

  return (
    <ScreenContainer>
      <ScrollView
        contentContainerStyle={{ padding: space[5], flexGrow: 1 }}
        keyboardShouldPersistTaps="handled"
      >
        <Text
          style={{
            color: color.ink[1],
            fontFamily: font.display,
            fontSize: size['2xl'],
            marginBottom: space[4],
          }}
        >
          Editar perfil
        </Text>

        <View style={{ alignItems: 'center', marginBottom: space[5] }}>
          <Pressable
            accessibilityRole="button"
            accessibilityLabel="Trocar avatar"
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
            {avatarUrl ? (
              <Image
                source={{ uri: avatarUrl }}
                style={{ width: 120, height: 120 }}
                contentFit="cover"
              />
            ) : (
              <MascotPlaceholder variant="cheering" size={120} label="Avatar" />
            )}
          </Pressable>
          <Pressable onPress={pickAvatar} style={{ marginTop: space[2] }}>
            <Text
              style={{
                color: color.accent,
                fontFamily: font.heading,
                fontSize: size.sm,
              }}
            >
              {uploadAvatar.isPending ? 'Enviando…' : avatarUrl ? 'Trocar foto' : 'Adicionar foto'}
            </Text>
          </Pressable>
        </View>

        <FormField
          label="Nome"
          value={displayName}
          onChangeText={setDisplayName}
          autoCapitalize="words"
          autoComplete="name"
          error={error ?? undefined}
        />
        <View style={{ flexDirection: 'row', gap: space[3] }}>
          <View style={{ flex: 1 }}>
            <Button label="Cancelar" variant="secondary" onPress={() => router.back()} fullWidth />
          </View>
          <View style={{ flex: 1 }}>
            <Button label="Salvar" onPress={onSave} loading={update.isPending} fullWidth />
          </View>
        </View>
      </ScrollView>
    </ScreenContainer>
  );
}
