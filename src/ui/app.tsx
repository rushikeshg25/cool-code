import React, { useState, useRef, useCallback, useEffect } from 'react';
import { Box, Text, Static, useInput, useApp, render } from 'ink';
import Spinner from 'ink-spinner';
import path from 'path';
import * as fs from 'fs';
import type { Processor } from '../core/processor';
import type { AgentMode, EffortLevel, TaskList } from '../types';
import type { StatusReporter } from './spinner';
import { renderMarkdown } from './utils/markdown';
import { copyToClipboard } from './clipboard';
import { COMMANDS, matchCommands, nextMode } from './commands';
import { c, glyph, modeColor, palette } from './theme';
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
  const [effort, setEffort] = useState<EffortLevel>(processor.getEffort());
  const [processing, setProcessing] = useState(false);
  const [tasks, setTasks] = useState<TaskList | null>(processor.getTaskList());
  const [suggestIndex, setSuggestIndex] = useState(0);
  const confirmRef = useRef<{ resolve: (v: boolean) => void } | null>(null);
  const [confirmMsg, setConfirmMsg] = useState<string | null>(null);
  const [planMenu, setPlanMenu] = useState(false);
  const [planMenuIndex, setPlanMenuIndex] = useState(0);
  const [elapsed, setElapsed] = useState(0);

  // Tick a seconds counter while a turn is running, for the spinner readout.
  useEffect(() => {
    if (!status) {
      setElapsed(0);
      return;
    }
    setElapsed(0);
    const started = Date.now();
    const id = setInterval(() => setElapsed(Math.floor((Date.now() - started) / 1000)), 1000);
    return () => clearInterval(id);
  }, [status]);

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
    append(`${c.accent(glyph.caret)} ${text}`);
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
      COMMANDS.forEach((cmd) => append(`  ${(cmd.usage || cmd.name).padEnd(34)}${cmd.description}`));
      append('  Shift+Tab                          cycle mode (plan / agent / ask)');
      append('  Ctrl+E                             toggle effort (low / high)');
      return;
    }
    if (name === '/mode') {
      if (arg && (['plan', 'agent', 'ask'] as string[]).includes(arg)) {
        processor.setMode(arg as AgentMode);
        setMode(arg as AgentMode);
        append(`Mode switched to ${arg}`);
      } else if (arg) {
        append('Usage: /mode [plan|agent|ask]');
      } else {
        append(`Current mode: ${mode} (use /mode plan|agent|ask)`);
      }
      return;
    }
    if (name === '/effort') {
      if (arg && (['low', 'high'] as string[]).includes(arg)) {
        processor.setEffort(arg as EffortLevel);
        setEffort(arg as EffortLevel);
        append(`Effort switched to ${arg}`);
      } else if (arg) {
        append('Usage: /effort [low|high]');
      } else {
        append(`Current effort: ${effort} (use /effort low|high)`);
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
      append(`${c.accent(glyph.caret)} mode: ${next}`);
      return;
    }
    // Ctrl+E toggles reasoning effort.
    if (key.ctrl && char === 'e') {
      const next: EffortLevel = processor.getEffort() === 'high' ? 'low' : 'high';
      processor.setEffort(next);
      setEffort(next);
      append(`${c.accent(glyph.caret)} effort: ${next}`);
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
          {tasks.items.map((it) => {
            const marker =
              it.status === 'done'
                ? { g: glyph.ok, color: palette.success }
                : it.status === 'in-progress'
                  ? { g: glyph.caret, color: palette.accent }
                  : it.status === 'failed'
                    ? { g: glyph.fail, color: palette.error }
                    : { g: glyph.bullet, color: palette.dim };
            return (
              <Box key={it.id} flexDirection="column">
                <Text>
                  <Text color={marker.color}>{`${marker.g} `}</Text>
                  {it.title}
                </Text>
                {it.detail ? <Text color={palette.dim}>{`      ${it.detail}`}</Text> : null}
              </Box>
            );
          })}
        </Box>
      )}

      <Box marginTop={1}>
        <Text color={palette.dim}>
          {path.basename(rootDir)}
          {` ${glyph.bullet} `}
          <Text color={modeColor[mode]}>{mode}</Text>
          {` ${glyph.bullet} `}
          {`${effort} effort`}
          {` ${glyph.bullet} `}
          {status_.model}
          {` ${glyph.bullet} `}
          {`${status_.messageCount} msgs`}
          {` ${glyph.bullet} `}
          {`${Math.round(status_.totalTokens / 100) / 10}k tok`}
        </Text>
      </Box>

      {status ? (
        <Box>
          <Text color={palette.accent}>
            <Spinner type="dots" />
          </Text>
          <Text color={palette.dim}>{` ${status}${elapsed > 0 ? ` (${elapsed}s)` : ''}`}</Text>
        </Box>
      ) : confirmMsg ? (
        <Box flexDirection="column">
          <Text color={palette.warn}>{confirmMsg}</Text>
          <Text color={palette.dim}>Proceed? [y/N]</Text>
        </Box>
      ) : planMenu ? (
        <Box flexDirection="column">
          <Text bold color={palette.warn}>Plan ready. What next?</Text>
          {PLAN_OPTIONS.map((label, i) => (
            <Text key={label} color={i === planMenuIndex ? palette.accent : palette.dim}>
              {`${i === planMenuIndex ? glyph.caret : ' '} ${i + 1}. ${label}`}
            </Text>
          ))}
          <Text color={palette.dim}>{`Up/Down + Enter, or press 1/2 ${glyph.bullet} Esc to keep planning`}</Text>
        </Box>
      ) : (
        <Box flexDirection="column">
          <Box borderStyle="round" borderColor={palette.dim} paddingX={1}>
            <Text color={palette.accent}>{`${glyph.caret} `}</Text>
            <Text>{input}</Text>
            <Text color={palette.dim}>{'█'}</Text>
          </Box>
          {suggestions.length > 0 && (
            <Box flexDirection="column" marginLeft={2} marginTop={1}>
              {suggestions.map((s, i) => {
                const active = i === suggestIndex % suggestions.length;
                const args = s.usage ? s.usage.slice(s.name.length).trim() : '';
                return (
                  <Text key={s.name}>
                    <Text color={active ? palette.accent : palette.dim}>
                      {active ? glyph.caret : ' '} {s.name}
                    </Text>
                    {args ? <Text color={palette.dim}>{` ${args}`}</Text> : null}
                    <Text color={palette.dim}>{`  ${s.description}`}</Text>
                  </Text>
                );
              })}
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
