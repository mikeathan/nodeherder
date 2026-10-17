import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, copyFileSync, writeFileSync, readFileSync, existsSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawnSync } from 'node:child_process';

function fixture(t) {
  const root = mkdtempSync(join(tmpdir(), 'nodeherder defaults '));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  mkdirSync(join(root, 'scripts'));
  mkdirSync(join(root, 'backend/models/settings'), { recursive: true });
  const script = join(root, 'scripts/sync-reporting-defaults.mjs');
  copyFileSync(new URL('./sync-reporting-defaults.mjs', import.meta.url), script);
  const source = join(root, 'backend/models/settings/reporting-defaults.json');
  const target = join(root, 'frontend/src/domain/reporting-defaults.json');
  const canonical = '{\n  "commandConfirmationTimeoutMs": 7500\n}\n';
  writeFileSync(source, canonical);
  // Deliberately run elsewhere: all paths must be based on the script location.
  const run = (...args) => spawnSync(process.execPath, [script, ...args], { cwd: tmpdir(), encoding: 'utf8' });
  return { source, target, canonical, run };
}

test('missing artifact fails check without creating or changing files', (t) => {
  const f = fixture(t);
  const result = f.run('--check');
  assert.equal(result.status, 1);
  assert.match(result.stderr, /--write/);
  assert.equal(existsSync(f.target), false);
  assert.equal(readFileSync(f.source, 'utf8'), f.canonical);
});

test('explicit generation is deterministic and stale checks are read-only', (t) => {
  const f = fixture(t);
  assert.equal(f.run('--write').status, 0);
  assert.equal(readFileSync(f.target, 'utf8'), f.canonical);
  assert.equal(f.run('--check').status, 0);
  assert.equal(f.run().status, 0);
  writeFileSync(f.target, '{}\n');
  assert.equal(f.run('--check').status, 1);
  assert.equal(readFileSync(f.target, 'utf8'), '{}\n');
  assert.equal(f.run('--write').status, 0);
  assert.equal(readFileSync(f.target, 'utf8'), f.canonical);
});

test('invalid canonical JSON does not overwrite the generated artifact', (t) => {
  const f = fixture(t);
  assert.equal(f.run('--write').status, 0);
  writeFileSync(f.source, '{bad json');
  assert.equal(f.run('--write').status, 1);
  assert.equal(f.run('--check').status, 1);
  assert.equal(readFileSync(f.target, 'utf8'), f.canonical);
});

test('unsupported arguments fail without writing an artifact', (t) => {
  const f = fixture(t);
  assert.equal(f.run('--invalid').status, 1);
  assert.equal(f.run('--check', '--write').status, 1);
  assert.equal(existsSync(f.target), false);
});
