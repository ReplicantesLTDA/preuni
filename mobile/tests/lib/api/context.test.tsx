import { render } from '@testing-library/react-native';
import { Text } from 'react-native';
import { renderHook } from '@testing-library/react-native';
import { ApiProvider, useApi } from '@/lib/api/context';

function Probe() {
  const api = useApi();
  return <Text>{typeof api.request}</Text>;
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
});
