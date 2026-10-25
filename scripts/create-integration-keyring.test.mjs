import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, readFile, rm, stat } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { createIntegrationKeyring } from './create-integration-keyring.mjs';
test('provisioning refuses replacement; rotation retains old keys', async () => {
  const directory = await mkdtemp(join(tmpdir(), 'nodeherder-key-test-'));
  try { const path = join(directory, 'integration-keyring.json'); await createIntegrationKeyring(path); const before = JSON.parse(await readFile(path, 'utf8')); assert.equal((await stat(path)).mode & 0o777, 0o600); await assert.rejects(createIntegrationKeyring(path)); await createIntegrationKeyring(path, true); const after = JSON.parse(await readFile(path, 'utf8')); assert.notEqual(after.active_key_id, before.active_key_id); assert.equal(after.keys[before.active_key_id], before.keys[before.active_key_id]); }
  finally { await rm(directory, { recursive: true, force: true }); }
});
