import { render, fireEvent } from '@testing-library/react-native';
import { NextActivityCard } from '@/features/trilha/components/NextActivityCard';

describe('NextActivityCard', () => {
  it('renders with the default mascot and no body', () => {
    const { getByText, queryByText } = render(
      <NextActivityCard title="Título" ctaLabel="Ir" onPress={() => undefined} />,
    );
    expect(getByText('Título')).toBeTruthy();
    expect(queryByText(/body/)).toBeNull();
  });

  it('renders body text and triggers onPress with an explicit mascot', () => {
    const onPress = jest.fn();
    const { getByText } = render(
      <NextActivityCard
        title="Título"
        body="Descrição"
        ctaLabel="Ir"
        onPress={onPress}
        mascot="reading"
      />,
    );
    expect(getByText('Descrição')).toBeTruthy();
    fireEvent.press(getByText('Ir'));
    expect(onPress).toHaveBeenCalled();
  });
});
