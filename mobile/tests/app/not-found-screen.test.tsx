import { render } from '@testing-library/react-native';
import { ThemeProvider } from '@/theme';

jest.mock('expo-router', () => {
  const { Text } = require('react-native');
  return {
    Stack: { Screen: () => null },
    Link: ({ href, children }: { href: string; children: React.ReactNode }) => (
      <Text accessibilityRole="link" accessibilityLabel={`link:${href}`}>
        {children}
      </Text>
    ),
  };
});

const NotFound = require('../../app/+not-found').default;

function Wrapper({ children }: { children: React.ReactNode }) {
  return <ThemeProvider>{children}</ThemeProvider>;
}

describe('NotFound', () => {
  it('shows the empty state and a link back home', () => {
    const { getByText, getByLabelText } = render(<NotFound />, { wrapper: Wrapper });

    expect(getByText('Página não encontrada')).toBeTruthy();
    expect(getByLabelText('link:/')).toBeTruthy();
  });
});
