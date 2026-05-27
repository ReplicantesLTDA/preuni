import { render } from '@testing-library/react-native';
import { MascotPlaceholder } from '@/components/MascotPlaceholder';

describe('MascotPlaceholder', () => {
  it.each(['thinking', 'reading', 'cheering'] as const)('renders %s variant', (variant) => {
    const { getByLabelText } = render(<MascotPlaceholder variant={variant} label="Logo" />);
    expect(getByLabelText(`Logo — ${variant}`)).toBeTruthy();
  });
});
