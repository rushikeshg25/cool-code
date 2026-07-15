import { describe, it, expect } from 'vitest';
import { getByPath, setByPath } from './config';

describe('setByPath prototype-pollution guard', () => {
  it('rejects __proto__ and does not pollute Object.prototype', () => {
    const obj: any = {};
    expect(() => setByPath(obj, '__proto__.polluted', true)).toThrow(
      /Illegal config key/
    );
    expect(({} as any).polluted).toBeUndefined();
  });

  it('rejects constructor.prototype paths', () => {
    expect(() => setByPath({}, 'constructor.prototype.x', 1)).toThrow(
      /Illegal config key/
    );
    expect(({} as any).x).toBeUndefined();
  });

  it('still sets normal nested keys', () => {
    const obj: any = {};
    setByPath(obj, 'llm.model', 'gemini-2.5-flash');
    expect(obj.llm.model).toBe('gemini-2.5-flash');
  });
});

describe('getByPath prototype-pollution guard', () => {
  it('returns undefined for unsafe segments', () => {
    expect(getByPath({}, '__proto__.polluted')).toBeUndefined();
  });
});
