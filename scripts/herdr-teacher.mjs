#!/usr/bin/env node
// herdr-teacher — the pedagogical team on top of the herdr-soho engine.
//
// herdr-soho is the transport (spawn, dispatch, wait, lanes, family check,
// roster). This wrapper only:
//   1. points the engine at this skill's roles (HERDR_SOHO_ROLES);
//   2. layers the teacher config (skill defaults < user < project) and
//      exports each key as HERDR_SOHO_<KEY>, unless the environment already
//      sets it (an explicit env var or flag still wins);
//   3. enforces the segregation rules the engine cannot express on its own
//      (lanes on, professor never spawned, reviewer never on a creator or
//      researcher pane);
//   4. runs the engine with the arguments unchanged.
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const SCRIPT_DIR = path.dirname(fs.realpathSync(fileURLToPath(import.meta.url)));
const SKILL_DIR = path.resolve(SCRIPT_DIR, '..');

// Roles that only the professor (this pane) plays.
const ORCHESTRATOR_ROLES = new Set(['professor', 'planner']);
// Lane each segregated role must live in. The reviewer never shares a pane
// with the roles that produce the material it audits.
const AUTHOR_ROLES = new Set(['content-creator', 'researcher']);
const REVIEW_ROLE = 'reviewer';

function die(msg, code) {
  process.stderr.write(`herdr-teacher: ${msg}\n`);
  process.exit(code);
}

// Same normalization as herdr-soho (config.mjs normalizeKey).
export function normalizeKey(raw) {
  return raw.replace(/\s+/g, '').replace(/[.-]/g, '_').replace(/[^a-zA-Z0-9_]/g, '');
}

// key=value, one per line, `#` comments. Returns [{ key, value }] in order.
export function parseConfig(text) {
  const out = [];
  for (const raw of text.replace(/\r\n?/g, '\n').split('\n')) {
    const line = raw.trim();
    if (line === '' || line.startsWith('#')) continue;
    const eq = line.indexOf('=');
    if (eq <= 0) continue;
    out.push({ key: line.slice(0, eq).trim(), value: line.slice(eq + 1).trim() });
  }
  return out;
}

export function userConfigPath(env = process.env, platform = process.platform) {
  if (env.XDG_CONFIG_HOME) return path.join(env.XDG_CONFIG_HOME, 'herdr-teacher', 'config');
  if (platform === 'win32') {
    const base = env.APPDATA || path.join(os.homedir(), 'AppData', 'Roaming');
    return path.join(base, 'herdr-teacher', 'config');
  }
  return path.join(os.homedir(), '.config', 'herdr-teacher', 'config');
}

