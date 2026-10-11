/*
 * Recent activity (spec 007 FR-05, US-09 / AC-27). In-memory only (FE-05), bounded.
 * A snapshot of last-known values gives each change a "from" value, because the store
 * is already updated when observers run.
 */
import { DeviceUpdate } from '@/types/device';

export const ACTIVITY_LIMIT = 200;

export type ActivityChange = { expose: string; from: unknown; to: unknown };
export type ActivityEntry = { id: number; deviceId: string; at: number; changes: ActivityChange[] };
/** Last-known value per "deviceId|expose". */
export type ValueSnapshot = Map<string, unknown>;

export const snapshotKey = (deviceId: string, expose: string) => `${deviceId}|${expose}`;

export function snapshotDevices(devices: { id: string; exposes: Record<string, { data: unknown }> }[]): ValueSnapshot {
  const snap: ValueSnapshot = new Map();
  for (const d of devices) for (const [name, e] of Object.entries(d.exposes)) snap.set(snapshotKey(d.id, name), e.data);
  return snap;
}

let seq = 0;

/**
 * Applies one device update: records changed values only (repeated identical values are not
 * activity) for exposes the snapshot knows, updates the snapshot, and caps the list.
 */
export function recordUpdate(entries: ActivityEntry[], snap: ValueSnapshot, update: DeviceUpdate, now: number, limit = ACTIVITY_LIMIT): ActivityEntry[] {
  const changes: ActivityChange[] = [];
  for (const [expose, to] of Object.entries(update.data ?? {})) {
    const key = snapshotKey(update.id, expose);
    if (!snap.has(key)) continue;
    const from = snap.get(key);
    snap.set(key, to);
    if (from !== to) changes.push({ expose, from, to });
  }
  if (!changes.length) return entries;
  return [{ id: ++seq, deviceId: update.id, at: now, changes }, ...entries].slice(0, limit);
}
