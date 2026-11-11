/**
 * Normalizes an entity identity for catalog lookup.
 *
 * Mirrors ESPHome's `object_id` derivation: lowercase, spaces and punctuation to `_`,
 * with repeated separators collapsed. It intentionally does not split camelCase, so
 * `mmWave Bluetooth` becomes `mmwave_bluetooth`.
 */
export function normalizeEntityKey(value: unknown): string {
  if (typeof value !== 'string') return '';
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '_')
    .replace(/_+/g, '_')
    .replace(/^_+|_+$/g, '');
}
