/*
 * Night essentials for Panel mode. A small, validated browser preference: at most four
 * device/expose controls shown after the panel simplifies for the night.
 */
export const PANEL_ESSENTIAL_LIMIT = 4;

export type PanelEssential = {
  deviceId: string;
  expose: string;
  label: string;
};

/** Accepts only the bounded shape; unknown or legacy data falls back to no preference. */
export function parsePanelEssentials(raw: unknown): PanelEssential[] {
  if (!Array.isArray(raw)) return [];
  const seen = new Set<string>();
  const result: PanelEssential[] = [];
  for (const entry of raw) {
    if (typeof entry !== 'object' || entry === null) continue;
    const deviceId = Reflect.get(entry, 'deviceId');
    const expose = Reflect.get(entry, 'expose');
    const label = Reflect.get(entry, 'label');
    if (typeof deviceId !== 'string' || typeof expose !== 'string' || typeof label !== 'string') continue;
    if (!deviceId || !expose || !label) continue;
    const key = `${deviceId}|${expose}`;
    if (seen.has(key)) continue;
    seen.add(key);
    result.push({ deviceId, expose, label });
    if (result.length === PANEL_ESSENTIAL_LIMIT) break;
  }
  return result;
}

/** Adds or removes one control while keeping the stored order and the limit. */
export function togglePanelEssential(current: readonly PanelEssential[], candidate: PanelEssential): PanelEssential[] {
  const key = `${candidate.deviceId}|${candidate.expose}`;
  const exists = current.some((item) => `${item.deviceId}|${item.expose}` === key);
  if (exists) return current.filter((item) => `${item.deviceId}|${item.expose}` !== key);
  if (current.length >= PANEL_ESSENTIAL_LIMIT) return [...current];
  return [...current, candidate];
}
