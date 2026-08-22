import { render, fireEvent } from '@testing-library/react-native';
import { OtpInput } from '@/components/OtpInput';

describe('OtpInput', () => {
  it('emits digits stripped of non-digit chars', () => {
    const onChange = jest.fn();
    const { getByLabelText } = render(<OtpInput value="" onChange={onChange} />);
    fireEvent.changeText(getByLabelText('Código de verificação'), '1a2b3c4-5 6');
    expect(onChange).toHaveBeenCalledWith('123456');
  });

  it('caps input at length', () => {
    const onChange = jest.fn();
    const { getByLabelText } = render(<OtpInput value="" onChange={onChange} length={4} />);
    fireEvent.changeText(getByLabelText('Código de verificação'), '987654');
    expect(onChange).toHaveBeenCalledWith('9876');
  });
});
