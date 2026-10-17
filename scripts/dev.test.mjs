import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, realpathSync, mkdirSync, copyFileSync, writeFileSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawn } from 'node:child_process';
import { once } from 'node:events';
import { setTimeout as delay } from 'node:timers/promises';

function fixture(t, mode = '') {
  const root = realpathSync(mkdtempSync(join(tmpdir(), 'nodeherder dev ')));
  for (const dir of ['scripts', 'backend', 'frontend', 'bin']) mkdirSync(join(root, dir));
  copyFileSync(new URL('./dev.mjs', import.meta.url), join(root, 'scripts/dev.mjs'));
  const log = join(root, 'calls');
  const record = `const fs = require('node:fs'); fs.appendFileSync(process.env.CALL_LOG, JSON.stringify({ tool: TOOL, args: process.argv.slice(2), cwd: process.cwd(), env: process.env.APP_ENV, pid: process.pid }) + '\\n');`;
  const service = `
if (process.env.TEST_MODE === 'stubborn') process.on('SIGTERM', () => {});
const { spawn } = require('node:child_process');
const descendant = spawn(process.execPath, ['-e', 'setInterval(() => {}, 1000)'], { stdio: 'ignore' });
fs.appendFileSync(process.env.CALL_LOG, JSON.stringify({tool: 'descendant', pid: descendant.pid}) + '\\n');
setInterval(() => {}, 1000);
`;
  const tools = {
    go: `const TOOL = 'go'; ${record}
if (process.env.TEST_MODE === 'build-fail') process.exit(7);
fs.copyFileSync(process.env.FAKE_BACKEND, 'nodeherder'); fs.chmodSync('nodeherder', 0o755);`,
    npm: `const TOOL = 'npm'; ${record}
if (process.env.TEST_MODE === 'frontend-fail') process.exit(9);
${service}`,
    backend: `const TOOL = 'backend'; ${record} ${service}`,
  };
  for (const [name, body] of Object.entries(tools)) {
    writeFileSync(join(root, 'bin', name), `#!${process.execPath}\n${body}\n`, { mode: 0o755 });
  }
  const child = spawn(process.execPath, [join(root, 'scripts/dev.mjs')], {
    cwd: tmpdir(), stdio: ['ignore', 'pipe', 'pipe'],
    env: { ...process.env, PATH: join(root, 'bin'), APP_ENV: 'development', CALL_LOG: log,
      FAKE_BACKEND: join(root, 'bin/backend'), TEST_MODE: mode },
  });
  const closed = once(child, 'close');
  let output = '';
  child.stdout.on('data', (chunk) => { output += chunk; });
  child.stderr.on('data', (chunk) => { output += chunk; });
  const calls = () => {
    try { return readFileSync(log, 'utf8').trim().split('\n').filter(Boolean).map(JSON.parse); }
    catch (error) { if (error.code === 'ENOENT') return []; throw error; }
  };
  t.after(async () => {
    if (child.exitCode === null && child.signalCode === null) child.kill('SIGTERM');
    await closed;
    rmSync(root, { recursive: true, force: true });
  });
  return { root, child, closed, calls, output: () => output };
}

async function until(predicate) {
  const deadline = Date.now() + 5000;
  while (!predicate()) {
    assert.ok(Date.now() < deadline, 'Timed out waiting for fake services');
    await delay(20);
  }
}

function assertStopped(calls) {
  for (const { pid } of calls) {
    assert.throws(() => process.kill(pid, 0), { code: 'ESRCH' });
  }
}

for (const signal of ['SIGINT', 'SIGTERM']) {
  test(`starts in component directories and cleans process trees on ${signal}`, { timeout: 10000 }, async (t) => {
    const f = fixture(t);
    await until(() => f.calls().filter((c) => c.tool === 'descendant').length === 2);
    const calls = f.calls();
    assert.equal(calls[0].tool, 'go');
    assert.deepEqual(calls[0].args, ['build', '-o', 'nodeherder', '.']);
    for (const call of calls.filter((c) => c.tool !== 'descendant')) {
      assert.equal(call.cwd, join(f.root, call.tool === 'npm' ? 'frontend' : 'backend'));
      assert.equal(call.env, 'development');
    }
    assert.deepEqual(calls.find((c) => c.tool === 'npm').args, ['run', 'dev', '--', '--strictPort']);
    f.child.kill(signal);
    assert.equal((await f.closed)[0], signal === 'SIGINT' ? 130 : 143);
    await until(() => {
      try { assertStopped(calls); return true; } catch { return false; }
    });
  });
}

test('build failure prevents starting services and preserves failure code', { timeout: 10000 }, async (t) => {
  const f = fixture(t, 'build-fail');
  assert.equal((await f.closed)[0], 7, f.output());
  assert.deepEqual(f.calls().map((c) => c.tool), ['go']);
});

test('frontend failure stops the backend and returns failure', { timeout: 10000 }, async (t) => {
  const f = fixture(t, 'frontend-fail');
  assert.equal((await f.closed)[0], 9, f.output());
  await until(() => {
    try { assertStopped(f.calls()); return true; } catch { return false; }
  });
});


test('forces shutdown when services ignore SIGTERM', { timeout: 12000 }, async (t) => {
  const f = fixture(t, 'stubborn');
  await until(() => f.calls().filter((c) => c.tool === 'descendant').length === 2);
  f.child.kill('SIGTERM');
  assert.equal((await f.closed)[0], 143, f.output());
  await until(() => {
    try { assertStopped(f.calls()); return true; } catch { return false; }
  });
});
