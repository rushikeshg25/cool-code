import chalk from 'chalk';
import type { AgentMode } from '../types';

// Single-accent visual system for the CLI. Everything structural is dim gray;
// the accent (cyan) draws the eye to the one thing that matters on a line.
// Semantic colors are reserved for success / warning / error signals only.
// Ink takes color-name strings, the console world uses the chalk helpers —
// keep both here so colors and glyphs are never re-scattered across the UI.

export const palette = {
  accent: 'cyan',
  dim: 'gray',
  success: 'green',
  warn: 'yellow',
  error: 'red',
} as const;

export const c = {
  accent: chalk.cyan,
  dim: chalk.gray,
  success: chalk.green,
  warn: chalk.yellow,
  error: chalk.red,
  bold: chalk.bold,
};

// Minimal glyphs — used where they carry meaning, never as decoration.
export const glyph = {
  caret: '›',
  ok: '✓',
  fail: '✗',
  bullet: '·',
};

// Mode accent (Ink color names). Muted but distinguishable per mode.
export const modeColor: Record<AgentMode, string> = {
  plan: 'yellow',
  agent: 'green',
  ask: 'blue',
};
