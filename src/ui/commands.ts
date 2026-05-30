import type { AgentMode } from '../types';

export interface SlashCommand {
  name: string; // includes the leading slash, e.g. "/help"
  description: string;
}

// Single source of truth for slash commands, reused by the autocomplete
// dropdown, /help, and the dispatcher.
export const COMMANDS: SlashCommand[] = [
  { name: '/help', description: 'Show available commands' },
  { name: '/mode', description: 'Show or switch mode (plan | agent | ask)' },
  { name: '/pin', description: 'Pin a file into context (e.g. /pin src/index.ts)' },
  { name: '/unpin', description: 'Unpin a file (or list pinned files)' },
  { name: '/context', description: 'Preview context, pinned files, and token usage' },
  { name: '/sessions', description: 'List saved sessions for this directory' },
  { name: '/install-skill', description: 'Install a skill from a local path or git URL' },
  { name: '/clear', description: 'Clear the screen' },
  { name: '/exit', description: 'Exit the session' },
];

export const MODES: AgentMode[] = ['plan', 'agent', 'ask'];

// Returns the next mode in the plan -> agent -> ask -> plan cycle.
export function nextMode(mode: AgentMode): AgentMode {
  const i = MODES.indexOf(mode);
  return MODES[(i + 1) % MODES.length];
}

// Commands matching the first whitespace-delimited token of the input, but only
// while the user is typing a command (the line starts with "/").
export function matchCommands(input: string): SlashCommand[] {
  if (!input.startsWith('/')) return [];
  const token = input.split(/\s+/)[0];
  if (token.length <= 1) return COMMANDS;
  return COMMANDS.filter((c) => c.name.startsWith(token));
}
