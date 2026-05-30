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
    const plan = new Processor(process.cwd(), DEFAULT_CONFIG, {
      quiet: true,
      mode: 'plan',
    });
    expect(plan.getContextPreview().prompt).toContain('[PLAN MODE]');

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

describe('Processor plan-mode termination', () => {
  it('ends the turn after a task-list-only final response (no infinite loop)', async () => {
    const processor = new Processor(process.cwd(), DEFAULT_CONFIG, {
      quiet: true,
      mode: 'plan',
    });
    // A final response whose only tool call is the internal update_task_list.
    (processor as any).LLM.StreamResponse = async () =>
      JSON.stringify({
        text: 'Here is the plan.',
        tool_calls: [
          {
            tool: 'update_task_list',
            toolOptions: {
              goal: 'do the thing',
              items: [{ id: '1', title: 'step one', status: 'todo' }],
            },
          },
        ],
      });

    // Before the fix this would loop forever; the test timeout guards that.
    const result = await processor.processQuery('plan it');
    expect(result).toBe('Here is the plan.');
    expect(processor.getTaskList()?.goal).toBe('do the thing');
  }, 5000);
});
