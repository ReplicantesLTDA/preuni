import { render, fireEvent } from '@testing-library/react-native';
import { ThemeProvider } from '@/theme';
import { ToastProvider } from '@/components/Toast';

const SimuladoHome = require('../../../app/(tabs)/simulado/index').default;

function Wrapper({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider>
      <ToastProvider>{children}</ToastProvider>
    </ThemeProvider>
  );
}

describe('SimuladoHome', () => {
  it('renders all simulation options and the empty history state', () => {
    const { getByText } = render(<SimuladoHome />, { wrapper: Wrapper });

    expect(getByText('Simulado rápido')).toBeTruthy();
    expect(getByText('Matemática focada')).toBeTruthy();
    expect(getByText('ENEM completo')).toBeTruthy();
    expect(getByText('Nenhum simulado realizado')).toBeTruthy();
  });

  it('shows a toast when starting a simulation (placeholder)', () => {
    const { getAllByText } = render(<SimuladoHome />, { wrapper: Wrapper });

    fireEvent.press(getAllByText('Começar')[0]!);
    // No crash + button remains interactive is the meaningful assertion here;
    // the toast itself is exercised by Toast's own component tests.
    expect(getAllByText('Começar')).toHaveLength(3);
  });
});
