import { render, fireEvent } from '@testing-library/react-native';
import { SubjectGrid, DEFAULT_SUBJECTS } from '@/features/trilha/components/SubjectGrid';

describe('SubjectGrid', () => {
  it('renders the default subject list', () => {
    const { getByText } = render(<SubjectGrid />);
    for (const s of DEFAULT_SUBJECTS) {
      expect(getByText(s.label)).toBeTruthy();
    }
  });

  it('does not crash when a tile is pressed without an onSelect handler', () => {
    const { getByText } = render(<SubjectGrid />);
    fireEvent.press(getByText(DEFAULT_SUBJECTS[0]!.label));
  });

  it('calls onSelect with the pressed subject when provided', () => {
    const onSelect = jest.fn();
    const { getByText } = render(<SubjectGrid onSelect={onSelect} />);
    fireEvent.press(getByText(DEFAULT_SUBJECTS[0]!.label));
    expect(onSelect).toHaveBeenCalledWith(DEFAULT_SUBJECTS[0]);
  });
});
