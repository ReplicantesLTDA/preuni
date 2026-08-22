// Convert keys deeply. Skips arrays of primitives. Leaves non-plain-objects (Date, etc.) untouched.

function isPlainObject(v: unknown): v is Record<string, unknown> {
  if (v === null || typeof v !== 'object') return false;
  const proto = Object.getPrototypeOf(v);
  return proto === Object.prototype || proto === null;
}

const camelToSnake = (k: string): string => k.replace(/[A-Z]/g, (m) => `_${m.toLowerCase()}`);
const snakeToCamel = (k: string): string => k.replace(/_([a-z0-9])/g, (_, c: string) => c.toUpperCase());

export function toSnakeCase(input: unknown): unknown {
  if (Array.isArray(input)) return input.map(toSnakeCase);
  if (isPlainObject(input)) {
    const out: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(input)) {
      out[camelToSnake(k)] = toSnakeCase(v);
    }
    return out;
  }
  return input;
}

export function toCamelCase(input: unknown): unknown {
  if (Array.isArray(input)) return input.map(toCamelCase);
  if (isPlainObject(input)) {
    const out: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(input)) {
      out[snakeToCamel(k)] = toCamelCase(v);
    }
    return out;
  }
  return input;
}
