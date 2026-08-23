import { render, waitFor } from '@testing-library/react-native';
import { Text } from 'react-native';
import { renderHook } from '@testing-library/react-native';
import { ApiProvider, useApi } from '@/lib/api/context';
import { fetchMock } from '../../lib/mockFetch';
import { z } from 'zod';

function Probe() {
  const api = useApi();
  return <Text>{typeof api.request}</Text>;
}

function Requester() {
  const api = useApi();
  api.request({ method: 'GET', path: '/v1/ping', schema: z.object({}).passthrough() });
  return null;
}

describe('ApiProvider / useApi', () => {
  it('provides a client with a request function to descendants', () => {
    const { getByText } = render(
      <ApiProvider baseUrl="http://api.test">
        <Probe />
      </ApiProvider>,
    );
    expect(getByText('function')).toBeTruthy();
  });

  it('throws when useApi is called outside an ApiProvider', () => {
    const { result } = renderHook(() => {
      try {
        return useApi();
      } catch (e) {
        return e;
      }
    });
    expect(result.current).toBeInstanceOf(Error);
    expect((result.current as Error).message).toMatch(/useApi must be used within/);
  });

  it('logs the request in development mode', async () => {
    fetchMock.install();
    fetchMock.on('GET', '/v1/ping', { status: 200, body: {} });
    const originalEnv = process.env.NODE_ENV;
    // @ts-expect-error -- test-only override of a normally-readonly env var
    process.env.NODE_ENV = 'development';
    const logSpy = jest.spyOn(console, 'log').mockImplementation(() => {});

    try {
      render(
        <ApiProvider baseUrl="http://api.test">
          <Requester />
        </ApiProvider>,
      );
      await waitFor(() => expect(logSpy).toHaveBeenCalled());
      expect(logSpy.mock.calls[0]?.[0]).toMatch(/\[api\] GET \/v1\/ping/);
    } finally {
      // @ts-expect-error -- restoring the same test-only override
      process.env.NODE_ENV = originalEnv;
      logSpy.mockRestore();
      fetchMock.reset();
    }
  });
});
