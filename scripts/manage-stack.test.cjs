const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawnSync } = require('node:child_process');

const source = fs.readFileSync(path.join(__dirname, 'manage-stack.sh'), 'utf8');
const versionSource = fs.readFileSync(path.join(__dirname, '../frontend/scripts/get-version.cjs'), 'utf8');

function fixture(t) {
  const root = fs.realpathSync(fs.mkdtempSync(path.join(os.tmpdir(), 'nodeherder stack ')));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  for (const dir of ['scripts', 'frontend/scripts', 'backend', 'bin', 'elsewhere']) {
    fs.mkdirSync(path.join(root, dir), { recursive: true });
  }
  fs.writeFileSync(path.join(root, 'scripts/manage-stack.sh'), source);
  fs.writeFileSync(path.join(root, 'frontend/scripts/get-version.cjs'), versionSource);
  // Only explicitly supplied tools are visible: never fall through to real Docker/Node/Git.
  for (const tool of ['dirname', 'sed', 'tr', 'mkdir', 'grep']) {
    const found = spawnSync('/bin/sh', ['-c', `command -v ${tool}`], { encoding: 'utf8' });
    assert.equal(found.status, 0);
    fs.symlinkSync(found.stdout.trim(), path.join(root, 'bin', tool));
  }
  for (const file of ['docker-compose.backend.yml', 'docker-compose.frontend.yml', 'docker-compose.connections.yml']) {
    fs.writeFileSync(path.join(root, file), 'services: {}\n');
  }
  fs.writeFileSync(path.join(root, 'frontend/package.json'), '{"version":"1.2.3"}\n');
  fs.writeFileSync(path.join(root, 'bin/docker'), `#!/bin/bash
printf '%s\\t' "$@" >> "$MOCK_LOG"
printf '\\n' >> "$MOCK_LOG"
if [[ " $* " == *" config "* && " $* " != *" --quiet "* && -n "\${MOCK_PROJECT:-}" ]]; then printf 'name: %s\\n' "$MOCK_PROJECT"; fi
if [[ " $* " == *" network ls "* && -n "\${MOCK_NETWORK:-}" ]]; then printf '%s\\n' "$MOCK_NETWORK"; fi
if [[ " $* " == *" network inspect "* && -n "\${MOCK_GATEWAY:-}" ]]; then printf '%s\\n' "$MOCK_GATEWAY"; fi
if [[ " $* " == *" ps -aq frontend "* && -n "\${MOCK_FRONTEND_IP:-}" ]]; then printf 'frontend-id\\n'; fi
if [[ " $* " == *" inspect frontend-id "* && -n "\${MOCK_FRONTEND_IP:-}" ]]; then printf '%s\\n' "$MOCK_FRONTEND_IP"; fi
if [[ -n "\${MOCK_FAIL:-}" && " $* " == *" $MOCK_FAIL "* ]]; then exit 17; fi
`, { mode: 0o755 });
  const log = path.join(root, 'calls');
  return {
    root,
    run(args, extra = {}) {
      fs.writeFileSync(log, '');
      const result = spawnSync('/bin/bash', [path.join(root, 'scripts/manage-stack.sh'), ...args], {
        cwd: path.join(root, 'elsewhere'), encoding: 'utf8',
        env: { ...process.env, APP_VERSION: '1.2.3', NO_COLOR: '1',
          PATH: `${root}/bin`, MOCK_LOG: log,
          // Compose reports a normalized project name and the networks it owns.
          MOCK_PROJECT: extra.MOCK_GATEWAY ? 'fixture' : '',
          MOCK_NETWORK: extra.MOCK_GATEWAY ? 'fixture_nodeherder' : '',
          ...extra },
      });
      const calls = fs.readFileSync(log, 'utf8').trim().split('\n').filter(Boolean)
        .map(line => line.split('\t').filter(Boolean));
      return { ...result, calls, output: result.stdout + result.stderr };
    },
  };
}

const has = (r, value) => r.calls.some(args => args.includes(value));
const operation = (r, name) => r.calls.find(args => args.includes(name));

test('backend-only start succeeds from another directory', t => {
  const f = fixture(t), r = f.run(['start', 'backend']);
  assert.equal(r.status, 0, r.output);
  const up = operation(r, 'up');
  assert.deepEqual(up.slice(up.indexOf('up')), ['up', '-d', 'backend', 'mqtt', 'zigbee2mqtt']);
  assert.ok(up.includes(path.join(f.root, 'docker-compose.backend.yml')));
  assert.ok(!up.includes(path.join(f.root, 'docker-compose.frontend.yml')));
});

