const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawnSync } = require('node:child_process');

// Opt-in: real Compose config only, never containers or production environment files.
test('real Compose configuration', {
  skip: process.env.NODEHERDER_COMPOSE_TESTS !== '1' && 'Set NODEHERDER_COMPOSE_TESTS=1 to enable',
}, async t => {
  function fixture(t) {
    const root = fs.realpathSync(fs.mkdtempSync(path.join(os.tmpdir(), 'nodeherder compose ')));
    t.after(() => fs.rmSync(root, { recursive: true, force: true }));
    for (const dir of ['scripts', 'backend', 'frontend', 'elsewhere', 'bin', '.docker']) {
      fs.mkdirSync(path.join(root, dir));
    }
    for (const file of ['scripts/manage-stack.sh', 'docker-compose.backend.yml', 'docker-compose.frontend.yml']) {
      fs.copyFileSync(path.join(__dirname, '..', file), path.join(root, file));
    }
    fs.writeFileSync(path.join(root, 'backend/.env'), 'MQTT_USER=test\nMQTT_PASS=fixture-only\n');
    fs.writeFileSync(path.join(root, 'backend/.env.production'), 'APP_ENV=production\n');
    const env = {
      PATH: process.env.PATH, NO_COLOR: '1', COMPOSE_PROJECT_NAME: 'nodeherder-check',
      DOCKER_CONFIG: path.join(root, '.docker'), DOCKER_HOST: 'tcp://127.0.0.1:1',
      DATA_ROOT: path.join(root, 'data'), Z2M_DEVICE: '/dev/fixture-zigbee', Z2M_SERIAL_PORT: '/dev/ttyUSB0',
    };
    // Optional standalone binary adapter, useful when the Docker CLI lacks plugin discovery.
    if (process.env.NODEHERDER_COMPOSE_BINARY) {
      env.NODEHERDER_COMPOSE_BINARY = process.env.NODEHERDER_COMPOSE_BINARY;
      env.PATH = `${path.join(root, 'bin')}${path.delimiter}${env.PATH}`;
      fs.writeFileSync(path.join(root, 'bin/docker'), `#!/bin/sh
[ "$1" = compose ] || exit 91
shift
exec "$NODEHERDER_COMPOSE_BINARY" "$@"
`, { mode: 0o755 });
    }
    const run = (cmd, args, overrides = {}) => spawnSync(cmd, args, {
      cwd: path.join(root, 'elsewhere'), env: { ...env, ...overrides }, encoding: 'utf8', timeout: 15000,
    });
    return {
      root, env, run,
      validate: (scope, overrides) => run('/bin/bash', [path.join(root, 'scripts/manage-stack.sh'), 'validate', scope], overrides),
    };
  }

  for (const scope of ['backend', 'frontend', 'all']) {
    await t.test(`${scope} validates from another directory without a daemon`, t => {
      const r = fixture(t).validate(scope);
      assert.equal(r.status, 0, r.stderr);
      assert.match(r.stdout, /Configuration valid/);
      assert.doesNotMatch(r.stdout, /fixture-only/);
    });
  }

  for (const name of ['DATA_ROOT', 'Z2M_DEVICE', 'Z2M_SERIAL_PORT']) {
    for (const value of ['', undefined]) {
      await t.test(`${name} rejects ${value === '' ? 'empty' : 'missing'} values`, t => {
        const r = fixture(t).validate('backend', { [name]: value });
        assert.notEqual(r.status, 0);
        assert.match(r.stderr, new RegExp(`Set ${name}`));
        assert.doesNotMatch(r.stdout, /Completed/);
      });
    }
  }

  await t.test('frontend needs neither backend environment files nor Zigbee variables', t => {
    const f = fixture(t);
    for (const file of ['.env', '.env.production']) fs.unlinkSync(path.join(f.root, 'backend', file));
    const r = f.validate('frontend', { DATA_ROOT: undefined, Z2M_DEVICE: undefined, Z2M_SERIAL_PORT: undefined });
    assert.equal(r.status, 0, r.stderr);
  });

  await t.test('combined model preserves build paths, mounts, ports and shared project network', t => {
    const f = fixture(t);
    const r = f.run('docker', ['compose', '--project-directory', f.root,
      '-f', path.join(f.root, 'docker-compose.backend.yml'),
      '-f', path.join(f.root, 'docker-compose.frontend.yml'), 'config', '--format', 'json']);
    assert.equal(r.status, 0, r.stderr);
    const model = JSON.parse(r.stdout), services = model.services;
    assert.deepEqual(Object.keys(services).sort(), ['backend', 'frontend', 'mqtt', 'zigbee2mqtt']);
    assert.equal(model.name, 'nodeherder-check');
    assert.equal(model.networks.nodeherder.name, 'nodeherder-check_nodeherder');
    for (const [name, port] of [['backend', 4110], ['frontend', 80], ['mqtt', 1883], ['zigbee2mqtt', 8089]]) {
      assert.equal(String(services[name].ports[0].published), String(port));
      assert.ok(Object.hasOwn(services[name].networks, 'nodeherder'));
    }
    for (const name of ['backend', 'frontend']) {
      assert.equal(services[name].build.context, path.join(f.root, name));
    }
    assert.ok(services.backend.volumes.some(v => v.source === path.join(f.root, 'backend/data')));
    assert.ok(services.mqtt.volumes.some(v => v.source === path.join(f.env.DATA_ROOT, 'mqtt/data')));
    assert.ok(services.zigbee2mqtt.devices.some(d => d.source === f.env.Z2M_DEVICE && d.target === f.env.Z2M_SERIAL_PORT));
  });
});
