import { describe, it, expect } from 'vitest';
import { inferProvider, resolveProvider } from './llm';

describe('inferProvider', () => {
  it('maps Gemini models to google', () => {
    expect(inferProvider('gemini-2.5-flash')).toBe('google');
  });

  it('maps GPT/o-series models to openai', () => {
    expect(inferProvider('gpt-4o')).toBe('openai');
    expect(inferProvider('o3-mini')).toBe('openai');
  });

  it('maps Claude models to anthropic', () => {
    expect(inferProvider('claude-sonnet-4-6')).toBe('anthropic');
  });
});

describe('resolveProvider', () => {
  it('prefers an explicit provider over inference', () => {
    expect(resolveProvider({ model: 'gpt-4o', provider: 'google' })).toBe('google');
  });

  it('falls back to inference when provider is unset', () => {
    expect(resolveProvider({ model: 'claude-3-5-sonnet' })).toBe('anthropic');
  });
});
