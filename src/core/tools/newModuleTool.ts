import * as fs from 'fs';
import * as path from 'path';
import type { ToolResult } from '../../types';
import { toPascalCase } from './toolUtils';

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
  const moduleFile = path.join(moduleDir, `${options.moduleName}.ts`);
  const indexFile = path.join(moduleDir, 'index.ts');

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

  return {
    DisplayResult: 'Module created',
    LLMresult: `Created ${moduleDir}`,
  };
}