test('all uses both files and one up command', t => {
  const f = fixture(t), r = f.run(['start']);
  assert.equal(r.status, 0, r.output);
  const up = operation(r, 'up');
  assert.ok(up.includes(path.join(f.root, 'docker-compose.frontend.yml')));
  assert.equal(r.calls.filter(c => c.includes('up')).length, 1);
  assert.deepEqual(up.slice(-4), ['backend', 'mqtt', 'zigbee2mqtt', 'frontend']);
});

test('frontend operations do not reference backend configuration', t => {
  const f = fixture(t), r = f.run(['start', 'frontend']);
  assert.equal(r.status, 0, r.output);
  assert.ok(!r.calls.flat().includes(path.join(f.root, 'docker-compose.backend.yml')));
  assert.deepEqual(operation(r, 'up').slice(-1), ['frontend']);
});

for (const args of [['start', 'unknown'], ['unknown'], ['start', 'all', 'extra'],
  ['stop', '--no-cache'], ['start', '--follow'], ['logs', '--tail'],
  ['logs', '--tail', '-1'], ['logs', '--tail', 'bad'], ['start', '--bad']]) {
  test(`invalid input ${args.join(' ')} fails before Docker`, t => {
    const r = fixture(t).run(args);
    assert.notEqual(r.status, 0);
    assert.equal(r.calls.length, 0);
  });
}

test('help does not invoke Docker', t => {
  for (const args of [[], ['help'], ['--help']]) {
    const r = fixture(t).run(args);
    assert.equal(r.status, 0, r.output);
    assert.equal(r.calls.length, 0);
    assert.match(r.output, /validate/);
  }
});

