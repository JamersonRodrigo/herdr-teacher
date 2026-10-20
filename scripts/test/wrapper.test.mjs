// Tests for scripts/herdr-teacher.mjs: config layering, env export and the
// segregation guards. Run: node --test scripts/test/*.test.mjs
import { test } from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { normalizeKey, parseConfig, engineEnv, guard, loadTeacherConfig } from '../herdr-teacher.mjs';

const WRAPPER = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', 'herdr-teacher.mjs');

function tmpDir() {
  return fs.mkdtempSync(path.join(os.tmpdir(), 'herdr-teacher-test-'));
}

test('normalizeKey matches herdr-soho', () => {
  assert.equal(normalizeKey('lane.pesquisa.roles'), 'lane_pesquisa_roles');
  assert.equal(normalizeKey('role.content-creator.kind'), 'role_content_creator_kind');
  assert.equal(normalizeKey(' teacher.material_dir '), 'teacher_material_dir');
});

test('parseConfig skips comments, blanks and lines without a key', () => {
  const got = parseConfig('# c\n\na=1\r\n=x\nb = two words \nnoeq\n');
  assert.deepEqual(got, [{ key: 'a', value: '1' }, { key: 'b', value: 'two words' }]);
});

test('engineEnv exports keys, keeps explicit env and never exports teacher.*', () => {
  const entries = new Map([
    ['lanes', { key: 'lanes', value: 'on' }],
    ['family_check', { key: 'family_check', value: 'strict' }],
    ['state_dir', { key: 'state_dir', value: '' }],
    ['teacher_material_dir', { key: 'teacher.material_dir', value: 'aulas' }],
  ]);
  const { env, exported } = engineEnv(entries, { HERDR_SOHO_FAMILY_CHECK: 'warn' });
  assert.equal(env.HERDR_SOHO_LANES, 'on');
  assert.equal(env.HERDR_SOHO_FAMILY_CHECK, 'warn');
  assert.equal(env.HERDR_SOHO_STATE_DIR, undefined);
  assert.equal(env.HERDR_SOHO_TEACHER_MATERIAL_DIR, undefined);
  assert.deepEqual(exported, ['HERDR_SOHO_LANES']);
  assert.ok(fs.existsSync(path.join(env.HERDR_SOHO_ROLES, 'reviewer.md')));
});

test('project config wins over the skill defaults', () => {
  const root = tmpDir();
  fs.mkdirSync(path.join(root, '.agents'));
  fs.writeFileSync(path.join(root, '.agents', 'herdr-teacher.conf'), 'lane.revisao.kind=codex\n');
  const entries = loadTeacherConfig({ XDG_CONFIG_HOME: path.join(root, 'nouser') }, root);
  assert.equal(entries.get('lane_revisao_kind').value, 'codex');
  assert.equal(entries.get('lane_revisao_kind').source, 'project');
  assert.equal(entries.get('lane_criacao_kind').source, 'defaults');
});

test('the shipped defaults keep creator and reviewer in different CLIs', () => {
  const entries = loadTeacherConfig({ XDG_CONFIG_HOME: path.join(tmpDir(), 'nouser') }, tmpDir());
  assert.notEqual(entries.get('lane_criacao_kind').value, entries.get('lane_revisao_kind').value);
  assert.equal(entries.get('lane_revisao_roles').value, 'reviewer');
  assert.ok(!entries.get('lane_pesquisa_roles').value.split(',').includes('reviewer'));
  assert.ok(!entries.get('lane_criacao_roles').value.split(',').includes('reviewer'));
});

test('guards', () => {
  const env = { HERDR_SOHO_LANES: 'on' };
  assert.equal(guard(['spawn', 'professor'], env).code, 12);
  assert.equal(guard(['run', 'planner', 'b.md'], env).code, 12);
  assert.equal(guard(['spawn', 'reviewer', '--name', 'x'], env).code, 5);
  assert.equal(guard(['spawn', 'reviewer'], env), null);
  assert.equal(guard(['dispatch', 'pesquisa', 'b.md', '--role', 'reviewer'], env).code, 5);
  assert.equal(guard(['dispatch', 'criacao', 'b.md', '--role=reviewer'], env).code, 5);
  assert.equal(guard(['dispatch', 'revisao', 'b.md', '--role', 'content-creator'], env).code, 5);
  assert.equal(guard(['dispatch', 'revisao-2', 'b.md', '--role', 'researcher'], env).code, 5);
  assert.equal(guard(['dispatch', 'revisao', 'b.md', '--role', 'reviewer'], env), null);
  assert.equal(guard(['dispatch', 'pesquisa', 'b.md', '--role', 'debater'], env), null);
  assert.equal(guard(['spawn', 'researcher'], { HERDR_SOHO_LANES: 'off' }).code, 5);
  assert.equal(guard(['doctor', '--fix'], env).code, 2);
  assert.equal(guard(['doctor'], env), null);
  assert.equal(guard(['setup', '--dry-run'], env).code, 2);
  assert.equal(guard(['config', 'set', 'a', 'b'], env).code, 2);
  assert.equal(guard(['config'], env), null);
});

test('CLI: teacher-env runs without the engine and spawn professor exits 12', () => {
  const r = spawnSync(process.execPath, [WRAPPER, 'teacher-env'], { encoding: 'utf8', timeout: 20000 });
  assert.equal(r.status, 0, r.stderr);
  assert.match(r.stdout, /exported to the engine: .*HERDR_SOHO_LANE_REVISAO_ROLES/);
  const s = spawnSync(process.execPath, [WRAPPER, 'spawn', 'professor'], {
    encoding: 'utf8', timeout: 20000, env: { ...process.env, HERDR_TEACHER_SOHO_DIR: process.env.HERDR_TEACHER_SOHO_DIR || '' },
  });
  // Without herdr-soho installed the wrapper exits 2 before the guard.
  assert.ok(s.status === 12 || s.status === 2, `status ${s.status}: ${s.stderr}`);
});
