import { render, fireEvent } from '@testing-library/react-native';
import { NextActivityCard } from '@/features/trilha/components/NextActivityCard';
import { ReadinessGauge } from '@/features/trilha/components/ReadinessGauge';
import { SubjectGrid, DEFAULT_SUBJECTS } from '@/features/trilha/components/SubjectGrid';

describe('NextActivityCard', () => {
  it('renders title, body, CTA and fires onPress', () => {
    const onPress = jest.fn();
    const { getByText } = render(
      <NextActivityCard title="Próximo" body="vamos lá" ctaLabel="Começar" onPress={onPress} />,
    );
    expect(getByText('Próximo')).toBeTruthy();
    expect(getByText('vamos lá')).toBeTruthy();
    fireEvent.press(getByText('Começar'));
    expect(onPress).toHaveBeenCalledTimes(1);
  });
});

describe('ReadinessGauge', () => {
  it('clamps and renders score', () => {
    const { getByText } = render(<ReadinessGauge score={62.5} />);
    expect(getByText('63', { exact: false })).toBeTruthy();
  });

  it('clamps below 0', () => {
    const { getByText } = render(<ReadinessGauge score={-10} />);
    expect(getByText('0', { exact: false })).toBeTruthy();
  });

  it('clamps above 100', () => {
    const { getByText } = render(<ReadinessGauge score={150} />);
    expect(getByText('100', { exact: false })).toBeTruthy();
  });
});

describe('SubjectGrid', () => {
  it('renders 6 default subjects and reports selection', () => {
    const onSelect = jest.fn();
    const { getByLabelText } = render(<SubjectGrid onSelect={onSelect} />);
    DEFAULT_SUBJECTS.forEach((s) => expect(getByLabelText(s.label)).toBeTruthy());
    fireEvent.press(getByLabelText('Matemática'));
    expect(onSelect).toHaveBeenCalledWith(expect.objectContaining({ id: 'matematica' }));
  });
});
