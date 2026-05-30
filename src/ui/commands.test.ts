import { describe, it, expect } from 'vitest';
import { matchCommands, nextMode, COMMANDS } from './commands';

describe('matchCommands', () => {
  it('returns nothing for non-command input', () => {
    expect(matchCommands('hello world')).toEqual([]);
  });

  it('returns all commands for a lone slash', () => {
    expect(matchCommands('/')).toEqual(COMMANDS);
  });

  it('filters by the typed prefix', () => {
    const names = matchCommands('/p').map((c) => c.name);
    expect(names).toContain('/pin');
    expect(names).not.toContain('/help');
  });

  it('matches a full command name', () => {
    expect(matchCommands('/install-skill ./x').map((c) => c.name)).toEqual([
      '/install-skill',
    ]);
  });
});

describe('nextMode', () => {
  it('cycles plan -> agent -> ask -> plan', () => {
    expect(nextMode('plan')).toBe('agent');
    expect(nextMode('agent')).toBe('ask');
    expect(nextMode('ask')).toBe('plan');
  });
});
