import { useState } from 'react';
import { ScrollView, Text, View } from 'react-native';
import { Button } from '@/components/Button';
import { Card } from '@/components/Card';
import { EmptyState } from '@/components/EmptyState';
import { FormField } from '@/components/FormField';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useToast } from '@/components/Toast';
import { useTheme } from '@/theme';
import { t } from '@/lib/i18n/pt-BR';
import { useFriends, useRemoveFriend, useSendFriendRequest } from '@/features/social/hooks';

export default function FriendsScreen() {
  const { color, space, font, size } = useTheme();
  const toast = useToast();
  const { data: friends, isLoading } = useFriends();
  const sendRequest = useSendFriendRequest();
  const removeFriend = useRemoveFriend();
  const [addresseeId, setAddresseeId] = useState('');

  function onAdd() {
    if (!addresseeId.trim()) return;
    sendRequest.mutate(addresseeId.trim(), {
      onSuccess: () => {
        toast.show(t.social.requestSent, 'success');
        setAddresseeId('');
      },
      onError: () => toast.show('Não foi possível enviar o pedido.', 'error'),
    });
  }

  return (
    <ScreenContainer>
      <ScrollView contentContainerStyle={{ padding: space[5], paddingBottom: space[8] }}>
        <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['2xl'], marginBottom: space[4] }}>
          {t.social.title}
        </Text>

        <Card>
          <FormField
            label={t.social.addFriend}
            value={addresseeId}
            onChangeText={setAddresseeId}
            placeholder={t.social.addFriendPlaceholder}
            autoCapitalize="none"
          />
          <Button label={t.social.addFriend} onPress={onAdd} loading={sendRequest.isPending} fullWidth />
        </Card>

        <View style={{ height: space[5] }} />

        {isLoading ? null : friends && friends.length > 0 ? (
          <View style={{ gap: space[3] }}>
            {friends.map((f) => (
              <Card key={f.userId}>
                <Text style={{ color: color.ink[1], fontFamily: font.heading, fontSize: size.md }}>
                  {f.displayName}
                </Text>
                <Text style={{ color: color.ink[2], fontFamily: font.body, fontSize: size.sm, marginTop: space[1] }}>
                  {t.trilha.streak(f.currentStreak)}
                </Text>
                <Text style={{ color: color.ink[2], fontFamily: font.body, fontSize: size.sm }}>
                  {f.latestOverallScore != null ? `${t.redacao.overallScore}: ${f.latestOverallScore}` : t.social.noGradeYet}
                </Text>
                <View style={{ height: space[3] }} />
                <Button
                  label={t.social.remove}
                  variant="secondary"
                  onPress={() => removeFriend.mutate(f.friendshipId)}
                  fullWidth
                />
              </Card>
            ))}
          </View>
        ) : (
          <EmptyState title={t.social.title} body={t.social.empty} mascot="reading" />
        )}
      </ScrollView>
    </ScreenContainer>
  );
}
