import { parseApiErrorEnvelope } from '@/types/api';

describe('parseApiErrorEnvelope', () => {
  it('parses a valid envelope', () => {
    const env = parseApiErrorEnvelope(
      JSON.stringify({ error: { code: 'X', field: 'email', message: 'inválido' } }),
    );
    expect(env).toEqual({ error: { code: 'X', field: 'email', message: 'inválido' } });
  });

  it('returns null for unparseable JSON', () => {
    expect(parseApiErrorEnvelope('<html>not json</html>')).toBeNull();
  });

  it('returns null for JSON without error.message', () => {
    expect(parseApiErrorEnvelope(JSON.stringify({ foo: 'bar' }))).toBeNull();
  });
});
