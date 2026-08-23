import { render, fireEvent, renderHook, act } from '@testing-library/react-native';
import { Text } from 'react-native';
import { Card } from '@/components/Card';
import { ScreenContainer } from '@/components/ScreenContainer';
import { Skeleton } from '@/components/Skeleton';
import { ErrorState } from '@/components/ErrorState';
import { ToastProvider, useToast } from '@/components/Toast';

describe('Card', () => {
  it('renders with default padding', () => {
    const { getByText } = render(
      <Card>
        <Text>content</Text>
      </Card>,
    );
    expect(getByText('content')).toBeTruthy();
  });

  it('renders with padded=false', () => {
    const { getByText } = render(
      <Card padded={false}>
        <Text>content</Text>
      </Card>,
    );
    expect(getByText('content')).toBeTruthy();
  });
});

describe('ScreenContainer', () => {
  it('renders with default keyboard-avoiding behavior', () => {
    const { getByText } = render(
      <ScreenContainer>
        <Text>body</Text>
      </ScreenContainer>,
    );
    expect(getByText('body')).toBeTruthy();
  });

  it('renders without keyboard avoidance when avoidKeyboard=false', () => {
    const { getByText } = render(
      <ScreenContainer avoidKeyboard={false}>
        <Text>body</Text>
      </ScreenContainer>,
    );
    expect(getByText('body')).toBeTruthy();
  });
});

describe('Skeleton', () => {
  it('renders with default height/width', () => {
    const { toJSON } = render(<Skeleton />);
    expect(toJSON()).toBeTruthy();
  });

  it('renders with an explicit borderRadius', () => {
    const { toJSON } = render(<Skeleton borderRadius={20} height={40} width={80} />);
    expect(toJSON()).toBeTruthy();
  });
});

describe('ErrorState', () => {
  it('renders with default title/body and no retry button', () => {
    const { getByText, queryByText } = render(<ErrorState />);
    expect(getByText('Algo deu errado')).toBeTruthy();
    expect(queryByText('Tentar novamente')).toBeNull();
  });

  it('renders a retry button when onRetry is provided', () => {
    const onRetry = jest.fn();
    const { getByText } = render(<ErrorState title="Falhou" body="detalhe" onRetry={onRetry} />);
    fireEvent.press(getByText('Tentar novamente'));
    expect(onRetry).toHaveBeenCalled();
  });
});

describe('Toast', () => {
  it('shows a toast with the default info variant and it auto-dismisses', () => {
    jest.useFakeTimers();
    function Trigger() {
      const toast = useToast();
      return <Text onPress={() => toast.show('oi')}>trigger</Text>;
    }
    const { getByText, queryByText } = render(
      <ToastProvider>
        <Trigger />
      </ToastProvider>,
    );
    fireEvent.press(getByText('trigger'));
    expect(getByText('oi')).toBeTruthy();
    act(() => {
      jest.advanceTimersByTime(4001);
    });
    expect(queryByText('oi')).toBeNull();
    jest.useRealTimers();
  });

  it('throws when useToast is called outside a ToastProvider', () => {
    const { result } = renderHook(() => {
      try {
        return useToast();
      } catch (e) {
        return e;
      }
    });
    expect(result.current).toBeInstanceOf(Error);
  });
});
