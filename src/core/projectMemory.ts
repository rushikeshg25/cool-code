import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';

export const PROJECT_MEMORY_FILE = 'COOLCODE.md';

/**
 * Loads persistent project instructions (like CLAUDE.md). The project-root file
 * takes precedence; a global `~/.coolcode/COOLCODE.md` is used as a fallback so
 * users can set machine-wide defaults. Returns null when neither exists.
 */
export function loadProjectInstructions(rootDir: string): string | null {
  const candidates = [
    path.join(rootDir, PROJECT_MEMORY_FILE),
    path.join(os.homedir(), '.coolcode', PROJECT_MEMORY_FILE),
  ];

  for (const candidate of candidates) {
    try {
      if (fs.existsSync(candidate)) {
        const content = fs.readFileSync(candidate, 'utf-8').trim();
        if (content) return content;
      }
    } catch {
      // Ignore unreadable memory files; they are optional.
    }
  }
  return null;
}
