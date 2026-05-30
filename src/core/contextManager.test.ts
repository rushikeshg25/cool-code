import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { ContextManager } from './contextManager';

let dir: string;
const noIgnore = () => false;

beforeEach(() => {
  dir = fs.mkdtempSync(path.join(os.tmpdir(), 'ctx-test-'));
});

afterEach(() => {
  fs.rmSync(dir, { recursive: true, force: true });
});

describe('ContextManager pinned files', () => {
  it('does not leak guardrail-blocked files (.env) into the prompt', () => {
    const secret = path.join(dir, '.env');
    fs.writeFileSync(secret, 'API_SECRET=supersecretvalue');

    const cm = new ContextManager(dir, noIgnore);
    cm.pinFile(secret);

    const prompt = cm.buildPrompt();
    expect(prompt).not.toContain('supersecretvalue');
    expect(prompt).toContain('blocked by guardrails');
  });

  it('includes the contents of a normal pinned file', () => {
    const file = path.join(dir, 'notes.txt');
    fs.writeFileSync(file, 'hello-from-notes');

    const cm = new ContextManager(dir, noIgnore);
    cm.pinFile(file);

    expect(cm.buildPrompt()).toContain('hello-from-notes');
  });
});

describe('ContextManager prompt structure', () => {
  it('places the static tool-info section before the dynamic project state', () => {
    const cm = new ContextManager(dir, noIgnore);
    const prompt = cm.buildPrompt();
    expect(prompt.indexOf('These are your Tools')).toBeLessThan(
      prompt.indexOf('--- Recent Conversation ---')
    );
  });
});

describe('ContextManager summarization is reductive', () => {
  it('trims old messages and keeps the summary', () => {
    const cm = new ContextManager(dir, noIgnore);
    for (let i = 0; i < 30; i++) cm.addUserMessage(`message number ${i}`);
    expect(cm.getStats().messageCount).toBe(30);

    cm.applySummary('SUMMARY-TEXT', 6);
    expect(cm.getStats().messageCount).toBe(6);

    const prompt = cm.buildPrompt();
    expect(prompt).toContain('SUMMARY-TEXT');
    expect(prompt).toContain('message number 29');
    expect(prompt).not.toContain('message number 0');
  });
});
