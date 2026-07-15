import type { AgentMode } from '../types';

export interface SlashCommand {
  name: string; // includes the leading slash, e.g. "/help"
  description: string;
  usage?: string; // e.g. "/pin <file>"
  example?: string; // e.g. "/pin src/index.ts"
}

// Single source of truth for slash commands, reused by the autocomplete
// dropdown, /help, and the dispatcher.
export const COMMANDS: SlashCommand[] = [
  { name: '/help', description: 'Show available commands' },
  {
    name: '/mode',
    description: 'Show or switch mode (plan | agent | ask)',
    usage: '/mode [plan|agent|ask]',
    example: '/mode agent',
  },
  {
    name: '/effort',
    description: 'Show or switch effort (low | high)',
    usage: '/effort [low|high]',
    example: '/effort high',
  },
  {
    name: '/pin',
    description: 'Pin a file into context',
    usage: '/pin <file>',
    example: '/pin src/index.ts',
  },
  {
    name: '/unpin',
    description: 'Unpin a file (or list pinned files)',
    usage: '/unpin [file]',
    example: '/unpin src/index.ts',
  },
  { name: '/context', description: 'Preview context, pinned files, and token usage' },
  { name: '/sessions', description: 'List saved sessions for this directory' },
  {
    name: '/install-skill',
    description: 'Install a skill from a local path or git URL',
    usage: '/install-skill <local-path|git-url> [--global]',
    example: '/install-skill https://github.com/owner/repo',
  },
  { name: '/clear', description: 'Clear the screen' },
  { name: '/exit', description: 'Exit the session' },
];

export const MODES: AgentMode[] = ['plan', 'agent', 'ask'];

// Returns the next mode in the plan -> agent -> ask -> plan cycle.
export function nextMode(mode: AgentMode): AgentMode {
  const i = MODES.indexOf(mode);
  return MODES[(i + 1) % MODES.length];
}

// True when every character of `needle` appears in `haystack` in order (a
// forgiving subsequence match), so "/insk" matches "/install-skill".
function fuzzyMatches(needle: string, haystack: string): boolean {
  let i = 0;
  for (const ch of haystack) {
    if (i < needle.length && needle[i] === ch) i++;
  }
  return i === needle.length;
}

// Commands matching the first whitespace-delimited token of the input, but only
// while the user is typing a command (the line starts with "/"). Matches by
// prefix first, then falls back to a fuzzy subsequence match.
export function matchCommands(input: string): SlashCommand[] {
  if (!input.startsWith('/')) return [];
  const token = input.split(/\s+/)[0].toLowerCase();
  if (token.length <= 1) return COMMANDS;
  const prefix = COMMANDS.filter((c) => c.name.startsWith(token));
  if (prefix.length > 0) return prefix;
  return COMMANDS.filter((c) => fuzzyMatches(token, c.name));
}