test('dry-run previews only, including rebuild, without invoking Docker', t => {
  const r = fixture(t).run(['rebuild', 'all', '--dry-run', '--no-cache']);
  assert.equal(r.status, 0, r.output);
  assert.equal(r.calls.length, 0);
  assert.match(r.output, /--no-cache/);
  assert.match(r.output, /APP_VERSION=1.2.3/);
  assert.match(r.output, /--no-build/);
  assert.doesNotMatch(r.output, /\x1b\[/);
});

for (const step of ['pull', 'build']) {
  test(`rebuild ${step} failure leaves containers untouched`, t => {
    const r = fixture(t).run(['rebuild'], { MOCK_FAIL: step });
    assert.equal(r.status, 17, r.output);
    for (const forbidden of ['stop', 'down', 'rm', 'up']) assert.equal(has(r, forbidden), false);
    assert.doesNotMatch(r.output, /Completed/);
  });
}

test('rebuild pulls/builds before replacement and uses cache by default', t => {
  const r = fixture(t).run(['rebuild']);
  assert.equal(r.status, 0, r.output);
  assert.ok(r.calls.indexOf(operation(r, 'pull')) < r.calls.indexOf(operation(r, 'build')));
  assert.ok(r.calls.indexOf(operation(r, 'build')) < r.calls.indexOf(operation(r, 'up')));
  assert.ok(operation(r, 'pull').includes('--ignore-buildable'));
  assert.ok(operation(r, 'build').includes('APP_VERSION=1.2.3'));
  assert.equal(has(r, '--no-cache'), false);
  for (const forbidden of ['stop', 'down', 'rm', '--volumes', '--remove-orphans', '--rmi']) {
    assert.equal(has(r, forbidden), false);
  }
});

test('partial down preserves the shared network and persistent data', t => {
  const r = fixture(t).run(['down', 'backend']);
  assert.equal(r.status, 0, r.output);
  assert.equal(has(r, 'down'), false);
  assert.deepEqual(operation(r, 'rm').slice(-4), ['--force', 'backend', 'mqtt', 'zigbee2mqtt']);
  assert.equal(has(r, '--volumes'), false);
});

test('all down uses one combined down, without deleting volumes/images/orphans', t => {
  const r = fixture(t).run(['down']);
  assert.equal(r.status, 0, r.output);
  assert.equal(r.calls.filter(c => c.includes('down')).length, 1);
  for (const flag of ['--volumes', '--rmi', '--remove-orphans']) assert.equal(has(r, flag), false);
});

test('logs accepts scoped follow and tail', t => {
  const r = fixture(t).run(['logs', 'frontend', '--follow', '--tail', '25']);
  assert.equal(r.status, 0, r.output);
  assert.deepEqual(operation(r, 'logs').slice(-5), ['--tail', '25', '--follow', '--timestamps', 'frontend']);
});

test('validate uses quiet config and needs no daemon', t => {
  const r = fixture(t).run(['validate'], { MOCK_FAIL: 'info' });
  assert.equal(r.status, 0, r.output);
  assert.ok(operation(r, 'config').includes('--quiet'));
  assert.equal(has(r, 'info'), false);
});

for (const step of ['config', 'info']) {
  test(`failed ${step} preflight prevents mutations`, t => {
    const r = fixture(t).run(['start'], { MOCK_FAIL: step });
    assert.notEqual(r.status, 0);
    assert.equal(has(r, 'up'), false);
  });
}

test('status includes stopped containers', t => {
  const r = fixture(t).run(['status', 'backend']);
  assert.equal(r.status, 0, r.output);
  assert.ok(operation(r, 'ps').includes('--all'));
});

test('build works without host Node using package version fallback', t => {
  const r = fixture(t).run(['build', 'frontend'], { APP_VERSION: '' });
  assert.equal(r.status, 0, r.output);
  assert.ok(operation(r, 'build').includes('APP_VERSION=1.2.3'));
  assert.equal(has(r, 'up'), false);
});

test('explicit version is validated before Docker', t => {
  const r = fixture(t).run(['build'], { APP_VERSION: 'invalid version' });
  assert.notEqual(r.status, 0);
  assert.equal(r.calls.length, 0);
});

test('no-color disables escape codes', t => {
  const r = fixture(t).run(['status', '--no-color'], { NO_COLOR: '' });
  assert.equal(r.status, 0, r.output);
  assert.doesNotMatch(r.output, /\x1b\[/);
});

test('build forwards no-cache only when requested', t => {
  const r = fixture(t).run(['build', 'frontend', '--no-cache']);
  assert.equal(r.status, 0, r.output);
  assert.ok(operation(r, 'build').includes('--no-cache'));
});

test('Docker missing is actionable and fails before operations', t => {
  const f = fixture(t);
  fs.unlinkSync(path.join(f.root, 'bin/docker'));
  const r = f.run(['start']);
  assert.notEqual(r.status, 0);
  assert.match(r.output, /Docker is missing/);
  assert.equal(r.calls.length, 0);
});

test('Compose missing is actionable and fails before operations', t => {
  const r = fixture(t).run(['start'], { MOCK_FAIL: 'version' });
  assert.notEqual(r.status, 0);
  assert.match(r.output, /Compose v2 is unavailable/);
  assert.equal(has(r, 'up'), false);
});

test('dry-run works without Docker', t => {
  const f = fixture(t);
  fs.unlinkSync(path.join(f.root, 'bin/docker'));
  const r = f.run(['start', '--dry-run']);
  assert.equal(r.status, 0, r.output);
  assert.equal(r.calls.length, 0);
});

for (const [branch, expected] of [['main', '2.3.4'], ['feature/example', '2.3.4-dev'], ['', '2.3.4']]) {
  test(`version helper resolves repository cwd on branch ${branch || 'detached'}`, t => {
    const f = fixture(t);
    fs.writeFileSync(path.join(f.root, 'bin/git'), `#!/bin/bash
[[ "$PWD" == "$EXPECTED_ROOT" ]] || exit 9
case "$1" in describe) printf 'v2.3.4-2-gabcdef\\n';; branch) printf '%s\\n' "$TEST_BRANCH";; esac
`, { mode: 0o755 });
    const r = spawnSync(process.execPath, [path.join(f.root, 'frontend/scripts/get-version.cjs')], {
      cwd: path.join(f.root, 'elsewhere'), encoding: 'utf8',
      env: { ...process.env, PATH: path.join(f.root, 'bin'), EXPECTED_ROOT: f.root, TEST_BRANCH: branch },
    });
    assert.equal(r.status, 0, r.stderr);
    assert.equal(r.stdout.trim(), expected);
  });
}

test('version helper fallback prints only the package version', t => {
  const f = fixture(t);
  const r = spawnSync(process.execPath, [path.join(f.root, 'frontend/scripts/get-version.cjs')], {
    cwd: path.join(f.root, 'elsewhere'), encoding: 'utf8',
    env: { ...process.env, PATH: path.join(f.root, 'bin') },
  });
  assert.equal(r.status, 0, r.stderr);
  assert.equal(r.stdout, '1.2.3\n');
});

test('build uses git-aware version when host Node is available', t => {
  const f = fixture(t);
  fs.symlinkSync(process.execPath, path.join(f.root, 'bin/node'));
  const r = f.run(['build', 'frontend'], { APP_VERSION: '' });
  assert.equal(r.status, 0, r.output);
  assert.ok(operation(r, 'build').includes('APP_VERSION=1.2.3'));
});

for (const action of ['stop', 'restart']) {
  for (const scope of ['all', 'backend', 'frontend']) {
    test(`${action} ${scope} targets only selected services`, t => {
      const r = fixture(t).run([action, scope]);
      assert.equal(r.status, 0, r.output);
      const services = scope === 'frontend' ? ['frontend'] : ['backend', 'mqtt', 'zigbee2mqtt'];
      if (scope === 'all') services.push('frontend');
      const call = operation(r, action);
      assert.deepEqual(call.slice(call.indexOf(action) + 1), services);
      for (const forbidden of ['up', 'build', 'down', 'rm']) assert.equal(has(r, forbidden), false);
    });
  }
}

test('frontend down stops before removing only frontend', t => {
  const r = fixture(t).run(['down', 'frontend']);
  assert.equal(r.status, 0, r.output);
  assert.deepEqual(operation(r, 'stop').slice(-1), ['frontend']);
  assert.deepEqual(operation(r, 'rm').slice(-2), ['--force', 'frontend']);
  assert.ok(r.calls.indexOf(operation(r, 'stop')) < r.calls.indexOf(operation(r, 'rm')));
  assert.equal(has(r, 'down'), false);
});

test('failed partial stop prevents removal', t => {
  const r = fixture(t).run(['down', 'backend'], { MOCK_FAIL: 'stop' });
  assert.equal(r.status, 17, r.output);
  assert.equal(has(r, 'rm'), false);
  assert.doesNotMatch(r.output, /Completed/);
});

for (const action of ['down', 'rebuild']) {
  test(`failed ${action} final step preserves failure status`, t => {
    const r = fixture(t).run([action], { MOCK_FAIL: action === 'down' ? 'down' : 'up' });
    assert.equal(r.status, 17, r.output);
    assert.doesNotMatch(r.output, /Completed|Containers started/);
  });
}

for (const action of ['start', 'stop', 'down', 'restart', 'pull', 'build', 'rebuild', 'status', 'logs', 'validate']) {
  test(`${action} dry-run never invokes Docker`, t => {
    const f = fixture(t);
    fs.unlinkSync(path.join(f.root, 'bin/docker'));
    const r = f.run([action, '--dry-run']);
    assert.equal(r.status, 0, r.output);
    assert.equal(r.calls.length, 0);
    assert.match(r.output, /Preview complete/);
  });
}

test('start resolves trusted proxy peers for a clean install', t => {
  const f = fixture(t);
  const r = f.run(['start', 'backend'], { MOCK_GATEWAY: '172.18.0.1' });
  assert.equal(r.status, 0, r.output);
  const override = path.join(f.root, '.generated', 'compose.proxy-allowlist.yml');
  assert.equal(
    fs.readFileSync(override, 'utf8'),
    'services:\n  backend:\n    environment:\n      TRUSTED_PROXY_CIDRS: "172.18.0.1/32"\n',
  );
  assert.ok(operation(r, 'up').includes(override));
  assert.match(r.output, /Trusted proxy peers: 172\.18\.0\.1\/32/);
});

test('resolved peers include the frontend container without duplicating the gateway', t => {
  const f = fixture(t);
  const override = path.join(f.root, '.generated', 'compose.proxy-allowlist.yml');
  const both = f.run(['start'], { MOCK_GATEWAY: '172.18.0.1', MOCK_FRONTEND_IP: '172.18.0.2' });
  assert.equal(both.status, 0, both.output);
  assert.match(fs.readFileSync(override, 'utf8'), /TRUSTED_PROXY_CIDRS: "172\.18\.0\.1\/32,172\.18\.0\.2\/32"/);

  const same = f.run(['start'], { MOCK_GATEWAY: '172.18.0.1', MOCK_FRONTEND_IP: '172.18.0.1' });
  assert.equal(same.status, 0, same.output);
  assert.match(fs.readFileSync(override, 'utf8'), /TRUSTED_PROXY_CIDRS: "172\.18\.0\.1\/32"/);
});

test('configured proxy allowlist is never overridden', t => {
  const f = fixture(t);
  fs.writeFileSync(path.join(f.root, 'backend/.env'), 'TRUSTED_PROXY_CIDRS=10.0.0.5/32\n');
  const fromFile = f.run(['start', 'backend'], { MOCK_GATEWAY: '172.18.0.1' });
  assert.equal(fromFile.status, 0, fromFile.output);
  assert.equal(has(fromFile, 'create'), false);
  assert.equal(fs.existsSync(path.join(f.root, '.generated')), false);
  assert.match(fromFile.output, /Using configured TRUSTED_PROXY_CIDRS/);

  const fromShell = f.run(['start', 'backend'], { MOCK_GATEWAY: '172.18.0.1', TRUSTED_PROXY_CIDRS: '10.0.0.6/32' });
  assert.equal(fromShell.status, 0, fromShell.output);
  assert.equal(fs.existsSync(path.join(f.root, '.generated')), false);
});

test('unresolvable proxy peer warns and still starts', t => {
  const f = fixture(t);
  const r = f.run(['start', 'backend']);
  assert.equal(r.status, 0, r.output);
  assert.match(r.output, /TRUSTED_PROXY_CIDRS manually/);
  assert.equal(fs.existsSync(path.join(f.root, '.generated')), false);
  assert.ok(has(r, 'up'));

  const placeholder = f.run(['start', 'backend'], { MOCK_GATEWAY: '<no value>' });
  assert.equal(placeholder.status, 0, placeholder.output);
  assert.equal(fs.existsSync(path.join(f.root, '.generated')), false);
});

test('frontend-only start never resolves backend proxy peers', t => {
  const f = fixture(t);
  const r = f.run(['start', 'frontend'], { MOCK_GATEWAY: '172.18.0.1' });
  assert.equal(r.status, 0, r.output);
  assert.equal(has(r, 'create'), false);
  assert.equal(fs.existsSync(path.join(f.root, '.generated')), false);
});

test('malformed inspector output is never trusted', t => {
  const f = fixture(t);
  const override = path.join(f.root, '.generated', 'compose.proxy-allowlist.yml');
  const r = f.run(['start'], { MOCK_GATEWAY: 'invalid IP', MOCK_FRONTEND_IP: 'not-an-address' });
  assert.equal(r.status, 0, r.output);
  assert.match(r.output, /TRUSTED_PROXY_CIDRS manually/);
  assert.equal(fs.existsSync(override), false);

  const ipv6 = f.run(['start', 'backend'], { MOCK_GATEWAY: 'fd00::1' });
  assert.equal(ipv6.status, 0, ipv6.output);
  assert.match(fs.readFileSync(override, 'utf8'), /TRUSTED_PROXY_CIDRS: "fd00::1\/128"/);
});

test('project name comes from Compose, not the checkout directory', t => {
  const f = fixture(t);
  // A directory Compose normalizes differently must still resolve its own network.
  const r = f.run(['start', 'backend'], {
    MOCK_PROJECT: 'nodeherder', MOCK_NETWORK: 'nodeherder_nodeherder', MOCK_GATEWAY: '172.19.0.1',
  });
  assert.equal(r.status, 0, r.output);
  const override = path.join(f.root, '.generated', 'compose.proxy-allowlist.yml');
  assert.match(fs.readFileSync(override, 'utf8'), /TRUSTED_PROXY_CIDRS: "172\.19\.0\.1\/32"/);
  assert.ok(
    r.calls.some((args) => args.includes('label=com.docker.compose.project=nodeherder')),
    'network lookup must use the Compose project label',
  );
});

test('a project without a resolvable gateway warns instead of trusting a guess', t => {
  const f = fixture(t);
  const r = f.run(['start', 'backend'], { MOCK_PROJECT: 'nodeherder' });
  assert.equal(r.status, 0, r.output);
  assert.match(r.output, /TRUSTED_PROXY_CIDRS manually/);
  assert.equal(fs.existsSync(path.join(f.root, '.generated')), false);
});

test('Connections key mount is explicit and never affects frontend-only commands', t => {
  const f = fixture(t);
  const r = f.run(['start', 'backend', '--connections']);
  assert.equal(r.status, 0, r.output);
  assert.ok(operation(r, 'up').includes(path.join(f.root, 'docker-compose.connections.yml')));
  assert.ok(!operation(f.run(['start', 'backend']), 'up').includes(path.join(f.root, 'docker-compose.connections.yml')));
  const bad = f.run(['start', 'frontend', '--connections']);
  assert.notEqual(bad.status, 0);
  assert.equal(bad.calls.length, 0);
});
