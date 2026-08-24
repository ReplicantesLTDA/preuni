import { render, fireEvent } from '@testing-library/react-native';
import { ThemeProvider } from '@/theme';
import { ToastProvider } from '@/components/Toast';
import { useSessionStore } from '@/stores/sessionStore';

const mockPush = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ push: mockPush }),
}));

const OnboardingProfileScreen = require('../../../app/(onboarding)/profile').default;

function Wrapper({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider>
      <ToastProvider>{children}</ToastProvider>
    </ThemeProvider>
  );
}

describe('OnboardingProfileScreen', () => {
  beforeEach(() => {
    mockPush.mockClear();
    useSessionStore.setState({ status: 'authed', student: null });
  });

  it('rejects a too-short display name', () => {
    const { getByLabelText, getByText } = render(<OnboardingProfileScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Como podemos te chamar?'), 'A');
    fireEvent.press(getByText('Continuar'));

    expect(getByText('Mínimo de 2 caracteres.')).toBeTruthy();
    expect(mockPush).not.toHaveBeenCalled();
  });

  it('rejects a too-long display name', () => {
    const { getByLabelText, getByText } = render(<OnboardingProfileScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Como podemos te chamar?'), 'A'.repeat(65));
    fireEvent.press(getByText('Continuar'));

    expect(getByText('Máximo de 64 caracteres.')).toBeTruthy();
    expect(mockPush).not.toHaveBeenCalled();
  });

  it('patches the session store and navigates to interests', () => {
    const { getByLabelText, getByText } = render(<OnboardingProfileScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Como podemos te chamar?'), 'Maria');
    fireEvent.press(getByText('Continuar'));

    expect(mockPush).toHaveBeenCalledWith('/(onboarding)/interests');
  });
});
