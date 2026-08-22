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
});
