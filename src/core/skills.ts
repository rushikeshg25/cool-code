import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';

export interface Skill {
  name: string;
  description: string;
  body: string;
  path: string;
}

const SKILL_FILE = 'SKILL.md';

function skillDirs(rootDir: string): string[] {
  return [
    path.join(rootDir, '.coolcode', 'skills'),
    path.join(os.homedir(), '.coolcode', 'skills'),
  ];
}

// Parses a leading `--- ... ---` frontmatter block for `name:` / `description:`
// and returns it alongside the remaining markdown body. Falls back to the
// directory name and the first non-empty line when frontmatter is absent.
function parseSkillFile(filePath: string, dirName: string): Skill {
  const raw = fs.readFileSync(filePath, 'utf-8');
  let name = dirName;
  let description = '';
  let body = raw.trim();

  const fmMatch = raw.match(/^---\s*\n([\s\S]*?)\n---\s*\n?/);
  if (fmMatch) {
    const front = fmMatch[1];
    const nameMatch = front.match(/^\s*name:\s*(.+)\s*$/m);
    const descMatch = front.match(/^\s*description:\s*(.+)\s*$/m);
    if (nameMatch) name = nameMatch[1].trim();
    if (descMatch) description = descMatch[1].trim();
    body = raw.slice(fmMatch[0].length).trim();
  }

  if (!description) {
    description =
      body.split('\n').find((l) => l.trim().length > 0)?.trim() ?? '';
  }

  return { name, description, body, path: filePath };
}

/**
 * Discovers skills under `.coolcode/skills/<name>/SKILL.md` (project) and the
 * global `~/.coolcode/skills/<name>/SKILL.md`. Project skills take precedence on
 * name collisions.
 */
export function discoverSkills(rootDir: string): Skill[] {
  const byName = new Map<string, Skill>();

  for (const base of skillDirs(rootDir)) {
    let entries: fs.Dirent[];
    try {
      entries = fs.readdirSync(base, { withFileTypes: true });
    } catch {
      continue; // directory absent — fine
    }
    for (const entry of entries) {
      if (!entry.isDirectory()) continue;
      const file = path.join(base, entry.name, SKILL_FILE);
      if (!fs.existsSync(file)) continue;
      try {
        const skill = parseSkillFile(file, entry.name);
        if (!byName.has(skill.name)) byName.set(skill.name, skill);
      } catch {
        // Skip unreadable/malformed skill files.
      }
    }
  }

  return Array.from(byName.values());
}

/** A compact `name: description` catalog for the system prompt, or '' if none. */
export function buildSkillsCatalog(rootDir: string): string {
  const skills = discoverSkills(rootDir);
  if (skills.length === 0) return '';
  const lines = skills.map((s) => `- ${s.name}: ${s.description}`);
  return `--- Available Skills ---\nUse the 'use_skill' tool with a skill's name to load its full instructions when relevant.\n${lines.join(
    '\n'
  )}`;
}

/** Returns the full SKILL.md body for a named skill, or null if not found. */
export function getSkillBody(rootDir: string, name: string): string | null {
  const skill = discoverSkills(rootDir).find((s) => s.name === name);
  return skill ? skill.body : null;
}
