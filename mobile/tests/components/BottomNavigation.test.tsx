import { render, fireEvent } from '@testing-library/react-native';
import { BottomNavigation } from '@/components/BottomNavigation';

describe('BottomNavigation', () => {
  it('renders all four tabs', () => {
    const { getByLabelText } = render(<BottomNavigation current="trilha" onSelect={() => undefined} />);
    ['Trilha', 'Redação', 'Simulado', 'Perfil'].forEach((label) => {
      expect(getByLabelText(label)).toBeTruthy();
    });
  });

  it('triggers onSelect with the tab key', () => {
    const onSelect = jest.fn();
    const { getByLabelText } = render(<BottomNavigation current="trilha" onSelect={onSelect} />);
    fireEvent.press(getByLabelText('Perfil'));
    expect(onSelect).toHaveBeenCalledWith('perfil');
  });
});
