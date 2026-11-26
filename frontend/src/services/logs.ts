/*
 * Log archive access. Routes are protected and already exist on the hub; this module only
 * validates names/sizes at the browser boundary and owns object-URL cleanup for downloads.
 */
import { get, post } from '@/contracts/api';

/** One archived log file. `name` is the path expected by the load endpoint. */
export type LogFile = { name: string; size: number; modified: string };

const MAX_DOWNLOAD_BYTES = 20 * 1024 * 1024;

/** Validate a hub-provided log name and derive the display filename. */
export function logFileName(path: string): string {
  const name = path.split(/[\\/]/).pop() ?? '';
  return name;
}

function invalid(path: string): boolean {
  return !path || path.startsWith('/') || path.includes('\0') || path.includes('..') || !path.endsWith('.log');
}

function isLogFile(value: unknown): value is LogFile {
  if (typeof value !== 'object' || value === null) return false;
  const name = Reflect.get(value, 'name');
  const size = Reflect.get(value, 'size');
  const modified = Reflect.get(value, 'modified');
  return typeof name === 'string' && typeof size === 'number' && Number.isFinite(size) && typeof modified === 'string' && Number.isFinite(Date.parse(modified));
}

/** Lists archived log files with size and modification time. */
export async function listLogFiles(): Promise<LogFile[]> {
  const response = await get('logfiles');
  if (!response.ok) throw new Error(`Could not list log files (${response.status}).`);
  const payload: unknown = await response.json();
  if (!Array.isArray(payload) || !payload.every(isLogFile)) {
    throw new Error('The hub returned an unexpected log list.');
  }
  return payload;
}

/** Loads one archived file as text. The caller owns confirmation and download. */
export async function loadLogFile(path: string): Promise<string> {
  if (invalid(path)) throw new Error('That log file name is not valid.');
  const response = await post('logfile', { file: path, action: 'load' });
  if (!response.ok) throw new Error(`Could not load ${logFileName(path)} (${response.status}).`);
  const declared = Number(response.headers.get('content-length'));
  if (Number.isFinite(declared) && declared > MAX_DOWNLOAD_BYTES) throw new Error('That log file is too large to download here.');
  const text = await response.text();
  if (text.length > MAX_DOWNLOAD_BYTES) throw new Error('That log file is too large to download here.');
  return text;
}

/** Triggers a browser download, releasing the object URL even if the click throws. */
export function downloadTextFile(filename: string, contents: string, type = 'text/plain;charset=utf-8'): void {
  const url = URL.createObjectURL(new Blob([contents], { type }));
  try {
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    link.click();
  } finally {
    URL.revokeObjectURL(url);
  }
}

/** Loads and downloads one archived log file. */
export async function downloadLogFile(path: string): Promise<void> {
  const text = await loadLogFile(path);
  downloadTextFile(logFileName(path), text);
}
