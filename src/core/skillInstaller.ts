import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { spawnSync } from 'child_process';
import { parseSkillFile } from './skills';

export interface InstallSkillOptions {
  global?: boolean;
}

export interface InstallResult {
  installed: string[];
  dest: string;
  error?: string;
}

// Recognises safe git remote forms so we can clone instead of copy. Only
// https, scp-style git@, and ssh:// are accepted — this deliberately excludes
// transports git treats as command helpers (ext::, file://) and the bare
// ".git" suffix, which could otherwise smuggle an "ext::sh -c ..." payload
// into `git clone` and run arbitrary commands.
export function isGitUrl(source: string): boolean {
  return /^(https:\/\/|git@|ssh:\/\/)/.test(source);
}

function sanitizeName(name: string): string {
  return (
    name
      .replace(/[^a-zA-Z0-9_-]+/g, '-')
      .replace(/^-+|-+$/g, '')
      .toLowerCase() || 'skill'
  );
}

// Finds directories containing a SKILL.md: the root itself, its immediate
// subdirectories, and a conventional `skills/<name>` layout.
function collectSkillDirs(root: string): string[] {
  const found = new Set<string>();
  const check = (dir: string) => {
    if (fs.existsSync(path.join(dir, 'SKILL.md'))) found.add(dir);
  };
  check(root);
  for (const sub of [root, path.join(root, 'skills')]) {
    let entries: fs.Dirent[] = [];
    try {
      entries = fs.readdirSync(sub, { withFileTypes: true });
    } catch {
      continue;
    }
    for (const e of entries) {
      if (e.isDirectory()) check(path.join(sub, e.name));
    }
  }
  return Array.from(found);
}

function nameForSkill(skillMdPath: string): string {
  const dirName = path.basename(path.dirname(skillMdPath));
  try {
    return parseSkillFile(skillMdPath, dirName).name;
  } catch {
    return dirName;
  }
}

/**
 * Installs one or more skills from a local path or a git URL into the skills
 * directory (project `.coolcode/skills/` by default, or the global
 * `~/.coolcode/skills/` when `global` is set). A source may be a single
 * SKILL.md, a skill directory, or a repo/directory containing several skills.
 */
export function installSkill(
  source: string,
  options: InstallSkillOptions,
  rootDir: string
): InstallResult {
  const destBase = options.global
    ? path.join(os.homedir(), '.coolcode', 'skills')
    : path.join(rootDir, '.coolcode', 'skills');

  let tempDir: string | null = null;
  try {
    // Resolve the source into a working directory + the set of skill folders.
    const items: { src: string; name: string; fileOnly: boolean }[] = [];

    if (isGitUrl(source)) {
      tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'coolcode-skill-'));
      const res = spawnSync(
        'git',
        [
          '-c',
          'protocol.ext.allow=never',
          '-c',
          'protocol.file.allow=never',
          'clone',
          '--depth',
          '1',
          '--',
          source,
          tempDir,
        ],
        { stdio: 'pipe' }
      );
      if (res.status !== 0) {
        return {
          installed: [],
          dest: destBase,
          error: `git clone failed: ${res.stderr?.toString().trim() || 'unknown error'}`,
        };
      }
      for (const dir of collectSkillDirs(tempDir)) {
        items.push({ src: dir, name: nameForSkill(path.join(dir, 'SKILL.md')), fileOnly: false });
      }
    } else {
      const abs = path.resolve(source);
      if (!fs.existsSync(abs)) {
        return { installed: [], dest: destBase, error: `Source not found: ${abs}` };
      }
      const stat = fs.statSync(abs);
      if (stat.isFile()) {
        if (path.basename(abs) !== 'SKILL.md') {
          return {
            installed: [],
            dest: destBase,
            error: 'Expected a SKILL.md file or a directory containing one.',
          };
        }
        items.push({ src: abs, name: nameForSkill(abs), fileOnly: true });
      } else {
        for (const dir of collectSkillDirs(abs)) {
          items.push({ src: dir, name: nameForSkill(path.join(dir, 'SKILL.md')), fileOnly: false });
        }
      }
    }

    if (items.length === 0) {
      return {
        installed: [],
        dest: destBase,
        error: 'No SKILL.md found in the provided source.',
      };
    }

    fs.mkdirSync(destBase, { recursive: true });
    const installed: string[] = [];
    for (const item of items) {
      const destDir = path.join(destBase, sanitizeName(item.name));
      fs.rmSync(destDir, { recursive: true, force: true });
      if (item.fileOnly) {
        fs.mkdirSync(destDir, { recursive: true });
        fs.copyFileSync(item.src, path.join(destDir, 'SKILL.md'));
      } else {
        fs.cpSync(item.src, destDir, { recursive: true });
      }
      installed.push(item.name);
    }

    return { installed, dest: destBase };
  } catch (error) {
    return {
      installed: [],
      dest: destBase,
      error: error instanceof Error ? error.message : String(error),
    };
  } finally {
    if (tempDir) fs.rmSync(tempDir, { recursive: true, force: true });
  }
}
