import { Link, Stack } from 'expo-router';
import { View } from 'react-native';
import { EmptyState } from '@/components/EmptyState';

export default function NotFound() {
  return (
    <>
      <Stack.Screen options={{ title: 'Página não encontrada' }} />
      <View style={{ flex: 1 }}>
        <EmptyState
          title="Página não encontrada"
          body="O endereço que você acessou não existe."
          mascot="thinking"
        />
        <Link href="/" style={{ alignSelf: 'center', padding: 16 }}>
          Voltar para o início
        </Link>
      </View>
    </>
  );
}
