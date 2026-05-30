import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { discoverSkills, buildSkillsCatalog, getSkillBody } from './skills';
import { useSkill } from './tools/useSkillTool';

let dir: string;

function writeSkill(name: string, content: string) {
  const skillDir = path.join(dir, '.coolcode', 'skills', name);
  fs.mkdirSync(skillDir, { recursive: true });
  fs.writeFileSync(path.join(skillDir, 'SKILL.md'), content);
}

beforeEach(() => {
  dir = fs.mkdtempSync(path.join(os.tmpdir(), 'skills-test-'));
});

afterEach(() => {
  fs.rmSync(dir, { recursive: true, force: true });
});

describe('skills discovery', () => {
  it('parses frontmatter name/description and body', () => {
    writeSkill(
      'pdf',
      '---\nname: pdf-filler\ndescription: Fill PDF forms\n---\nStep 1: open the form.\nStep 2: fill it.'
    );
    const skills = discoverSkills(dir);
    expect(skills).toHaveLength(1);
    expect(skills[0].name).toBe('pdf-filler');
    expect(skills[0].description).toBe('Fill PDF forms');
    expect(skills[0].body).toContain('Step 1: open the form.');
  });

  it('builds a catalog of name: description lines', () => {
    writeSkill('a', '---\nname: alpha\ndescription: does alpha\n---\nbody');
    const catalog = buildSkillsCatalog(dir);
    expect(catalog).toContain('Available Skills');
    expect(catalog).toContain('- alpha: does alpha');
  });

  it('returns empty catalog when no skills exist', () => {
    expect(buildSkillsCatalog(dir)).toBe('');
  });
});

describe('use_skill tool', () => {
  it('returns the skill body for a known skill', () => {
    writeSkill('x', '---\nname: xtool\ndescription: d\n---\nDO THE THING');
    const result = useSkill({ name: 'xtool' }, dir);
    expect(result.LLMresult).toContain('DO THE THING');
  });

  it('returns a clean error for an unknown skill', () => {
    const result = useSkill({ name: 'nope' }, dir);
    expect(result.DisplayResult).toBe('Skill not found');
    expect(result.LLMresult).toMatch(/No skill named/);
  });
});
