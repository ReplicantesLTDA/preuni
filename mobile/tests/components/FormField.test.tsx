import { render, fireEvent } from '@testing-library/react-native';
import { FormField } from '@/components/FormField';

describe('FormField', () => {
  it('renders label and error text', () => {
    const { getByText, getByLabelText } = render(
      <FormField label="E-mail" value="" onChangeText={() => undefined} error="inválido" />,
    );
    expect(getByText('E-mail')).toBeTruthy();
    expect(getByText('inválido')).toBeTruthy();
    expect(getByLabelText('E-mail')).toBeTruthy();
  });

  it('emits onChangeText', () => {
    const onChange = jest.fn();
    const { getByLabelText } = render(
      <FormField label="E-mail" value="" onChangeText={onChange} />,
    );
    fireEvent.changeText(getByLabelText('E-mail'), 'foo@bar.com');
    expect(onChange).toHaveBeenCalledWith('foo@bar.com');
  });

  it('shows helper text only when no error', () => {
    const { getByText, queryByText, rerender } = render(
      <FormField label="X" value="" onChangeText={() => undefined} helper="dica" />,
    );
    expect(getByText('dica')).toBeTruthy();
    rerender(
      <FormField label="X" value="" onChangeText={() => undefined} helper="dica" error="ruim" />,
    );
    expect(queryByText('dica')).toBeNull();
  });
});
