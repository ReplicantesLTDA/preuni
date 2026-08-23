import { render, fireEvent } from '@testing-library/react-native';
import { ThemeProvider } from '@/theme';

const mockPush = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ push: mockPush }),
}));

const AuthWelcome = require('../../../app/(auth)/welcome').default;

function Wrapper({ children }: { children: React.ReactNode }) {
  return <ThemeProvider>{children}</ThemeProvider>;
}

describe('AuthWelcome', () => {
  beforeEach(() => mockPush.mockClear());

  it('navigates to login', () => {
    const { getByText } = render(<AuthWelcome />, { wrapper: Wrapper });
    fireEvent.press(getByText('Entrar'));
    expect(mockPush).toHaveBeenCalledWith('/(auth)/login');
  });

  it('navigates to register', () => {
    const { getByText } = render(<AuthWelcome />, { wrapper: Wrapper });
    fireEvent.press(getByText('Criar conta'));
    expect(mockPush).toHaveBeenCalledWith('/(auth)/register');
  });
});
