import React, { useState, useRef, useCallback, useEffect } from 'react';
import { Box, Text, Static, useInput, useApp, render } from 'ink';
import Spinner from 'ink-spinner';
import path from 'path';
import * as fs from 'fs';
import type { Processor } from '../core/processor';
import type { AgentMode, TaskList } from '../types';
import type { StatusReporter } from './spinner';
import { renderMarkdown } from './utils/markdown';
import { copyToClipboard } from './clipboard';
import { COMMANDS, matchCommands, nextMode } from './commands';
import { installSkill } from '../core/skillInstaller';
import {
  saveSession,
  listSessions,
  newSessionId,
  type SessionData,
} from '../core/session';

interface AppProps {
  processor: Processor;
  rootDir: string;
  copy: boolean;
  sessionId: string;
}

const modeColor: Record<AgentMode, string> = {
  plan: 'yellow',
  agent: 'green',
  ask: 'blue',
};

const PLAN_OPTIONS = [
  'Start implementation (switch to Agent and proceed)',
  'Keep planning (talk / refine the plan)',
];

let logCounter = 0;
interface LogLine {
  key: number;
  text: string;
}

export function App({ processor, rootDir, copy, sessionId }: AppProps) {
  const { exit } = useApp();
  const [log, setLog] = useState<LogLine[]>([]);
  const [status, setStatus] = useState<string | null>(null);
  const [input, setInput] = useState('');
  const [mode, setMode] = useState<AgentMode>(processor.getMode());
  const [processing, setProcessing] = useState(false);
  const [tasks, setTasks] = useState<TaskList | null>(processor.getTaskList());
  const [suggestIndex, setSuggestIndex] = useState(0);
  const confirmRef = useRef<{ resolve: (v: boolean) => void } | null>(null);
  const [confirmMsg, setConfirmMsg] = useState<string | null>(null);
  const [planMenu, setPlanMenu] = useState(false);
  const [planMenuIndex, setPlanMenuIndex] = useState(0);

  const append = useCallback((text: string) => {
    setLog((prev) => [...prev, { key: logCounter++, text }]);
  }, []);

  const suggestions = !processing && input.startsWith('/') ? matchCommands(input) : [];

  // StatusReporter bridge: the processor calls these during a turn.
  const reporter: StatusReporter = {
    start: (t) => setStatus(t || 'Working...'),
    updateText: (t) => setStatus(t),
    succeed: (t) => {
      if (t) append(renderMarkdown(t));
      setStatus(null);
    },
    fail: (t) => {
      if (t) append(renderMarkdown(t));
      setStatus(null);
    },
    stop: () => setStatus(null),
  };

  const persist = () => {
    const snap = processor.snapshot();
    const data: SessionData = {
      id: sessionId,
      cwd: rootDir,
      updatedAt: new Date().toISOString(),
      ...snap,
    };
    saveSession(data);
  };

  const confirm = useCallback((message: string): Promise<boolean> => {
    return new Promise((resolve) => {
      confirmRef.current = { resolve };
      setConfirmMsg(message);
    });
  }, []);

  // Wire processor confirmation callbacks to the Ink overlay (once).
  useEffect(() => {
    processor.setConfirmHandlers(
      (m) => confirm(m),
      (m, preview) => confirm(preview ? `${m}\n${preview}` : m)
    );
  }, [processor, confirm]);

  const runQuery = async (text: string) => {
    setProcessing(true);
    append(`${'❯'} ${text}`);
    try {
      const result = await processor.processQuery(text, reporter);
      setTasks(processor.getTaskList());
      persist();
      if (copy && result) {
        const copied = copyToClipboard(result);
        append(copied.success ? 'Copied to clipboard.' : `Copy failed: ${copied.error}`);
      }
      // After a plan-mode turn that produced a plan, offer next actions.
      if (processor.getMode() === 'plan' && result) {
        setPlanMenuIndex(0);
        setPlanMenu(true);
      }
    } catch (err) {
      append(`Error: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setStatus(null);
      setProcessing(false);
    }
  };

  const dispatchCommand = (raw: string) => {
    const parts = raw.trim().split(/\s+/);
    const name = parts[0].toLowerCase();
    const arg = parts.slice(1).find((p) => !p.startsWith('--'));

    if (name === '/exit' || name === '/quit') {
      exit();
      return;
    }
    if (name === '/clear') {
      setLog([]);
      return;
    }
    if (name === '/help') {
      append('Commands:');
      COMMANDS.forEach((c) => append(`  ${c.name}  -  ${c.description}`));
      append('  Shift+Tab cycles mode (plan -> agent -> ask)');
      return;
    }
    if (name === '/mode') {
      if (arg && (['plan', 'agent', 'ask'] as string[]).includes(arg)) {
        processor.setMode(arg as AgentMode);
        setMode(arg as AgentMode);
        append(`Mode switched to ${arg.toUpperCase()}`);
      } else {
        append(`Current mode: ${mode.toUpperCase()} (use /mode plan|agent|ask)`);
      }
      return;
    }
    if (name === '/pin') {
      if (!arg) return append('Usage: /pin <file>');
      const abs = path.resolve(rootDir, arg);
      if (fs.existsSync(abs)) {
        processor.pinFile(abs);
        append(`Pinned: ${arg}`);
      } else {
        append(`File not found: ${arg}`);
      }
      return;
    }
    if (name === '/unpin') {
      if (arg) {
        processor.unpinFile(path.resolve(rootDir, arg));
        append(`Unpinned: ${arg}`);
      } else {
        const pinned = processor.getPinnedFiles();
        append(pinned.length ? `Pinned: ${pinned.map((f) => path.relative(rootDir, f)).join(', ')}` : 'No files pinned.');
      }
      return;
    }
    if (name === '/context') {
      const { stats } = processor.getContextPreview();
      append(`Context: ${stats.messageCount} messages, ~${Math.round(stats.totalTokens / 100) / 10}k tokens, ${processor.getPinnedFiles().length} pinned`);
      return;
    }
    if (name === '/sessions') {
      const sessions = listSessions(rootDir);
      if (!sessions.length) return append('No saved sessions for this directory.');
      append('Saved sessions (newest first):');
      sessions.slice(0, 10).forEach((s) =>
        append(`  ${s.id.slice(0, 8)}  ${s.updatedAt}  ${s.conversations.length} msgs`)
      );
      return;
    }
    if (name === '/install-skill') {
      if (!arg) return append('Usage: /install-skill <local-path|git-url> [--global]');
      const global = parts.includes('--global');
      const result = installSkill(arg, { global }, rootDir);
      if (result.error) return append(`Install failed: ${result.error}`);
      if (!result.installed.length) return append('No skills found in the source.');
      processor.reloadSkills();
      append(`Installed: ${result.installed.join(', ')}`);
      return;
    }
    append(`Unknown command: ${name}. Type /help.`);
  };

  const submit = () => {
    const value = input.trim();
    setInput('');
    setSuggestIndex(0);
    if (!value) return;
    if (processing) {
      // Non-blocking queue: feed the running turn.
      processor.enqueueMessage(value);
      append(`(queued) ${value}`);
      return;
    }
    if (value.startsWith('/')) {
      dispatchCommand(value);
      return;
    }
    void runQuery(value);
  };

  const startImplementation = () => {
    setPlanMenu(false);
    processor.setMode('agent');
    setMode('agent');
    void runQuery('The plan above is approved. Proceed with implementing it now.');
  };

  useInput((char, key) => {
    // Post-plan action menu.
    if (planMenu) {
      if (key.upArrow) {
        setPlanMenuIndex((i) => (i + PLAN_OPTIONS.length - 1) % PLAN_OPTIONS.length);
        return;
      }
      if (key.downArrow) {
        setPlanMenuIndex((i) => (i + 1) % PLAN_OPTIONS.length);
        return;
      }
      if (char === '1') return startImplementation();
      if (char === '2' || key.escape) {
        setPlanMenu(false);
        return;
      }
      if (key.return) {
        if (planMenuIndex === 0) startImplementation();
        else setPlanMenu(false);
        return;
      }
      return;
    }

    // Confirmation overlay takes priority.
    if (confirmMsg) {
      if (char.toLowerCase() === 'y') {
        confirmRef.current?.resolve(true);
        confirmRef.current = null;
        setConfirmMsg(null);
      } else if (char.toLowerCase() === 'n' || key.escape || key.return) {
        confirmRef.current?.resolve(false);
        confirmRef.current = null;
        setConfirmMsg(null);
      }
      return;
    }

    if (key.ctrl && char === 'c') {
      exit();
      return;
    }
    // Shift+Tab cycles mode.
    if (key.tab && key.shift) {
      const next = nextMode(processor.getMode());
      processor.setMode(next);
      setMode(next);
      return;
    }
    // Tab accepts the highlighted command suggestion.
    if (key.tab) {
      if (suggestions.length) {
        setInput(suggestions[suggestIndex % suggestions.length].name + ' ');
        setSuggestIndex(0);
      }
      return;
    }
    if (key.upArrow) {
      if (suggestions.length) setSuggestIndex((i) => (i - 1 + suggestions.length) % suggestions.length);
      return;
    }
    if (key.downArrow) {
      if (suggestions.length) setSuggestIndex((i) => (i + 1) % suggestions.length);
      return;
    }
    if (key.return) {
      submit();
      return;
    }
    if (key.backspace || key.delete) {
      setInput((v) => v.slice(0, -1));
      return;
    }
    if (char && !key.ctrl && !key.meta) {
      setInput((v) => v + char);
    }
  });

  const status_ = processor.getStatus();

  return (
    <Box flexDirection="column">
      <Static items={log}>
        {(line) => <Text key={line.key}>{line.text}</Text>}
      </Static>

      {tasks && (
        <Box flexDirection="column" marginTop={1}>
          <Text bold>{`Goal: ${tasks.goal}`}</Text>
          {tasks.items.map((it) => (
            <Box key={it.id} flexDirection="column">
              <Text>
                {it.status === 'done' ? '[x] ' : it.status === 'in-progress' ? '[~] ' : it.status === 'failed' ? '[!] ' : '[ ] '}
                {it.title}
              </Text>
              {it.detail ? <Text color="gray">{`      ${it.detail}`}</Text> : null}
            </Box>
          ))}
        </Box>
      )}

      <Box marginTop={1}>
        <Text color="gray">
          {path.basename(rootDir)}
          {'  '}
          <Text color={modeColor[mode]}>{mode}</Text>
          {'  '}
          {status_.model}
          {'  '}
          {`${status_.messageCount} msgs`}
          {'  '}
          {`${Math.round(status_.totalTokens / 100) / 10}k tok`}
        </Text>
      </Box>

      {status ? (
        <Box>
          <Text color="cyan">
            <Spinner type="dots" />
          </Text>
          <Text>{' ' + status}</Text>
        </Box>
      ) : confirmMsg ? (
        <Box flexDirection="column">
          <Text color="yellow">{confirmMsg}</Text>
          <Text color="gray">Proceed? [y/N]</Text>
        </Box>
      ) : planMenu ? (
        <Box flexDirection="column">
          <Text bold color="yellow">Plan ready. What next?</Text>
          {PLAN_OPTIONS.map((label, i) => (
            <Text key={label} color={i === planMenuIndex ? 'cyan' : 'gray'}>
              {`${i === planMenuIndex ? '>' : ' '} ${i + 1}. ${label}`}
            </Text>
          ))}
          <Text color="gray">Up/Down + Enter, or press 1/2 - Esc to keep planning</Text>
        </Box>
      ) : (
        <Box flexDirection="column">
          <Box>
            <Text color="cyan">{'❯ '}</Text>
            <Text>{input}</Text>
            <Text color="gray">{'█'}</Text>
          </Box>
          {suggestions.length > 0 && (
            <Box flexDirection="column" marginLeft={2}>
              {suggestions.map((s, i) => (
                <Text key={s.name} color={i === suggestIndex % suggestions.length ? 'cyan' : 'gray'}>
                  {`${s.name}  ${s.description}`}
                </Text>
              ))}
            </Box>
          )}
        </Box>
      )}
    </Box>
  );
}

export async function runApp(props: AppProps): Promise<void> {
  const instance = render(<App {...props} />);
  await instance.waitUntilExit();
}

export { newSessionId };
