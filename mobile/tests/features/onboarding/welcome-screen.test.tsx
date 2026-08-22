import { render, fireEvent } from '@testing-library/react-native';
import { ThemeProvider } from '@/theme';

const mockPush = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ push: mockPush }),
}));

const OnboardingWelcome = require('../../../app/(onboarding)/welcome').default;

function Wrapper({ children }: { children: React.ReactNode }) {
  return <ThemeProvider>{children}</ThemeProvider>;
}

describe('OnboardingWelcome', () => {
  it('navigates to profile on continue', () => {
    const { getByText } = render(<OnboardingWelcome />, { wrapper: Wrapper });

    fireEvent.press(getByText('Continuar'));

    expect(mockPush).toHaveBeenCalledWith('/(onboarding)/profile');
  });
});