export function projectRoot(cwd = process.cwd()) {
  const r = spawnSync('git', ['rev-parse', '--show-toplevel'], { cwd, encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'] });
  const p = r.status === 0 ? (r.stdout || '').trim() : '';
  return p || cwd;
}

// Effective teacher config: Map<normalizedKey, { key, value, source }>.
export function loadTeacherConfig(env = process.env, cwd = process.cwd()) {
  const layers = [
    ['defaults', path.join(SKILL_DIR, 'config.defaults')],
    ['user', userConfigPath(env)],
    ['project', path.join(projectRoot(cwd), '.agents', 'herdr-teacher.conf')],
  ];
  const entries = new Map();
  for (const [source, file] of layers) {
    let text;
    try { text = fs.readFileSync(file, 'utf8'); } catch { continue; }
    for (const { key, value } of parseConfig(text)) entries.set(normalizeKey(key), { key, value, source, file });
  }
  return entries;
}

// The environment the engine runs with. `teacher.*` keys belong to this
// wrapper and the professor; they are never exported to the engine.
export function engineEnv(entries, env = process.env) {
  const out = { ...env };
  const exported = [];
  for (const [nk, e] of entries) {
    if (nk.startsWith('teacher_')) continue;
    const name = `HERDR_SOHO_${nk.toUpperCase()}`;
    if (out[name]) continue; // an explicit env var wins
    if (e.value === '') continue; // empty = unset, like herdr-soho
    out[name] = e.value;
    exported.push(name);
  }
  const rolesDir = path.join(SKILL_DIR, 'roles');
  if (!out.HERDR_SOHO_ROLES) out.HERDR_SOHO_ROLES = rolesDir;
  out.HERDR_TEACHER_SKILL_DIR = SKILL_DIR;
  return { env: out, exported };
}

function isSohoDir(d) {
  try { return fs.statSync(path.join(d, 'scripts', 'herdr-soho.mjs')).isFile(); } catch { return false; }
}

export function findSoho(env = process.env) {
  const candidates = [
    env.HERDR_TEACHER_SOHO_DIR,
    path.join(path.dirname(SKILL_DIR), 'herdr-soho'),
    path.join(os.homedir(), '.agents', 'skills', 'herdr-soho'),
    path.join(os.homedir(), '.claude', 'skills', 'herdr-soho'),
  ].filter(Boolean);
  for (const c of candidates) {
    let real;
    try { real = fs.realpathSync(c); } catch { continue; }
    if (isSohoDir(real)) return real;
  }
  return null;
}

// First positional argument after the subcommand (skipping `--flag value`
// pairs is not needed: role and agent are always the first positional).
function firstPositional(args) {
  return args.find((a) => !a.startsWith('-')) ?? '';
}

function flagValue(args, flag) {
  const i = args.indexOf(flag);
  if (i !== -1 && i + 1 < args.length) return args[i + 1];
  const eq = args.find((a) => a.startsWith(`${flag}=`));
  return eq ? eq.slice(flag.length + 1) : '';
}

// Segregation guards. Returns an error message or ''.
export function guard(argv, engine) {
  const [cmd, ...rest] = argv;
  const lanes = engine.HERDR_SOHO_LANES || 'on';
  if (lanes === 'off' && ['spawn', 'run', 'dispatch'].includes(cmd)) {
    return { code: 5, msg: 'lanes=off would let the reviewer reuse a researcher or content-creator pane. herdr-teacher requires lanes=on.' };
  }
  // These write herdr-soho's own files with the teacher's lanes in effect:
  // `doctor --fix` would delete the user's herdr-soho lanes (build/review do
  // not exist here), `setup` writes the herdr-soho block and hooks, and
  // `config set` writes the herdr-soho config. The teacher config is edited
  // by hand (config.defaults, user file, project file).
  if (cmd === 'doctor' && rest.includes('--fix')) {
    return { code: 2, msg: 'doctor --fix would rewrite your herdr-soho config against the teacher lanes. Run plain `doctor`; fix herdr-soho with its own script.' };
  }
  if (cmd === 'setup' || (cmd === 'config' && rest[0] === 'set')) {
    return { code: 2, msg: `${cmd}${cmd === 'config' ? ' set' : ''} writes herdr-soho files. Edit the teacher config instead (see: herdr-teacher teacher-env).` };
  }
  if (cmd === 'spawn' || cmd === 'run') {
    const role = firstPositional(rest);
    if (ORCHESTRATOR_ROLES.has(role)) {
      return { code: 12, msg: `spawn ${role}: the professor is the agent running this skill and does not open a pane. Teach in this session.` };
    }
    if (role === REVIEW_ROLE && rest.includes('--name')) {
      return { code: 5, msg: 'spawn reviewer --name: keep the lane name (revisao) so the segregation guard can recognize the reviewer pane.' };
    }
  }
  if (cmd === 'dispatch') {
    const agent = firstPositional(rest);
    const role = flagValue(rest, '--role');
    const reviewLane = (engine.HERDR_TEACHER_REVIEW_LANE || 'revisao');
    const isReviewPane = agent === reviewLane || agent.startsWith(`${reviewLane}-`);
    if (role === REVIEW_ROLE && !isReviewPane) {
      return { code: 5, msg: `dispatch ${agent} --role reviewer: the reviewer runs only in its own pane ('${reviewLane}'), never where research or content was produced.` };
    }
    if (AUTHOR_ROLES.has(role) && isReviewPane) {
      return { code: 5, msg: `dispatch ${agent} --role ${role}: the review pane never produces the material it audits.` };
    }
  }
  return null;
}

function printTeacherEnv(entries, exported, soho) {
  const lines = [];
  lines.push(`skill:   ${SKILL_DIR}`);
  lines.push(`engine:  ${soho ?? '(herdr-soho not found)'}`);
  lines.push(`roles:   ${path.join(SKILL_DIR, 'roles')}`);
  lines.push(`user:    ${userConfigPath()}`);
  lines.push(`project: ${path.join(projectRoot(), '.agents', 'herdr-teacher.conf')}`);
  lines.push('');
  for (const [, e] of entries) lines.push(`${e.key}=${e.value}  # ${e.source}`);
  lines.push('');
  lines.push(`exported to the engine: ${exported.length ? exported.join(' ') : '(none)'}`);
  process.stdout.write(`${lines.join('\n')}\n`);
}

function main() {
  const argv = process.argv.slice(2);
  const entries = loadTeacherConfig();
  const { env, exported } = engineEnv(entries);
  const reviewLane = entries.get('teacher_review_lane');
  if (reviewLane && !env.HERDR_TEACHER_REVIEW_LANE) env.HERDR_TEACHER_REVIEW_LANE = reviewLane.value;
  const soho = findSoho();

  if (argv[0] === 'teacher-env') {
    printTeacherEnv(entries, exported, soho);
    return 0;
  }
  if (!soho) {
    die('herdr-soho not found. Install it next to this skill (…/skills/herdr-soho) or set HERDR_TEACHER_SOHO_DIR.', 2);
  }
  const g = guard(argv, env);
  if (g) die(g.msg, g.code);

  env.HERDR_SOHO_SKILL_DIR = env.HERDR_SOHO_SKILL_DIR || soho;
  // HERDR_SOHO_BIN (the Go binary) is used only when set explicitly; the JS
  // engine is the one known to honor HERDR_SOHO_ROLES.
  const bin = env.HERDR_SOHO_BIN;
  const r = bin
    ? spawnSync(bin, argv, { stdio: 'inherit', env })
    : spawnSync(process.execPath, [path.join(soho, 'scripts', 'herdr-soho.mjs'), ...argv], { stdio: 'inherit', env });
  if (r.error) die(`could not run the engine: ${r.error.message}`, 2);
  if (r.signal) process.kill(process.pid, r.signal);
  return r.status ?? 1;
}

if (process.argv[1] && fs.realpathSync(process.argv[1]) === fs.realpathSync(fileURLToPath(import.meta.url))) {
  process.exit(main());
}
