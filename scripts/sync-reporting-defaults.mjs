#!/usr/bin/env node
// The backend JSON is authoritative. Commit the generated frontend artifact so
// frontend-only builds never need files outside their Docker/package boundary.
import { mkdir, readFile, writeFile } from 'node:fs/promises';

const source = new URL('../backend/models/settings/reporting-defaults.json', import.meta.url);
const target = new URL('../frontend/src/domain/reporting-defaults.json', import.meta.url);

async function main() {
  const mode = process.argv[2] ?? '--check';
  if (process.argv.length > 3 || !['--check', '--write'].includes(mode)) {
    throw new Error('Usage: node scripts/sync-reporting-defaults.mjs [--check|--write]');
  }
  const expected = await readFile(source);
  JSON.parse(expected.toString('utf8')); // Never generate or approve malformed JSON.
  if (mode === '--write') {
    await mkdir(new URL('.', target), { recursive: true });
    await writeFile(target, expected);
    console.log('Generated frontend reporting defaults from the canonical backend JSON.');
    return;
  }
  let actual;
  try {
    actual = await readFile(target);
  } catch (error) {
    if (error.code !== 'ENOENT') throw error;
  }
  if (!actual?.equals(expected)) {
    throw new Error('Frontend reporting defaults are missing or stale. Run node scripts/sync-reporting-defaults.mjs --write and commit both JSON files.');
  }
  console.log('Reporting defaults are synchronized.');
}

main().catch((error) => {
  console.error(error.message);
  process.exitCode = 1;
});
