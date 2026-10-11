/*
 * Device command lifecycle (NH-02, FE-03, spec 007 FR-04 / AC-04).
 * A command is "pending" from the moment it is sent until a device update reports the
 * requested value, or until it times out. One pending command per device expose: a
 * second request while pending is refused (no silent duplicates), and nothing is retried.
 */
import { DeviceUpdate } from '@/types/device';

export const COMMAND_TIMEOUT_MS = 5000;

export type PendingCommand = { key: string; deviceId: string; expose: string; value: unknown; sentAt: number };

export const commandKey = (deviceId: string, expose: string) => `${deviceId}|${expose}`;

/** Loose equality: buses may echo numbers as strings or vice versa. */
const sameValue = (a: unknown, b: unknown) => a === b || (a !== null && b !== null && a !== undefined && b !== undefined && String(a) === String(b));

/** Keys of pending commands confirmed by this update. */
export function confirmedBy(pending: Iterable<PendingCommand>, update: DeviceUpdate): string[] {
  const keys: string[] = [];
  for (const p of pending) {
    if (p.deviceId !== update.id || !update.data || !(p.expose in update.data)) continue;
    if (sameValue(update.data[p.expose], p.value)) keys.push(p.key);
  }
  return keys;
}

export function isExpired(p: PendingCommand, now: number, timeout = COMMAND_TIMEOUT_MS): boolean {
  return now - p.sentAt >= timeout;
}
