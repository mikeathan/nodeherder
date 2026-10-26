import { randomBytes, randomUUID } from 'node:crypto';
import { realpath, readFile, open, rename, unlink, lstat } from 'node:fs/promises';
import { dirname, basename, resolve, relative, isAbsolute, sep } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

// Explicit provisioning/rotation only. Runtime never invents replacement encryption keys.
export async function createIntegrationKeyring(destination, rotate = false) {
  const directory = await realpath(dirname(resolve(destination)));
  const root = await realpath(fileURLToPath(new URL('../', import.meta.url)));
  const location = relative(root, directory);
  if (!(location === '..' || location.startsWith(`..${sep}`) || isAbsolute(location))) throw new Error('Keep the private keyring outside the repository.');
  const target = resolve(directory, 'integration-keyring.json');
  if (basename(destination) !== 'integration-keyring.json') throw new Error('Use a file named integration-keyring.json in an existing private directory.');
  let keys = {};
  if (rotate) {
    const info = await lstat(target);
    if (!info.isFile() || (info.mode & 0o077) !== 0) throw new Error('The existing keyring must have mode 0600.');
    const previous = JSON.parse(await readFile(target, 'utf8'));
    if (!previous.active_key_id || !previous.keys || !previous.keys[previous.active_key_id]) throw new Error('Invalid existing keyring; restore its backup first.');
    keys = previous.keys;
  }
  const id = randomUUID();
  keys[id] = randomBytes(32).toString('base64');
  const content = `${JSON.stringify({ active_key_id: id, keys }, null, 2)}\n`;
  if (!rotate) await writePrivate(target, content);
  else {
    const temporary = `${target}.${randomUUID()}.tmp`;
    try { await writePrivate(temporary, content); await rename(temporary, target); }
    finally { await unlink(temporary).catch(() => undefined); }
  }
  const parent = await open(directory, 'r');
  try { await parent.sync(); } finally { await parent.close(); }
  return target;
}
async function writePrivate(path,content) { const handle=await open(path,'wx',0o600);try{await handle.writeFile(content);await handle.sync();}finally{await handle.close();} }
if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  try { await createIntegrationKeyring(process.argv[2] ?? '', process.argv[3] === '--rotate'); process.stdout.write('Keyring saved. Back it up separately; keep all retiring keys for retained database backups.\n'); }
  catch (error) { process.stderr.write(`${error.message}\n`); process.exitCode = 1; }
}
