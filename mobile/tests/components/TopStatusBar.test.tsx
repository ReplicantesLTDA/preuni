import { render, fireEvent } from '@testing-library/react-native';
import { TopStatusBar } from '@/components/TopStatusBar';

describe('TopStatusBar', () => {
  it('shows streak and xp values', () => {
    const { getByLabelText } = render(<TopStatusBar streak={7} xp={1250} />);
    expect(getByLabelText('7 dias de ofensiva')).toBeTruthy();
    expect(getByLabelText('1250 XP')).toBeTruthy();
  });

  it('triggers avatar onPress', () => {
    const onPress = jest.fn();
    const { getByLabelText } = render(
      <TopStatusBar streak={1} xp={0} onAvatarPress={onPress} />,
    );
    fireEvent.press(getByLabelText('Abrir perfil'));
    expect(onPress).toHaveBeenCalled();
  });

  it('renders in placeholder mode without crashing', () => {
    const { getByLabelText } = render(<TopStatusBar streak={3} xp={100} placeholder />);
    expect(getByLabelText('3 dias de ofensiva')).toBeTruthy();
  });

  it('renders an avatar image when avatarUrl is provided', () => {
    const { queryByText } = render(
      <TopStatusBar streak={1} xp={0} avatarUrl="https://example.com/avatar.png" />,
    );
    // The fallback emoji only renders when there's no avatarUrl.
    expect(queryByText('👤')).toBeNull();
  });
});
