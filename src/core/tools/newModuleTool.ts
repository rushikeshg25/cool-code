import * as fs from 'fs';
import * as path from 'path';
import type { ToolResult } from '../../types';
import { ensureAbsoluteWithinRoot, toPascalCase } from './toolUtils';
import { getErrorMessage } from '../utils';

export interface NewModuleOptions {
  moduleName: string;
  baseDir?: string;
  exportFromRootIndex?: boolean;
}

export function newModule(
  options: NewModuleOptions,
  rootPath: string
): ToolResult {
  if (!options.moduleName || options.moduleName.trim() === '') {
    return {
      DisplayResult: 'Fixing Issues',
      LLMresult: 'moduleName is required.',
    };
  }

  const baseDir = options.baseDir ?? 'src';
  const moduleDir = path.join(rootPath, baseDir, options.moduleName);

  // Reject baseDir/moduleName values (e.g. "../../etc") that escape the root.
  const containment = ensureAbsoluteWithinRoot(path.resolve(moduleDir), rootPath);
  if (containment) {
    return { DisplayResult: 'Fixing Issues', LLMresult: containment };
  }

  const moduleFile = path.join(moduleDir, `${options.moduleName}.ts`);
  const indexFile = path.join(moduleDir, 'index.ts');

  try {
    fs.mkdirSync(moduleDir, { recursive: true });

    const exportName = toPascalCase(options.moduleName) || 'NewModule';
    if (!fs.existsSync(moduleFile)) {
      fs.writeFileSync(
        moduleFile,
        `export const ${exportName} = () => {\n  // TODO: implement\n};\n`
      );
    }
    if (!fs.existsSync(indexFile)) {
      fs.writeFileSync(indexFile, `export * from './${options.moduleName}';\n`);
    }

    if (options.exportFromRootIndex) {
      const rootIndex = path.join(rootPath, baseDir, 'index.ts');
      const exportLine = `export * from './${options.moduleName}';\n`;
      if (fs.existsSync(rootIndex)) {
        const content = fs.readFileSync(rootIndex, 'utf-8');
        if (!content.includes(exportLine)) {
          fs.appendFileSync(rootIndex, exportLine);
        }
      } else {
        fs.writeFileSync(rootIndex, exportLine);
      }
    }
  } catch (error) {
    return {
      DisplayResult: 'Fixing Issues',
      LLMresult: `Failed to create module: ${getErrorMessage(error)}`,
    };
  }

  return {
    DisplayResult: 'Module created',
    LLMresult: `Created ${moduleDir}`,
  };
}
