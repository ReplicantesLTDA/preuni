import { render } from '@testing-library/react-native';
import { useSessionStore } from '@/stores/sessionStore';

jest.mock('expo-router', () => {
  const { Text } = require('react-native');
  return {
    Redirect: ({ href }: { href: string }) => <Text>{`redirect:${href}`}</Text>,
  };
});

const Index = require('../../app/index').default;

describe('Index', () => {
  afterEach(() => {
    useSessionStore.setState({ status: 'loading', student: null });
  });

  it('redirects to auth welcome when anonymous', () => {
    useSessionStore.setState({ status: 'anon', student: null });
    const { getByText } = render(<Index />);
    expect(getByText('redirect:/(auth)/welcome')).toBeTruthy();
  });

  it('redirects to onboarding welcome when authed but onboarding incomplete', () => {
    useSessionStore.setState({
      status: 'authed',
      student: {
        id: '1',
        displayName: 'Maria',
        email: 'm@a.com',
        xpTotal: 0,
        streakCount: 0,
        readinessScore: 0,
        onboardingCompleted: false,
      },
    });
    const { getByText } = render(<Index />);
    expect(getByText('redirect:/(onboarding)/welcome')).toBeTruthy();
  });

  it('redirects to the trilha tab when authed and onboarded', () => {
    useSessionStore.setState({
      status: 'authed',
      student: {
        id: '1',
        displayName: 'Maria',
        email: 'm@a.com',
        xpTotal: 0,
        streakCount: 0,
        readinessScore: 0,
        onboardingCompleted: true,
      },
    });
    const { getByText } = render(<Index />);
    expect(getByText('redirect:/(tabs)/trilha')).toBeTruthy();
  });
});
