import * as fs from 'fs';
import * as path from 'path';
import type { ToolResult } from '../../types';

export interface GenerateReadmeSectionOptions {
  title: string;
  bullets?: string[];
  content?: string;
}

export function generateReadmeSection(
  options: GenerateReadmeSectionOptions,
  rootPath: string
): ToolResult {
  if (!options.title || options.title.trim() === '') {
    return {
      DisplayResult: 'Fixing Issues',
      LLMresult: 'title is required.',
    };
  }

  const readmePath = path.join(rootPath, 'README.md');
  const heading = `## ${options.title}\n`;
  let body = '';

  if (options.content && options.content.trim() !== '') {
    body = `${options.content.trim()}\n`;
  } else if (options.bullets && options.bullets.length > 0) {
    body = options.bullets.map((b) => `- ${b}`).join('\n') + '\n';
  } else {
    body = '- TODO: add details\n';
  }

  const section = `\n${heading}\n${body}`;

  if (fs.existsSync(readmePath)) {
    const content = fs.readFileSync(readmePath, 'utf-8');
    if (content.includes(heading.trim())) {
      return {
        DisplayResult: 'Fixing Issues',
        LLMresult: `README already contains section "${options.title}".`,
      };
    }
    fs.appendFileSync(readmePath, section);
  } else {
    fs.writeFileSync(readmePath, `# ${path.basename(rootPath)}\n${section}`);
  }

  return {
    DisplayResult: 'README updated',
    LLMresult: `Added section "${options.title}" to README.md`,
  };
}
