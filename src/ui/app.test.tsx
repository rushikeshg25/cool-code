import { describe, it, expect } from 'vitest';
import React from 'react';
import { render } from 'ink-testing-library';
import { App } from './app';

// A minimal Processor stub covering only what App touches on mount/initial render.
function stubProcessor(mode = 'agent') {
  return {
    getMode: () => mode,
    getEffort: () => 'low',
    getTaskList: () => null,
    getStatus: () => ({ model: 'gemini-2.5-flash', messageCount: 0, totalTokens: 0, mode, effort: 'low' }),
    setConfirmHandlers: () => {},
  } as any;
}

describe('App (Ink)', () => {
  it('mounts and renders the header and prompt', () => {
    const { lastFrame, unmount } = render(
      <App processor={stubProcessor()} rootDir="/tmp/astro-paper" copy={false} sessionId="abcd1234" />
    );
    const frame = lastFrame() ?? '';
    expect(frame).toContain('astro-paper');
    expect(frame).toContain('agent');
    expect(frame).toContain('low effort');
    expect(frame).toContain('gemini-2.5-flash');
    expect(frame).toContain('›');
    unmount();
  });
});
