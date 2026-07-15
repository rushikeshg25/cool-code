import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { installSkill, isGitUrl } from './skillInstaller';
import { discoverSkills } from './skills';

let work: string;

beforeEach(() => {
  work = fs.mkdtempSync(path.join(os.tmpdir(), 'skill-install-'));
});

afterEach(() => {
  fs.rmSync(work, { recursive: true, force: true });
});

describe('isGitUrl', () => {
  it('detects git URLs and leaves local paths alone', () => {
    expect(isGitUrl('https://github.com/x/y.git')).toBe(true);
    expect(isGitUrl('git@github.com:x/y.git')).toBe(true);
    expect(isGitUrl('ssh://git@github.com/x/y.git')).toBe(true);
    expect(isGitUrl('./my-skill')).toBe(false);
    expect(isGitUrl('/abs/path/SKILL.md')).toBe(false);
  });

  it('rejects transports git treats as command helpers', () => {
    // These would run arbitrary commands via git if passed to `git clone`.
    expect(isGitUrl('ext::sh -c "id"')).toBe(false);
    expect(isGitUrl('ext::sh -c id.git')).toBe(false);
    expect(isGitUrl('file:///etc/passwd')).toBe(false);
    expect(isGitUrl('git://evil.example/repo')).toBe(false);
    expect(isGitUrl('http://insecure.example/x.git')).toBe(false);
  });
});

describe('installSkill (local)', () => {
  it('installs a skill directory and makes it discoverable', () => {
    const src = path.join(work, 'src-skill');
    fs.mkdirSync(src, { recursive: true });
    fs.writeFileSync(
      path.join(src, 'SKILL.md'),
      '---\nname: my-cc-skill\ndescription: from claude code\n---\nDo the thing.'
    );
    fs.writeFileSync(path.join(src, 'helper.py'), 'print(1)');

    const project = path.join(work, 'project');
    fs.mkdirSync(project, { recursive: true });

    const result = installSkill(src, {}, project);
    expect(result.error).toBeUndefined();
    expect(result.installed).toContain('my-cc-skill');

    // Supporting files are copied, and the skill is discoverable.
    expect(fs.existsSync(path.join(project, '.coolcode', 'skills', 'my-cc-skill', 'helper.py'))).toBe(true);
    const skills = discoverSkills(project);
    expect(skills.map((s) => s.name)).toContain('my-cc-skill');
  });

  it('installs from a single SKILL.md file path', () => {
    const file = path.join(work, 'SKILL.md');
    fs.writeFileSync(file, '---\nname: solo\ndescription: d\n---\nbody');
    const project = path.join(work, 'p2');
    fs.mkdirSync(project, { recursive: true });

    const result = installSkill(file, {}, project);
    expect(result.installed).toEqual(['solo']);
    expect(fs.existsSync(path.join(project, '.coolcode', 'skills', 'solo', 'SKILL.md'))).toBe(true);
  });

  it('installs multiple skills from a skills/ directory', () => {
    const repo = path.join(work, 'repo', 'skills');
    for (const n of ['alpha', 'beta']) {
      fs.mkdirSync(path.join(repo, n), { recursive: true });
      fs.writeFileSync(
        path.join(repo, n, 'SKILL.md'),
        `---\nname: ${n}\ndescription: ${n} skill\n---\nbody`
      );
    }
    const project = path.join(work, 'p3');
    fs.mkdirSync(project, { recursive: true });

    const result = installSkill(path.join(work, 'repo'), {}, project);
    expect(result.installed.sort()).toEqual(['alpha', 'beta']);
  });

  it('returns a clean error for a missing source', () => {
    const result = installSkill(path.join(work, 'nope'), {}, work);
    expect(result.error).toMatch(/not found/);
    expect(result.installed).toEqual([]);
  });
});
