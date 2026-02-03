import * as fs from 'fs';
import * as path from 'path';

export interface ProjectScan {
  rootDir: string;
  timestamp: string;
  entrypoints: string[];
  frameworks: string[];
  scripts: string[];
  languages: string[];
  hasTsConfig: boolean;
  hasReadme: boolean;
  isNextJS?: boolean;
  isAstro?: boolean;
}

const COMMON_ENTRYPOINTS = [
  'src/index.ts',
  'src/index.js',
  'src/main.ts',
  'src/main.js',
  'src/app.ts',
  'src/app.js',
  'src/server.ts',
  'src/server.js',
  'index.ts',
  'index.js',
  'app.ts',
  'app.js',
  'server.ts',
  'server.js',
];

const FRAMEWORK_MAP: Record<string, string[]> = {
  NextJS: ['next'],
  React: ['react', 'react-dom'],
  Vue: ['vue'],
  Svelte: ['svelte'],
  Express: ['express'],
  Fastify: ['fastify'],
  NestJS: ['@nestjs/core'],
  Prisma: ['prisma', '@prisma/client'],
  Tailwind: ['tailwindcss'],
  Vite: ['vite'],
};

const EXTENSION_LANGUAGE_MAP: Record<string, string> = {
  '.ts': 'TypeScript',
  '.tsx': 'TypeScript (TSX)',
  '.js': 'JavaScript',
  '.jsx': 'JavaScript (JSX)',
  '.py': 'Python',
  '.go': 'Go',
  '.rs': 'Rust',
  '.java': 'Java',
  '.rb': 'Ruby',
  '.php': 'PHP',
  '.cs': 'C#',
  '.cpp': 'C++',
  '.c': 'C',
  '.h': 'C/C++ Header',
  '.md': 'Markdown',
  '.json': 'JSON',
  '.yml': 'YAML',
  '.yaml': 'YAML',
  '.sh': 'Shell',
  '.sql': 'SQL',
  '.html': 'HTML',
  '.css': 'CSS',
};

export function scanProject(rootDir: string): ProjectScan {
  const entrypoints = COMMON_ENTRYPOINTS.filter((rel) =>
    fs.existsSync(path.join(rootDir, rel))
  );

  const packageJsonPath = path.join(rootDir, 'package.json');
  let scripts: string[] = [];
  let frameworks: string[] = [];

  if (fs.existsSync(packageJsonPath)) {
    try {
      const pkg = JSON.parse(fs.readFileSync(packageJsonPath, 'utf-8'));
      scripts = Object.keys(pkg.scripts || {});
      const deps = Object.keys({
        ...(pkg.dependencies || {}),
        ...(pkg.devDependencies || {}),
      });
      frameworks = detectFrameworks(deps);
    } catch {
      // Ignore invalid package.json
    }
  }

  const languages = detectLanguages(rootDir);

  return {
    rootDir,
    timestamp: new Date().toISOString(),
    entrypoints,
    frameworks,
    scripts,
    languages,
    hasTsConfig: fs.existsSync(path.join(rootDir, 'tsconfig.json')),
    hasReadme: fs.existsSync(path.join(rootDir, 'README.md')),
    isNextJS:
      frameworks.includes('NextJS') ||
      fs.existsSync(path.join(rootDir, 'next.config.js')) ||
      fs.existsSync(path.join(rootDir, 'next.config.mjs')),
    isAstro:
      frameworks.includes('Astro') ||
      fs.existsSync(path.join(rootDir, 'astro.config.mjs')),
  };
}

function detectFrameworks(deps: string[]): string[] {
  const found = new Set<string>();
  for (const [framework, keywords] of Object.entries(FRAMEWORK_MAP)) {
    if (keywords.some((k) => deps.includes(k))) {
      found.add(framework);
    }
  }
  return Array.from(found).sort();
}

function detectLanguages(rootDir: string): string[] {
  const counts: Record<string, number> = {};
  walkFiles(rootDir, (filePath) => {
    const ext = path.extname(filePath);
    const lang = EXTENSION_LANGUAGE_MAP[ext];
    if (lang) {
      counts[lang] = (counts[lang] || 0) + 1;
    }
  });
  return Object.keys(counts).sort((a, b) => counts[b] - counts[a]);
}

function walkFiles(dir: string, onFile: (filePath: string) => void) {
  let entries: fs.Dirent[];
  try {
    entries = fs.readdirSync(dir, { withFileTypes: true });
  } catch {
    return;
  }
  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === 'node_modules' || entry.name === '.git') continue;
      walkFiles(fullPath, onFile);
    } else {
      onFile(fullPath);
    }
  }
}
