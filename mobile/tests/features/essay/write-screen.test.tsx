import { render, waitFor, fireEvent } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

const mockPush = jest.fn();
const mockReplace = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ push: mockPush, replace: mockReplace }),
  useLocalSearchParams: () => ({}),
}));

const WriteEssayScreen = require('../../../app/(tabs)/redacao/write').default;

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  mockPush.mockClear();
  mockReplace.mockClear();
});

describe('WriteEssayScreen', () => {
  it('shows validation errors when submitting an empty form', () => {
    const { Wrapper } = buildWrapper();
    const { getByText } = render(<WriteEssayScreen />, { wrapper: Wrapper });

    fireEvent.press(getByText('Enviar redação'));

    expect(getByText('Informe o título do tema.')).toBeTruthy();
  });

  it('submits and navigates to the essay status screen', async () => {
    fetchMock.on('POST', '/v1/essays', {
      status: 202,
      body: { id: '00000000-0000-4000-a000-000000000001', status: 'pending', submitted_at: '2026-08-22T12:00:00Z' },
    });
    const { Wrapper } = buildWrapper();
    const { getByLabelText, getByText } = render(<WriteEssayScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Título do tema'), 'Tema');
    fireEvent.changeText(getByLabelText('Contexto do tema'), 'Contexto');
    fireEvent.changeText(getByLabelText('Sua redação'), 'Texto da redação');
    fireEvent.press(getByText('Enviar redação'));

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/redacao/00000000-0000-4000-a000-000000000001'));
  });
});
