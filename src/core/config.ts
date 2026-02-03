import * as fs from 'fs';
import * as path from 'path';

export interface CoolCodeConfig {
  llm: {
    model: string;
    temperature?: number;
    maxTokens?: number;
  };
  features?: {
    fileTreeMaxDepth?: number;
    scanCache?: boolean;
    allowDangerous?: boolean;
    confirmEdits?: boolean;
    maxContextTokens?: number;
  };
  guardrails?: {
    blockReadPatterns?: string[];
  };
}

export const DEFAULT_CONFIG: CoolCodeConfig = {
  llm: {
    model: 'gemini-2.5-flash',
  },
  features: {
    scanCache: true,
    allowDangerous: false,
    confirmEdits: false,
    maxContextTokens: 20000,
  },
  guardrails: {
    blockReadPatterns: [
      '.env',
      '.env.*',
      '*.pem',
      '*.key',
      'id_rsa',
      'id_ed25519',
      '.npmrc',
    ],
  },
};

export function getConfigPath(rootDir: string) {
  return path.join(rootDir, '.coolcode.json');
}

export function loadConfig(rootDir: string): CoolCodeConfig {
  const configPath = getConfigPath(rootDir);
  if (!fs.existsSync(configPath)) {
    return cloneConfig(DEFAULT_CONFIG);
  }
  try {
    const raw = fs.readFileSync(configPath, 'utf-8');
    const parsed = JSON.parse(raw);
    return mergeConfig(DEFAULT_CONFIG, parsed);
  } catch {
    return cloneConfig(DEFAULT_CONFIG);
  }
}

export function saveConfig(rootDir: string, config: CoolCodeConfig) {
  const configPath = getConfigPath(rootDir);
  fs.writeFileSync(configPath, JSON.stringify(config, null, 2));
}

export function getByPath(obj: any, pathStr: string) {
  const parts = pathStr.split('.').filter(Boolean);
  let current = obj;
  for (const part of parts) {
    if (current == null || typeof current !== 'object') {
      return undefined;
    }
    current = current[part];
  }
  return current;
}

export function setByPath(obj: any, pathStr: string, value: any) {
  const parts = pathStr.split('.').filter(Boolean);
  if (parts.length === 0) return;
  let current = obj;
  for (let i = 0; i < parts.length - 1; i++) {
    const part = parts[i];
    if (!current[part] || typeof current[part] !== 'object') {
      current[part] = {};
    }
    current = current[part];
  }
  current[parts[parts.length - 1]] = value;
}

export function parseConfigValue(raw: string) {
  const trimmed = raw.trim();
  if (trimmed === '') return '';
  if (
    trimmed === 'true' ||
    trimmed === 'false' ||
    trimmed === 'null' ||
    trimmed.startsWith('{') ||
    trimmed.startsWith('[') ||
    trimmed.startsWith('"') ||
    trimmed.startsWith("'")
  ) {
    try {
      return JSON.parse(trimmed.replace(/^'|'$/g, '"'));
    } catch {
      return raw;
    }
  }
  const asNumber = Number(trimmed);
  if (!Number.isNaN(asNumber) && Number.isFinite(asNumber)) {
    return asNumber;
  }
  return raw;
}

function mergeConfig(base: CoolCodeConfig, overrides: Partial<CoolCodeConfig> | any): CoolCodeConfig {
  const merged: CoolCodeConfig = {
    llm: {
      model: overrides?.llm?.model ?? base.llm.model,
      temperature:
        overrides?.llm?.temperature ?? base.llm.temperature ?? undefined,
      maxTokens: overrides?.llm?.maxTokens ?? base.llm.maxTokens ?? undefined,
    },
    features: {
      fileTreeMaxDepth:
        overrides?.features?.fileTreeMaxDepth ??
        base.features?.fileTreeMaxDepth,
      scanCache: overrides?.features?.scanCache ?? base.features?.scanCache,
      allowDangerous:
        overrides?.features?.allowDangerous ?? base.features?.allowDangerous,
      confirmEdits:
        overrides?.features?.confirmEdits ?? base.features?.confirmEdits,
      maxContextTokens:
        overrides?.features?.maxContextTokens ?? base.features?.maxContextTokens,
    },
    guardrails: {
      blockReadPatterns:
        overrides?.guardrails?.blockReadPatterns ??
        base.guardrails?.blockReadPatterns,
    },
  };
  return merged;
}

function cloneConfig(config: CoolCodeConfig): CoolCodeConfig {
  if (typeof structuredClone === 'function') {
    return structuredClone(config);
  }
  return JSON.parse(JSON.stringify(config));
}
