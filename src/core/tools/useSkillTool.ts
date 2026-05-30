import type { ToolResult } from '../../types';
import { discoverSkills, getSkillBody } from '../skills';

export interface UseSkillOptions {
  name: string;
}

export function useSkill(
  options: UseSkillOptions,
  rootPath: string
): ToolResult {
  if (!options.name || options.name.trim() === '') {
    return { DisplayResult: 'Fixing Issues', LLMresult: 'skill name is required.' };
  }

  const body = getSkillBody(rootPath, options.name);
  if (body === null) {
    const available = discoverSkills(rootPath)
      .map((s) => s.name)
      .join(', ');
    return {
      DisplayResult: 'Skill not found',
      LLMresult: `No skill named "${options.name}". Available skills: ${
        available || 'none'
      }.`,
    };
  }

  return {
    DisplayResult: `Loaded skill: ${options.name}`,
    LLMresult: `Instructions for skill "${options.name}":\n${body}`,
  };
}
