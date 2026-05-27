import { render, fireEvent } from '@testing-library/react-native';
import { Button } from '@/components/Button';

describe('Button', () => {
  it('renders the label and triggers onPress', () => {
    const onPress = jest.fn();
    const { getByText } = render(<Button label="Entrar" onPress={onPress} />);
    fireEvent.press(getByText('Entrar'));
    expect(onPress).toHaveBeenCalledTimes(1);
  });

  it('does not call onPress when disabled', () => {
    const onPress = jest.fn();
    const { getByRole } = render(<Button label="x" onPress={onPress} disabled />);
    fireEvent.press(getByRole('button'));
    expect(onPress).not.toHaveBeenCalled();
  });

  it('shows a spinner and blocks onPress while loading', () => {
    const onPress = jest.fn();
    const { queryByText, getByRole } = render(<Button label="x" onPress={onPress} loading />);
    expect(queryByText('x')).toBeNull();
    fireEvent.press(getByRole('button'));
    expect(onPress).not.toHaveBeenCalled();
  });

  it.each(['primary', 'secondary', 'ghost', 'danger'] as const)(
    'renders the %s variant',
    (variant) => {
      const { getByText } = render(<Button label="v" onPress={() => undefined} variant={variant} />);
      expect(getByText('v')).toBeTruthy();
    },
  );

  it('uses accessibilityLabel when provided', () => {
    const { getByLabelText } = render(
      <Button label="x" accessibilityLabel="continuar" onPress={() => undefined} />,
    );
    expect(getByLabelText('continuar')).toBeTruthy();
  });
});
