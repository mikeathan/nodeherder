/*
 * Entry / assumed-presence field definitions and entity option filtering.
 * Mappings are user-selected; no device capability is inferred from names.
 */
import type { Device, Expose } from '@/types/device';
import type { EntryRole } from '@/types/spatial';
import { exposeDisplayLabel } from '@/domain/exposes';

/** One editable entry/assumed-presence mapping row. */
export type EntryFieldDefinition = {
  role: EntryRole;
  label: string;
  type: 'binary' | 'numeric';
  writable: boolean;
};

/** Entry mapping rows in display order. */
export const ENTRY_FIELD_DEFINITIONS: readonly EntryFieldDefinition[] = [
  { role: 'enabled', label: 'Entry / Exit enabled', type: 'binary', writable: true },
  { role: 'timeout', label: 'Assume present timeout', type: 'numeric', writable: true },
  { role: 'threshold', label: 'Exit threshold', type: 'numeric', writable: true },
  { role: 'assumed', label: 'Assumed present', type: 'binary', writable: false },
  { role: 'remaining', label: 'Assumed presence remaining', type: 'numeric', writable: false },
];

/** Candidate entities for one entry field, sorted by display label. */
export function entryEntityOptions(device: Device, field: EntryFieldDefinition): Expose[] {
  return Object.values(device.exposes)
    .filter(
      (expose) =>
        expose.type === field.type && (field.writable ? expose.access_mode !== 'read' : expose.access_mode !== 'write')
    )
    .sort((left, right) => exposeDisplayLabel(left).localeCompare(exposeDisplayLabel(right)));
}
