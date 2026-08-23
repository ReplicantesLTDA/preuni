import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

const DataExportScreen = require('../../../app/(tabs)/perfil/data-export').default;

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('DataExportScreen', () => {
  it('requests a data export', async () => {
    let called = false;
    fetchMock.on('GET', '/v1/students/me/data-export', {
      status: 200,
      body: {},
      capture: () => {
        called = true;
      },
    });
    const { Wrapper } = buildWrapper();
    const { getByText } = render(<DataExportScreen />, { wrapper: Wrapper });

    fireEvent.press(getByText('Solicitar exportação'));

    await waitFor(() => expect(called).toBe(true));
  });

  it('shows a toast when the export request fails', async () => {
    fetchMock.on('GET', '/v1/students/me/data-export', {
      status: 500,
      bodyText: '{"error":{"code":"INTERNAL_ERROR","message":"boom"}}',
    });
    const { Wrapper } = buildWrapper();
    const { getByText, findByText } = render(<DataExportScreen />, { wrapper: Wrapper });

    fireEvent.press(getByText('Solicitar exportação'));

    expect(await findByText('boom')).toBeTruthy();
  });
});
