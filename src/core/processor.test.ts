import { describe, it, expect, beforeAll } from 'vitest';
import { Processor } from './processor';
import { DEFAULT_CONFIG } from './config';

beforeAll(() => {
  // The LLM constructor exits the process if this is missing; a dummy value is
  // enough since no network call happens during construction.
  process.env.GOOGLE_GENERATIVE_AI_API_KEY ||= 'test-key';
});

describe('Processor initial mode', () => {
  it('propagates the selected mode into the system prompt', () => {
    const planning = new Processor(process.cwd(), DEFAULT_CONFIG, {
      quiet: true,
      mode: 'planning',
    });
    expect(planning.getContextPreview().prompt).toContain('[PLANNING MODE]');

    const ask = new Processor(process.cwd(), DEFAULT_CONFIG, {
      quiet: true,
      mode: 'ask',
    });
    expect(ask.getContextPreview().prompt).toContain('[ASK MODE]');
  });

  it('defaults to agent mode', () => {
    const agent = new Processor(process.cwd(), DEFAULT_CONFIG, { quiet: true });
    expect(agent.getContextPreview().prompt).toContain('[AGENT MODE]');
  });
});
