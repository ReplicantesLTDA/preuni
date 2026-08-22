import { render, fireEvent } from '@testing-library/react-native';
import { EmptyState } from '@/components/EmptyState';

describe('EmptyState', () => {
  it('renders title + body and triggers CTA', () => {
    const onPress = jest.fn();
    const { getByText } = render(
      <EmptyState
        title="Vazio"
        body="Comece agora"
        cta={{ label: 'Começar', onPress }}
      />,
    );
    expect(getByText('Vazio')).toBeTruthy();
    expect(getByText('Comece agora')).toBeTruthy();
    fireEvent.press(getByText('Começar'));
    expect(onPress).toHaveBeenCalled();
  });
});
