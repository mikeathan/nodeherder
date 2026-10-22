import { HISTORY_PRESETS, validHistoryRange, type HistoryRange } from './range';
export type HistoryPreference = { version: 1; preset: string; expose: string; custom?: HistoryRange };
type PreferenceStorage = Pick<Storage, 'getItem' | 'setItem'>;
export const historyPreferenceKey = (deviceId: string) => `nh-history-pref:${deviceId}`;
export function readHistoryPreference(storage: PreferenceStorage, deviceId: string, exposes: string[]): HistoryPreference {
  const fallback: HistoryPreference = { version: 1, preset: '24', expose: exposes[0] ?? '' };
  try {
    const parsed: unknown = JSON.parse(storage.getItem(historyPreferenceKey(deviceId)) ?? 'null');
    if (!parsed || typeof parsed !== 'object') return fallback;
    const value = parsed as Record<string, unknown>;
    if (value.version !== 1) return fallback;
    if (typeof value.expose !== 'string' || !exposes.includes(value.expose)) return fallback;
    fallback.expose = value.expose;
    if (!HISTORY_PRESETS.some((p) => p.value === value.preset)) return fallback;
    if (value.preset !== 'custom') return { ...fallback, preset: String(value.preset) };
    if (!value.custom || typeof value.custom !== 'object') return fallback;
    const { from, to } = value.custom as Record<string, unknown>;
    if (typeof from !== 'number' || typeof to !== 'number' || !validHistoryRange({ from, to })) return fallback;
    return { ...fallback, preset: 'custom', custom: { from, to } };
  } catch { return fallback; }
}
export function writeHistoryPreference(storage: PreferenceStorage, deviceId: string, value: HistoryPreference) {
  try { storage.setItem(historyPreferenceKey(deviceId), JSON.stringify(value)); } catch { /* Preferences are optional. */ }
}
